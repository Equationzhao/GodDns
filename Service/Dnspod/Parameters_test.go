/*
 *     @Copyright
 *     @file: Parameters_test.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/17 下午9:54
 *     @last modified: 2023/3/17 下午8:07
 *
 *
 *
 */

package Dnspod

import (
	"encoding/json"
	"fmt"
	"testing"
)

var p Parameters

func init() {
	p = Parameters{
		PublicParameter: PublicParameter{

			LoginToken:   "550W_MOSS",
			Format:       "json",
			Lang:         "en",
			ErrorOnEmpty: "no",
		},

		ExternalParameter: ExternalParameter{
			Domain:     "example.com",
			RecordId:   2,
			Subdomain:  "s1",
			RecordLine: "默认",
			Value:      "fe80::1",
			TTL:        600,
			Type:       "AAAA",
		},
	}
}

func TestGenerateConfigInfo(t *testing.T) {

	info, err := p.SaveConfig(0)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(info)
}

func TestMarshal(t *testing.T) {

	res, _ := json.Marshal(p)
	fmt.Println(string(res))
}

func TestDnspodParameters_ToRequest(t *testing.T) {
	request, _ := p.ToRequest()

	fmt.Println(request)
}
