/*
 *
 *     @file: Config.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/29 下午11:24
 *     @last modified: 2023/3/29 下午11:11
 *
 *
 *
 */


package Dnspod

import (
	"GodDns/DDNS"
	"GodDns/Net"
	"GodDns/Util"
	"bytes"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
)

type Config struct {
	test bool
}

func init() {
	DDNS.Add2FactoryList(configFactoryInstance)
}

// GetName Get name of service
func (c Config) GetName() string {
	return serviceName
}

// GenerateDefaultConfigInfo Create default config
// Return: DDNS.ConfigStr , error
// if any error occurs, FileName will be ""
func (c Config) GenerateDefaultConfigInfo() (DDNS.ConfigStr, error) {
	P := GenerateDefaultConfigInfo()
	return c.GenerateConfigInfo(&P, 0)
}

// ReadConfig Read config file
// Parameters: sec ini.Section
// Return: DDNS.Parameters and error
// if any error occurs, returned Parameters will be nil
func (c Config) ReadConfig(sec ini.Section) ([]DDNS.Parameters, error) {
	var err error = nil

	// if no error, err=nil
	// if error occurs, err=error
	Unpack := func(sec ini.Section, key string, err *error) string {
		temp, err_ := sec.GetKey(key)
		*err = err_
		if err_ != nil {
			return ""
		}
		return temp.Value()
	}

	// sec.HasKey
	// todo sec.Key("login_token").Validate(func(key string) error {
	// use MissingKeyErr

	loginToken := Unpack(sec, "login_token", &err)
	if err != nil {
		return nil, err
	}

	format := Unpack(sec, "format", &err)
	if err != nil {
		return nil, err
	}

	lang := Unpack(sec, "lang", &err)
	if err != nil {
		return nil, err
	}

	errorOnEmpty := Unpack(sec, "error_on_empty", &err)
	if err != nil {
		return nil, err
	}

	domain := Unpack(sec, "domain", &err)
	if err != nil {
		return nil, err
	}

	recordIdTemp := Unpack(sec, "record_id", &err)
	if err != nil {
		return nil, err
	}
	recordId, err := strconv.ParseUint(recordIdTemp, 10, 32)
	if err != nil {
		return nil, err
	}

	recordLine := Unpack(sec, "record_line", &err)
	if err != nil {
		return nil, err
	}

	value := Unpack(sec, "value", &err)
	if err != nil {
		return nil, err
	}

	ttlTemp := Unpack(sec, "ttl", &err)
	if err != nil {
		return nil, err
	}
	ttl, err := strconv.ParseUint(ttlTemp, 10, 16)
	if err != nil {
		return nil, err
	}

	var device = getDefaultDevice()
	if sec.HasKey("device") {
		device = sec.Key("device").String()
	}

	var Type = getDefaultType()
	if sec.HasKey("type") {
		Type = Net.Type2Str(sec.Key("type").String())
	}

	// todo multi-subdomain eg. sub1,sub2
	// 1. remove ','
	// 2. split by ' '
	// 3. remove empty string
	// 4. remove duplicate subdomain
	// 5. copy other parameters and modify subdomains
	subdomain := Unpack(sec, "sub_domain", &err)
	if err != nil {
		return nil, err
	}

	subdomains := strings.Fields(strings.ReplaceAll(subdomain, ",", " "))
	Util.RemoveDuplicate(&subdomains)

	ps := make([]DDNS.Parameters, 0, len(subdomains))

	for _, s := range subdomains {
		if s == "" {
			continue
		}
		d := &Parameters{
			LoginToken:   loginToken,
			Format:       format,
			Lang:         lang,
			ErrorOnEmpty: errorOnEmpty,
			Domain:       domain,
			RecordId:     uint32(recordId),
			Subdomain:    s,
			RecordLine:   recordLine,
			Value:        value,
			TTL:          uint16(ttl),
			Type:         Type,
			device:       device,
		}
		ps = append(ps, d)
	}

	return ps, nil
}

// configFactoryInstance a Factory instance to make dnspod config
var configFactoryInstance ConfigFactory

var configInstance Config

// ConfigFactory is a factory that create a new Config
type ConfigFactory struct {
}

// GetName return the name of dnspod
func (c ConfigFactory) GetName() string {
	return serviceName
}

// Get return a singleton Config
func (c ConfigFactory) Get() DDNS.Config {
	return &configInstance
}

// New return a new Config
func (c ConfigFactory) New() *DDNS.Config {
	var config DDNS.Config = &Config{}
	return &config
}

// GenerateConfigInfo
// Generate KeyValue style config
func (c Config) GenerateConfigInfo(parameters DDNS.Parameters, No uint) (DDNS.ConfigStr, error) {

	buffer := bytes.NewBufferString(DDNS.ConfigHead(parameters, No))
	buffer.WriteString(Util.Convert2KeyValue(DDNS.Format, parameters))
	buffer.Write([]byte{'\n', '\n'})

	return DDNS.ConfigStr{
		Name:    "Dnspod",
		Content: buffer.String(),
	}, nil
}
