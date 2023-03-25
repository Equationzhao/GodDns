/*
 *     @Copyright
 *     @file: Util_test.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/25 下午5:41
 *     @last modified: 2023/3/25 下午5:07
 *
 *
 *
 */

package Tests_test

import (
	"GodDns/DDNS"
	"GodDns/Service/Dnspod"
	"GodDns/Util"
	"io"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

var p Dnspod.Parameters

func init() {
	p = Dnspod.Parameters{
		LoginToken:   "550W_MOSS",
		Format:       "json",
		Lang:         "en",
		ErrorOnEmpty: "no",
		Domain:       "example.com",
		RecordId:     2,
		Subdomain:    "s1",
		RecordLine:   "默认",
		Value:        "fe80::ad29:79b2:f25b:aec4%36",
		TTL:          600,
		Type:         "AAAA",
	}

}

func TestConfigFileGenerator(t *testing.T) {
	config := Dnspod.Config{}
	dnspod, err := config.GenerateDefaultConfigInfo()
	if err != nil {
		t.Error(err)
	}

	err = DDNS.ConfigureWriter("test.conf", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, dnspod)
	if err != nil {
		t.Error(err)
	}
}

func TestConvert2KeyValue(t *testing.T) {

	type B struct {
		X string
		x string
	}

	type A struct {
		Device     string `KeyValue:"device,device name" json:"device"`
		IP         string `json:"ip,omitempty,string"`
		Type       string
		unexported string
		B          B
	}

	a := A{Device: "device", IP: "ip", Type: "type", unexported: "123", B: B{X: "123", x: "321"}}

	t.Log("\n", Util.Convert2KeyValue("%s: %s", a))

	t.Log("\n", Util.Convert2KeyValue("%s = %v", &p))

}

func TestConvert2XWWWFormUrlencoded(t *testing.T) {

	type A struct {
		Device     string `KeyValue:"device" json:"device"`
		IP         string `json:"ip"`
		Type       string
		unexported string
	}

	a := A{Device: "device", IP: "ip", Type: "type", unexported: "123"}

	urlencoded := Util.Convert2XWWWFormUrlencoded(a)
	t.Log(urlencoded)

	res := Util.Convert2XWWWFormUrlencoded(&p)
	t.Log(res)

	as := []any{
		a,
		"1233",
		[]string{
			"123", "123", "123",
		},
		map[string]any{
			"name": "321",
			"2": A{
				Device:     "321",
				IP:         "4325",
				Type:       "trew",
				unexported: "432",
			},
		},
	}

	urlencoded = Util.Convert2XWWWFormUrlencoded(as)
	t.Log(urlencoded)

}

func TestConfigureReader(t *testing.T) {
	location, err := DDNS.GetDefaultConfigurationLocation()
	if err != nil {
		t.Error(err)
	}
	ps, err, errs := DDNS.ConfigureReader(location, Dnspod.ConfigFactory{})
	if err != nil {
		t.Error(err)
	}

	if errs != nil {
		t.Error(errs)
	}

	t.Log(ps)
}

func TestGetVariable(t *testing.T) {
	s := struct {
		Name string
		name string
	}{
		Name: "X",
		name: "x",
	}

	v, err := Util.GetVariable(s, "Name")
	if err != nil || v != s.Name {
		t.Error(err)
	}
	t.Logf("v(%s)=s.Name(%s)", v, s.Name)

	// should return an error
	// because the field name is unexported
	v, err = Util.GetVariable(s, "name")
	if err == nil {
		t.FailNow()
	}

}

func TestSetVariable(t *testing.T) {
	s := struct {
		Name string
		name string
	}{
		Name: "X",
		name: "x",
	}
	SCopy := s
	err := Util.SetVariable(&s, "Name", "Y")
	if err != nil || s == SCopy {
		t.Error(err)
	}
	t.Logf("\nBefore: s.Name(%s) \nAfter: s.Name(%s)", SCopy.Name, s.Name)
	err = Util.SetVariable(&s, "name", "y")

	if err == nil {
		t.FailNow()
	}

}

func testSetLog() (func() error, error) {
	file, err := os.OpenFile("test.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	logrus.SetOutput(io.MultiWriter(file, os.Stdout))
	return func() error {
		err := file.Close()
		if err != nil {
			return err
		}
		return nil
	}, nil
}

func testSetLog2() {
	logrus.Infof("test2")
}

func TestLog(t *testing.T) {
	f, err := testSetLog()
	if err != nil {
		logrus.Error(err)
	}

	logrus.Infof("test")
	testSetLog2()

	defer func() {
		err := f()
		if err != nil {
			logrus.Error(err)
		}
	}()

}

func TestGetTypeName(t *testing.T) {
	s := DDNS.Status{
		Name:   "Test",
		Msg:    "Hello",
		Status: DDNS.Success,
	}

	t.Log(Util.GetTypeName(s))
	t.Log(Util.GetTypeName(&s))

	b := make(map[string]int)
	c := make([]string, 10)

	t.Log(Util.GetTypeName(b))
	t.Log(Util.GetTypeName(c))

}
