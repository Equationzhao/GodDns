/*
 *     @Copyright
 *     @file: Config.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/17 下午8:04
 *     @last modified: 2023/3/17 下午8:02
 *
 *
 *
 */

package Dnspod

import (
	"DDNS/DDNS"
	"DDNS/Net"
	"DDNS/Util"
	"gopkg.in/ini.v1"
	"strconv"
)

type Config struct {
}

var ConfigFactoryInstance ConfigFactory

func init() {
	DDNS.ConfigFactoryList = append(DDNS.ConfigFactoryList, ConfigFactoryInstance)
}

func (c Config) GetName() string {
	return serviceName
}

// GenerateDefaultConfigInfo
// Create default config
// Parameters:none
// Return: Parameters , error
// if any error occurs, FileName will be ""
func (c Config) GenerateDefaultConfigInfo() (DDNS.ConfigStr, error) {
	P := new(Parameters)
	*P = GenerateDefaultConfigInfo()
	return c.GenerateConfigInfo(P, 0)
}

// ReadConfig
// Read config file
// Parameters:none
// Return: Parameters
// if any error occurs, returned Parameters will be nil
func (c Config) ReadConfig(sec ini.Section) (DDNS.Parameters, error) {
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
	//  todo sec.Key("login_token").Validate(func(key string) error {

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

	subdomain := Unpack(sec, "sub_domain", &err)
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

	d := new(Parameters)
	*d = Parameters{
		PublicParameter: PublicParameter{
			LoginToken:   loginToken,
			Format:       format,
			Lang:         lang,
			ErrorOnEmpty: errorOnEmpty,
		},
		ExternalParameter: ExternalParameter{
			Domain:     domain,
			RecordId:   uint32(recordId),
			Subdomain:  subdomain,
			RecordLine: recordLine,
			Value:      value,
			TTL:        uint16(ttl),
			Type:       Type,
		},
		device: device,
	}

	return d, nil
}

type ConfigFactory struct {
}

func (c ConfigFactory) GetName() string {
	return serviceName
}

func (c ConfigFactory) Get() DDNS.Config {
	return &Config{}
}

func (c ConfigFactory) New() *DDNS.Config {
	var config DDNS.Config = &Config{}
	return &config
}

// GenerateConfigInfo
// use reflect to get information of PublicParameter and ExternalParameter
func (c Config) GenerateConfigInfo(parameters DDNS.Parameters, No uint) (DDNS.ConfigStr, error) {
	head := DDNS.ConfigHead(parameters, No)

	body := Util.Convert2KeyValue(DDNS.Format, parameters)

	tail := "\n\n"

	content := head + body + tail

	return DDNS.ConfigStr{
		Name:    "Dnspod",
		Content: content,
	}, nil
}
