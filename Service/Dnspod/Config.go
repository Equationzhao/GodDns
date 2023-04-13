package Dnspod

import (
	"bytes"
	"strings"

	"GodDns/Net"
	"GodDns/Util"
	"GodDns/Util/Collections"
	"GodDns/core"
	"gopkg.in/ini.v1"
)

type Config struct{}

func init() {
	core.Add2FactoryList(configFactoryInstance)
}

// GetName Get name of service
func (c Config) GetName() string {
	return serviceName
}

// GenerateDefaultConfigInfo Create default config
// Return: DDNS.ConfigStr , error
// if any error occurs, FileName will be ""
func (c Config) GenerateDefaultConfigInfo() (core.ConfigStr, error) {
	P := GenerateDefaultConfigInfo()
	return c.GenerateConfigInfo(&P, 0)
}

// ReadConfig Read config file
// Parameters: sec ini.Section
// Return: DDNS.Parameters and error
// if any error occurs, returned Parameters will be nil
func (c Config) ReadConfig(sec ini.Section) ([]core.Parameters, error) {
	names := [11]string{
		"LoginToken", "Format", "Lang", "ErrorOnEmpty", "Domain",
		"RecordId", "RecordLine", "Value", "TTL", "Type", "Subdomain",
	}

	p := Parameters{}
	var subdomains []string
	for _, name := range names {
		if !sec.HasKey(name) {
			return nil, core.NewMissKeyErr(name, serviceName)
		} else {
			switch name {
			case "Subdomain":
				subdomain := sec.Key(name).String()
				subdomains = strings.Fields(strings.ReplaceAll(subdomain, ",", " "))
				Collections.RemoveDuplicate(&subdomains)
			case "TTL":
				ttl, err := sec.Key(name).Uint64()
				if err != nil {
					return nil, err
				}
				p.TTL = uint16(ttl)
			case "Type":
				p.Type = Net.Type2Str(sec.Key(name).String())
			default:
				err := Util.SetVariable(&p, name, sec.Key(name).String())
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// empty if not exist
	p.Device = sec.Key("Device").String()

	ps := make([]core.Parameters, 0, len(subdomains))
	for _, subdomain := range subdomains {
		if subdomain == "" {
			continue
		}
		ps = append(ps, &Parameters{
			LoginToken:   p.LoginToken,
			Format:       p.Format,
			Lang:         p.Lang,
			ErrorOnEmpty: p.ErrorOnEmpty,
			Domain:       p.Domain,
			RecordId:     p.RecordId,
			Subdomain:    subdomain,
			RecordLine:   p.RecordLine,
			Value:        p.Value,
			TTL:          p.TTL,
			Type:         p.Type,
			Device:       p.Device,
		})
	}
	return ps, nil
}

// configFactoryInstance a Factory instance to make dnspod config
var configFactoryInstance ConfigFactory

var configInstance Config

// ConfigFactory is a factory that create a new Config
type ConfigFactory struct{}

// GetName return the name of dnspod
func (c ConfigFactory) GetName() string {
	return serviceName
}

// Get return a singleton Config
func (c ConfigFactory) Get() core.Config {
	return &configInstance
}

// New return a new Config
func (c ConfigFactory) New() *core.Config {
	var config core.Config = &Config{}
	return &config
}

// GenerateConfigInfo
// Generate KeyValue style config
func (c Config) GenerateConfigInfo(parameters core.Parameters, No uint) (core.ConfigStr, error) {
	buffer := bytes.NewBufferString(core.ConfigHead(parameters, No))
	buffer.WriteString(Util.Convert2KeyValue(core.Format, parameters))
	buffer.Write([]byte{'\n', '\n'})

	return core.ConfigStr{
		Name:    "Dnspod",
		Content: buffer.String(),
	}, nil
}
