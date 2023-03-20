/*
 *     @Copyright
 *     @file: Request_test.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/20 下午11:29
 *     @last modified: 2023/3/20 下午11:27
 *
 *
 *
 */

package Dnspod

import (
	"github.com/sirupsen/logrus"
	"testing"
)

func TestRequest_GetRecordId(t *testing.T) {

	logrus.SetLevel(logrus.TraceLevel)

	p := Parameters{
		PublicParameter: PublicParameter{
			LoginToken:   "TOKEN",
			Format:       "json",
			Lang:         "cn",
			ErrorOnEmpty: "no",
		},
		ExternalParameter: ExternalParameter{
			Domain:     "domain.com",
			RecordId:   0,
			Subdomain:  "",
			RecordLine: "默认",
			Value:      "",
			TTL:        600,
			Type:       "A",
		},
	}

	r := Request{
		parameters: p,
	}
	status, err := r.GetRecordId()
	if err != nil {
		t.Error(err)
	}

	t.Log(status)

}

func TestRequest_MakeRequest(t *testing.T) {
	logrus.SetLevel(logrus.TraceLevel)

	p := Parameters{
		PublicParameter: PublicParameter{
			LoginToken:   "TOKEN",
			Format:       "json",
			Lang:         "en",
			ErrorOnEmpty: "no",
		},
		ExternalParameter: ExternalParameter{
			Domain:     "domain.com",
			RecordId:   0,
			Subdomain:  "",
			RecordLine: "默认",
			Value:      "",
			TTL:        600,
			Type:       "AAAA",
		},
	}

	r := Request{
		parameters: p,
	}
	status, err := r.GetRecordId()
	if err != nil {
		t.Error(err)
	}

	t.Log(status)

	err = r.MakeRequest()
	if err != nil {
		t.Error(err)
	}

	t.Log(status)
}
