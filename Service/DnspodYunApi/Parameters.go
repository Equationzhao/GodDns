/*
 *
 *     @file: Parameters.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/30 下午11:29
 *     @last modified: 2023/3/30 下午3:37
 *
 *
 *
 */

// Package DnspodYunApi use Tencent Cloud API to update DNS record
package DnspodYunApi

import (
	"GodDns/DDNS"
	"GodDns/Net"
)

type DnspodYun struct {
	SecretID   string
	SecretKey  string
	Domain     string
	SubDomain  string
	RecordId   string
	RecordLine string
	Value      string
	TTL        uint64
	Type       string
	device     string
}

func (s *DnspodYun) GetDevice() string {
	return s.device
}

func (s *DnspodYun) IsDeviceSet() bool {
	return s.device != ""
}

func (s *DnspodYun) SaveConfig(No uint) (DDNS.ConfigStr, error) {
	return configInstance.GenerateConfigInfo(s, No)
}

func (s *DnspodYun) ToRequest() (DDNS.Request, error) {
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
	return Net.Type2Num(s.Type)
}

func (s *DnspodYun) IsTypeSet() bool {
	return s.Type == "A" || s.Type == "AAAA"
}

func (s *DnspodYun) GetName() string {
	return serviceName
}

func (s *DnspodYun) getTotalDomain() string {
	return s.SubDomain + "." + s.Domain
}
