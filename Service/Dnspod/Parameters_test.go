package Dnspod

import (
	"fmt"
	"testing"
)

var p Parameters

func init() {
	p = Parameters{
		LoginToken:   "550W_MOSS",
		Format:       "json",
		Lang:         "en",
		ErrorOnEmpty: "no",
		Domain:       "example.com",
		RecordId:     2,
		Subdomain:    "s1",
		RecordLine:   "默认",
		Value:        "fe80::1",
		TTL:          600,
		Type:         "AAAA",
	}
}

func TestGenerateConfigInfo(t *testing.T) {

	info, err := p.SaveConfig(0)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(info.Content)
}
