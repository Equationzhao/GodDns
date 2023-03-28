/*
 *
 *     @file: config.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/28 下午3:58
 *     @last modified: 2023/3/25 下午5:42
 *
 *
 *
 */

// Package DnspodYunApi use Tencent Cloud API to update DNS record
package DnspodYunApi

import (
	"GodDns/DDNS"
	"GodDns/Net"
	"GodDns/Util"
	"gopkg.in/ini.v1"
	"strings"
)

func init() {
	DDNS.Add2FactoryList(factoryInstance)
}

const serviceName = "DnspodYun"

var factoryInstance Factory
var configInstance Config

type Factory struct {
}

func (f Factory) GetName() string {
	return serviceName
}

func (f Factory) Get() DDNS.Config {
	return &configInstance
}

func (f Factory) New() *DDNS.Config {
	var c DDNS.Config = Config{}
	return &c
}

type Config struct {
}

func (c Config) GetName() string {
	return serviceName
}

func (c Config) GenerateDefaultConfigInfo() (DDNS.ConfigStr, error) {
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

func (c Config) ReadConfig(sec ini.Section) ([]DDNS.Parameters, error) {
	names := [9]string{"SecretID", "SecretKey", "Domain", "SubDomain", "RecordId", "RecordLine", "Value", "TTL", "Type"}
	p := DnspodYun{}
	var subdomains []string
	for _, name := range names {
		if !sec.HasKey(name) {
			return nil, DDNS.NewMissKeyErr(name, serviceName)
		} else {
			switch name {
			case "SubDomain":
				subdomain := sec.Key(name).String()
				subdomains = strings.Split(strings.ReplaceAll(subdomain, ",", " "), " ")
				Util.RemoveDuplicate(&subdomains)
			case "TTL":
				ttl, err := sec.Key(name).Uint64()
				if err != nil {
					return nil, err
				}
				p.TTL = ttl
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
	ps := make([]DDNS.Parameters, 0, len(subdomains))
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

func (c Config) GenerateConfigInfo(parameters DDNS.Parameters, u uint) (DDNS.ConfigStr, error) {
	head := DDNS.ConfigHead(parameters, u)

	body := Util.Convert2KeyValue(DDNS.Format, parameters)

	tail := "\n\n"

	content := head + body + tail

	return DDNS.ConfigStr{
		Name:    "Dnspod_yun",
		Content: content,
	}, nil

}
