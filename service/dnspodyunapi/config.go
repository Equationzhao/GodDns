// Package DnspodYunApi use Tencent Cloud API to update DNS record
package dnspodyunapi

import (
	"bytes"
	"strings"

	"GodDns/core"
	"GodDns/netutil"
	"GodDns/util"
	"GodDns/util/collections"
	"gopkg.in/ini.v1"
)

func init() {
	core.Add2FactoryList(factoryInstance)
}

const serviceName = "DnspodYun"

var (
	factoryInstance Factory
	configInstance  Config
)

type Factory struct{}

func (f Factory) GetName() string {
	return serviceName
}

func (f Factory) Get() core.Config {
	return &configInstance
}

func (f Factory) New() *core.Config {
	var c core.Config = Config{}
	return &c
}

type Config struct{}

func (c Config) GetName() string {
	return serviceName
}

func (c Config) GenerateDefaultConfigInfo() (core.ConfigStr, error) {
	p := GenerateDefaultConfigInfo()
	return c.GenerateConfigInfo(&p, 0)
}

func GenerateDefaultConfigInfo() DnspodYun {
	return DnspodYun{
		SecretID:   "SecretID",
		SecretKey:  "SecretKey",
		Domain:     "example.com",
		SubDomain:  "www",
		RecordId:   "0",
		RecordLine: "默认",
		Value:      "1.2.3.4",
		TTL:        600,
		Type:       "A/AAAA/4/6",
	}
}

func (c Config) ReadConfig(sec ini.Section) ([]core.Parameters, error) {
	names := [9]string{"SecretID", "SecretKey", "Domain", "SubDomain", "RecordId", "RecordLine", "Value", "TTL", "Type"}
	p := DnspodYun{}
	var subdomains []string
	for _, name := range names {
		if !sec.HasKey(name) {
			return nil, core.NewMissKeyErr(name, serviceName)
		} else {
			switch name {
			case "SubDomain":
				subdomain := sec.Key(name).String()
				subdomains = strings.Fields(strings.ReplaceAll(subdomain, ",", " "))
				collections.RemoveDuplicate(&subdomains)
			case "TTL":
				ttl, err := sec.Key(name).Uint64()
				if err != nil {
					return nil, err
				}
				p.TTL = ttl
			case "Type":
				p.Type = netutil.Type2Str(sec.Key(name).String())
			default:
				err := util.SetVariable(&p, name, sec.Key(name).String())
				if err != nil {
					return nil, err
				}
			}
		}
	}
	ps := make([]core.Parameters, 0, len(subdomains))
	for _, subdomain := range subdomains {
		if subdomain == "" {
			continue
		}
		ps = append(ps, &DnspodYun{
			SecretID:   p.SecretID,
			SecretKey:  p.SecretKey,
			Domain:     p.Domain,
			SubDomain:  subdomain,
			RecordId:   p.RecordId,
			RecordLine: p.RecordLine,
			Value:      p.Value,
			TTL:        p.TTL,
			Type:       p.Type,
		})
	}

	return ps, nil
}

func (c Config) GenerateConfigInfo(parameters core.Parameters, u uint) (core.ConfigStr, error) {
	buffer := bytes.NewBufferString(core.ConfigHead(parameters, u))
	buffer.WriteString(util.Convert2KeyValue(core.Format, parameters))
	buffer.Write([]byte{'\n', '\n'})

	return core.ConfigStr{
		Name:    "Dnspod_yun",
		Content: buffer.String(),
	}, nil
}
