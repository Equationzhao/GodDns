package Dnspod

import "C"
import (
	"GodDns/Net"
	"GodDns/core"
)

const serviceName = "Dnspod"

// Parameters implements DeviceOverridable
// PublicParameter is public parameter of dnspod
// ExternalParameter is external parameter of dnspod ddns
// Device is Device name when overriding ip with specific Device/interface
type Parameters struct {
	LoginToken   string `json:"login_token,omitempty" xwwwformurlencoded:"login_token" KeyValue:"LoginToken,get from https://console.dnspod.cn/account/token/token, 'ID,Token'"`
	Format       string `json:"format,omitempty" xwwwformurlencoded:"format" KeyValue:"Format,data format, json(recommended) or xml(not support yet)"`
	Lang         string `json:"lang,omitempty" xwwwformurlencoded:"lang" KeyValue:"Lang,language, en or zh(recommended)"`
	ErrorOnEmpty string `json:"error_on_empty,omitempty" xwwwformurlencoded:"ErrorOnEmpty" KeyValue:"ErrorOnEmpty,return error if the data doesn't exist,no(recommended) or yes"`
	Domain       string `json:"domain,omitempty" xwwwformurlencoded:"domain" KeyValue:"domain,domain name"`
	RecordId     string `json:"record_id,omitempty" xwwwformurlencoded:"record_id" KeyValue:"RecordId,record id can be get by making http POST request with required Parameters to https://dnsapi.cn/Record.List, more at https://docs.dnspod.com/api/get-record-list/"`
	Subdomain    string `json:"sub_domain,omitempty" xwwwformurlencoded:"sub_domain" KeyValue:"Subdomain,record name like www., if you have multiple records to update, set like sub_domain=www,ftp,mail"`
	RecordLine   string `json:"record_line,omitempty" xwwwformurlencoded:"record_line" KeyValue:"RecordLine,The record line.You can get the list from the API.The default value is '默认'"`
	Value        string `json:"value,omitempty" xwwwformurlencoded:"value" KeyValue:"Value,IP address like 6.6.6.6"`
	TTL          uint16 `json:"ttl,omitempty" xwwwformurlencoded:"ttl" KeyValue:"TTL,Time-To-Live, 600(default)"`
	Type         string `json:"type,omitempty" xwwwformurlencoded:"type" KeyValue:"Type,A/AAAA/4/6"`
	Device       string `json:"-" xwwwformurlencoded:"-" KeyValue:"Device,device/net interface name"`
}

func (p *Parameters) Target() string {
	return p.getTotalDomain()
}

// IsDeviceSet return whether the Device is set
func (p *Parameters) IsDeviceSet() bool {
	return p.Device != ""
}

// IsTypeSet return whether the type is set correctly
func (p *Parameters) IsTypeSet() bool {
	return p.Type == "AAAA" || p.Type == "A"
}

// SetValue set ip
func (p *Parameters) SetValue(value string) {
	p.Value = value
}

// GetDevice return Device name
func (p *Parameters) GetDevice() string {
	return p.Device
}

// GetType return Type like "4" or "6" and "" if invalid type
func (p *Parameters) GetType() string {
	return Net.Type2Num(p.Type)
}

// SaveConfig return DDNS.ConfigStr
func (p *Parameters) SaveConfig(No uint) (core.ConfigStr, error) {
	return configInstance.GenerateConfigInfo(p, No)
}

// GetName return "Dnspod"
func (p *Parameters) GetName() string {
	return serviceName
}

// GetIP return ip value
func (p *Parameters) GetIP() string {
	return p.Value
}

// GenerateDefaultConfigInfo  return Default config
func GenerateDefaultConfigInfo() Parameters {
	return Parameters{
		LoginToken:   "Token",
		Format:       "json",
		Lang:         "en",
		ErrorOnEmpty: "no",
		Domain:       "example.com",
		RecordId:     "record id",
		Subdomain:    "www,mail,ftp...",
		RecordLine:   "默认",
		Value:        "1.2.3.4",
		TTL:          600,
		Type:         "A/AAAA/4/6",
		Device:       "your device/net interface name",
	}
}

type PublicParameter struct{}

type ExternalParameter struct{}

// ToRequest Convert to DDNS.Request
func (p *Parameters) ToRequest() (core.Request, error) {
	r := new(Request)
	err := r.Init(*p)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func getDefaultDevice() string {
	return ""
}

func getDefaultType() string {
	return ""
}

// getTotalDomain return subdomain+domain
func (p *Parameters) getTotalDomain() string {
	return p.Subdomain + "." + p.Domain
}
