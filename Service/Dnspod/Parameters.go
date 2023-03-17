/*
 *     @Copyright
 *     @file: Parameters.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/17 下午9:54
 *     @last modified: 2023/3/17 下午8:07
 *
 *
 *
 */

package Dnspod

import "C"
import (
	"DDNS/DDNS"
	"DDNS/Net"
	"DDNS/Util"
	"encoding/json"
)

const serviceName = "Dnspod"

type Parameters struct {
	PublicParameter   PublicParameter
	ExternalParameter ExternalParameter

	device string
}

func (p *Parameters) IsDeviceSet() bool {
	return p.device != ""
}

func (p *Parameters) IsTypeSet() bool {
	return p.ExternalParameter.Type == "AAAA" || p.ExternalParameter.Type == "A"
}

func (p *Parameters) SetValue(value string) {
	p.ExternalParameter.Value = value
}

func (p *Parameters) GetDevice() string {
	return p.device
}

func (p *Parameters) GetType() string {
	return Net.Type2Num(p.ExternalParameter.Type)
}

func (p *Parameters) SaveConfig(No uint) (DDNS.ConfigStr, error) {
	return Config{}.GenerateConfigInfo(p, No)
}

func (p *Parameters) GetName() string {
	return serviceName
}

func (p *Parameters) GetIP() string {
	return p.ExternalParameter.Value
}

func GenerateDefaultConfigInfo() Parameters {
	return Parameters{
		PublicParameter: PublicParameter{
			LoginToken:   "Token",
			Format:       "json",
			Lang:         "en",
			ErrorOnEmpty: "no",
		},

		ExternalParameter: ExternalParameter{
			Domain:     "example.com",
			RecordId:   0,
			Subdomain:  "www",
			RecordLine: "默认",
			Value:      "1.2.3.4",
			TTL:        600,
			Type:       "A/AAAA/4/6",
		},
	}
}

type PublicParameter struct {
	LoginToken   string `json:"login_token,omitempty" xwwwformurlencoded:"login_token" KeyValue:"login_token"`
	Format       string `json:"format,omitempty" xwwwformurlencoded:"format" KeyValue:"format"`
	Lang         string `json:"lang,omitempty" xwwwformurlencoded:"lang" KeyValue:"lang"`
	ErrorOnEmpty string `json:"error_on_empty,omitempty" xwwwformurlencoded:"error_on_empty" KeyValue:"error_on_empty"`
}

type ExternalParameter struct {
	Domain     string `json:"domain,omitempty" xwwwformurlencoded:"domain" KeyValue:"domain"`
	RecordId   uint32 `json:"record_id,omitempty" xwwwformurlencoded:"record_id" KeyValue:"record_id"`
	Subdomain  string `json:"sub_domain,omitempty" xwwwformurlencoded:"sub_domain" KeyValue:"sub_domain"`
	RecordLine string `json:"record_line,omitempty" xwwwformurlencoded:"record_line" KeyValue:"record_line"`
	Value      string `json:"value,omitempty" xwwwformurlencoded:"value" KeyValue:"value"`
	TTL        uint16 `json:"ttl,omitempty" xwwwformurlencoded:"ttl" KeyValue:"ttl"`
	Type       string `json:"type,omitempty" xwwwformurlencoded:"type" KeyValue:"type"`
}

// MarshalJSON
// rewrite Parameters marshal function
func (p *Parameters) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		LoginToken   string
		Format       string
		Lang         string
		ErrorOnEmpty string
		UserId       uint32
		Domain       string
		RecordId     uint32
		Subdomain    string
		RecordLine   string
		Value        string
		TTL          uint16
		Type         string
	}{
		LoginToken:   p.PublicParameter.LoginToken,
		Format:       p.PublicParameter.Format,
		Lang:         p.PublicParameter.Lang,
		ErrorOnEmpty: p.PublicParameter.ErrorOnEmpty,
		Domain:       p.ExternalParameter.Domain,
		RecordId:     p.ExternalParameter.RecordId,
		Subdomain:    p.ExternalParameter.Subdomain,
		RecordLine:   p.ExternalParameter.RecordLine,
		Value:        p.ExternalParameter.Value,
		TTL:          p.ExternalParameter.TTL,
		Type:         p.ExternalParameter.Type,
	})
}

// Convert2XWWWFormUrlencoded
// rewrite Parameters Convert2XWWWFormUrlencoded function
func (p *Parameters) Convert2XWWWFormUrlencoded() string {
	return Util.Convert2XWWWFormUrlencoded(p.PublicParameter) + "&" + Util.Convert2XWWWFormUrlencoded(p.ExternalParameter)
}

// Convert2KeyValue
// rewrite Parameters Convert2KeyValue function
func (p *Parameters) Convert2KeyValue(format string) string {
	return Util.Convert2KeyValue(format, p.PublicParameter) + Util.Convert2KeyValue(format, p.ExternalParameter)
}

func (p *Parameters) ToRequest() (DDNS.Request, error) {
	r := new(Request)
	err := r.Init(p)
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

func (p *Parameters) getTotalDomain() string {
	return p.ExternalParameter.Subdomain + "." + p.ExternalParameter.Domain
}
