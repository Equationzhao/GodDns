// Package dnspodyunapi use Tencent Cloud API to update DNS record
package dnspodyunapi

import (
	"GodDns/core"
	"GodDns/netutil"
)

type DnspodYun struct {
	core.DefaultMsgGroup `json:"-" xwwwformurlencoded:"-" KeyValue:"-"`
	SecretID             string
	SecretKey            string
	Domain               string
	SubDomain            string
	RecordId             string
	RecordLine           string
	Value                string
	TTL                  uint64
	Type                 string
	device               string
}

func (s *DnspodYun) GetDevice() string {
	return s.device
}

func (s *DnspodYun) IsDeviceSet() bool {
	return s.device != ""
}

func (s *DnspodYun) SaveConfig(No uint) (core.ConfigStr, error) {
	return configInstance.GenerateConfigInfo(s, No)
}

func (s *DnspodYun) ToRequest() (core.Request, error) {
	r := new(Request)
	r.Init(*s)
	return r, nil
}

func (s *DnspodYun) SetValue(ip string) {
	s.Value = ip
}

func (s *DnspodYun) GetIP() string {
	return s.Value
}

func (s *DnspodYun) GetType() string {
	return netutil.Type2Num(s.Type)
}

func (s *DnspodYun) IsTypeSet() bool {
	return s.Type == "A" || s.Type == "AAAA"
}

func (s *DnspodYun) GetName() string {
	return serviceName
}

func (s *DnspodYun) Target() string {
	return s.getTotalDomain()
}

func (s *DnspodYun) getTotalDomain() string {
	return s.SubDomain + "." + s.Domain
}
