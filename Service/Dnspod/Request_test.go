package Dnspod

import (
	"fmt"
	"testing"
	"time"

	"GodDns/Util"
	DDNS "GodDns/core"

	"github.com/sirupsen/logrus"
)

func TestXWWWFORMURLENCODED(t *testing.T) {
	p := Parameters{
		LoginToken:   "TOKEN",
		Format:       "json",
		Lang:         "cn",
		ErrorOnEmpty: "no",
		Domain:       "domain.com",
		RecordId:     "0",
		Subdomain:    "",
		RecordLine:   "默认",
		Value:        "",
		TTL:          600,
		Type:         "A",
	}
	fmt.Println(Util.Convert2XWWWFormUrlencoded(p))
}

func TestRequest_GetRecordId(t *testing.T) {
	p := Parameters{
		LoginToken:   "TOKEN",
		Format:       "json",
		Lang:         "cn",
		ErrorOnEmpty: "no",
		Domain:       "domain.com",
		RecordId:     "0",
		Subdomain:    "",
		RecordLine:   "默认",
		Value:        "",
		TTL:          600,
		Type:         "A",
	}

	r := Request{
		parameters: p,
	}

	done := make(chan bool)
	status := DDNS.Status{}
	var err error
	go func() {
		status, err = r.GetRecordId()
		done <- true
	}()
	if err != nil {
		t.Error(err)
	}

	t.Log(status)
}

func TestRequest_MakeRequest(t *testing.T) {
	logrus.SetLevel(logrus.TraceLevel)

	p := Parameters{
		LoginToken:   "TOKEN",
		Format:       "json",
		Lang:         "en",
		ErrorOnEmpty: "no",
		Domain:       "domain.com",
		RecordId:     "0",
		Subdomain:    "",
		RecordLine:   "默认",
		Value:        "",
		TTL:          600,
		Type:         "AAAA",
	}

	r := Request{
		parameters: p,
	}
	status := DDNS.Status{}
	done := make(chan bool)
	var err error
	go func() {
		status, err = r.GetRecordId()
		done <- true
	}()
	select {
	case <-done:
		if err != nil {
			t.Error(err)
		}
	case <-time.After(10 * time.Second):
		t.Failed()
	}

	t.Log(status)

	err = r.MakeRequest()
	if err != nil {
		t.Error(err)
	}

	t.Log(status)
}
