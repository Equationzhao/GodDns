/*
 *     @Copyright
 *     @file: Parameters.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/22 上午6:29
 *     @last modified: 2023/3/22 上午6:21
 *
 *
 *
 */

package Dnspod

import "C"
import (
	"GodDns/DDNS"
	"GodDns/Net"
)

const serviceName = "Dnspod"

// Parameters implements DeviceOverridable
// PublicParameter is public parameter of dnspod
// ExternalParameter is external parameter of dnspod ddns
// device is device name when overriding ip with specific device/interface
type Parameters struct {
	LoginToken   string `json:"login_token,omitempty" xwwwformurlencoded:"login_token" KeyValue:"login_token,get from https://console.dnspod.cn/account/token/token, 'ID,Token'"`
	Format       string `json:"format,omitempty" xwwwformurlencoded:"format" KeyValue:"format,data format, json(recommended) or xml(not support yet)"`
	Lang         string `json:"lang,omitempty" xwwwformurlencoded:"lang" KeyValue:"lang,language, en or zh(recommended)"`
	ErrorOnEmpty string `json:"error_on_empty,omitempty" xwwwformurlencoded:"error_on_empty" KeyValue:"error_on_empty,return error if the data doesn't exist,no(recommended) or yes"`
	Domain       string `json:"domain,omitempty" xwwwformurlencoded:"domain" KeyValue:"domain,domain name"`
	RecordId     uint32 `json:"record_id,omitempty" xwwwformurlencoded:"record_id" KeyValue:"record_id,record id can be get by making http POST request with required Parameters to https://dnsapi.cn/Record.List, more at https://docs.dnspod.com/api/get-record-list/"`
	Subdomain    string `json:"sub_domain,omitempty" xwwwformurlencoded:"sub_domain" KeyValue:"sub_domain,record name like www., if you have multiple records to update, set like sub_domain=www,ftp,mail"`
	RecordLine   string `json:"record_line,omitempty" xwwwformurlencoded:"record_line" KeyValue:"record_line,The record line.You can get the list from the API.The default value is '默认'"`
	Value        string `json:"value,omitempty" xwwwformurlencoded:"value" KeyValue:"value,IP address like 6.6.6.6"`
	TTL          uint16 `json:"ttl,omitempty" xwwwformurlencoded:"ttl" KeyValue:"ttl,Time-To-Live, 600(default)"`
	Type         string `json:"type,omitempty" xwwwformurlencoded:"type" KeyValue:"type,A/AAAA/4/6"`
	device       string
}

// IsDeviceSet return whether the device is set
func (p *Parameters) IsDeviceSet() bool {
	return p.device != ""
}

// IsTypeSet return whether the type is set correctly
func (p *Parameters) IsTypeSet() bool {
	return p.Type == "AAAA" || p.Type == "A"
}

// SetValue set ip
func (p *Parameters) SetValue(value string) {
	p.Value = value
}

// GetDevice return device name
func (p *Parameters) GetDevice() string {
	return p.device
}

// GetType return Type like "4" or "6" and "" if invalid type
func (p *Parameters) GetType() string {
	return Net.Type2Num(p.Type)
}

// SaveConfig return DDNS.ConfigStr
func (p *Parameters) SaveConfig(No uint) (DDNS.ConfigStr, error) {
	return Config{}.GenerateConfigInfo(p, No)
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
		RecordId:     0,
		Subdomain:    "www,mail,ftp...",
		RecordLine:   "默认",
		Value:        "1.2.3.4",
		TTL:          600,
		Type:         "A/AAAA/4/6",
	}
}

type PublicParameter struct {
}

type ExternalParameter struct {
}

// // MarshalJSON rewrite Parameters marshal function
// func (p *Parameters) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(struct {
// 		LoginToken   string
// 		Format       string
// 		Lang         string
// 		ErrorOnEmpty string
// 		UserId       uint32
// 		Domain       string
// 		RecordId     uint32
// 		Subdomain    string
// 		RecordLine   string
// 		Value        string
// 		TTL          uint16
// 		Type         string
// 	}{
// 		LoginToken:   p.LoginToken,
// 		Format:       p.Format,
// 		Lang:         p.Lang,
// 		ErrorOnEmpty: p.ErrorOnEmpty,
// 		Domain:       p.Domain,
// 		RecordId:     p.RecordId,
// 		Subdomain:    p.Subdomain,
// 		RecordLine:   p.RecordLine,
// 		Value:        p.Value,
// 		TTL:          p.TTL,
// 		Type:         p.Type,
// 	})
// }

// // Convert2XWWWFormUrlencoded rewrite Parameters Convert2XWWWFormUrlencoded function
// func (p *Parameters) Convert2XWWWFormUrlencoded() string {
// 	return Util.Convert2XWWWFormUrlencoded(p.PublicParameter) + "&" + Util.Convert2XWWWFormUrlencoded(p.ExternalParameter)
// }

// Convert2KeyValue rewrite Parameters Convert2KeyValue function
// func (p *Parameters) Convert2KeyValue(format string) string {
// 	return Util.Convert2KeyValue(format, p.PublicParameter) + Util.Convert2KeyValue(format, p.ExternalParameter)
// }

// ToRequest Convert to DDNS.Request
func (p *Parameters) ToRequest() (DDNS.Request, error) {
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
