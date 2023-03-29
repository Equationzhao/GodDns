/*
 *
 *     @file: Util_test.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/29 下午11:24
 *     @last modified: 2023/3/29 下午11:20
 *
 *
 *
 */

/*
 *
 *     @file: Util_test.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/28 下午3:58
 *     @last modified: 2023/3/27 下午10:54
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
	"net/url"
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

type TestStruct struct {
	Name     string `json:"name" xwwwformurlencoded:"name"`
	Age      int    `json:"age" xwwwformurlencoded:"age"`
	Nickname string `json:"-" xwwwformurlencoded:"nickname"`
}

type ConvertableXWWWFormUrlencodedMock struct{}

func (c ConvertableXWWWFormUrlencodedMock) Convert2XWWWFormUrlencoded() string {
	return "mock"
}

func TestConvert2XWWWFormUrlencoded(t *testing.T) {
	type args struct {
		i any
	}

	type A struct {
		Device     string `KeyValue:"device" json:"device"`
		IP         string `json:"ip"`
		unexported string
		Type       string
	}

	type B struct {
		x   string
		xx  string
		xxx string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "string",
			args: args{
				i: "test",
			},
			want: "=test",
		},
		{
			name: "map",
			args: args{
				i: map[string]string{
					"name": "test",
					"age":  "18",
				},
			},
			want: "name=test&age=18",
		},
		{
			name: "map with nested struct",
			args: args{
				i: map[string]interface{}{
					"name": "test",
					"age":  "18",
					"profile": TestStruct{
						Name:     "test",
						Age:      18,
						Nickname: "test",
					},
				},
			},

			want: "name=test&age=18&name=test&age=18&nickname=test", // the order of the key-value pairs is not guaranteed
		},
		{
			name: "struct",
			args: args{
				i: TestStruct{
					Name:     "test",
					Age:      18,
					Nickname: "test",
				},
			},
			want: "name=test&age=18&nickname=test",
		},
		{
			name: "slice",
			args: args{
				i: []interface{}{"test", TestStruct{
					Name:     "test",
					Age:      18,
					Nickname: "test",
				}},
			},
			want: "=test&name=test&age=18&nickname=test",
		},
		{
			name: "ConvertableXWWWFormUrlencoded interface",
			args: args{
				i: ConvertableXWWWFormUrlencodedMock{},
			},
			want: "mock",
		},
		{
			name: "nil",
			args: args{
				i: nil,
			},
			want: "",
		},
		{
			name: "empty",
			args: args{
				i: "",
			},
			want: "=",
		},
		{
			name: "empty struct",
			args: args{
				i: struct{}{},
			},
			want: "",
		},
		{
			name: "empty slice",
			args: args{
				i: []interface{}{},
			},
			want: "",
		},
		{
			name: "empty map",
			args: args{
				i: map[string]interface{}{},
			},
			want: "",
		},
		{
			name: "A",
			args: args{
				i: A{Device: "device", IP: "ip", Type: "type", unexported: "123"},
			},
			want: "device=device&ip=ip&Type=type",
		},
		{
			name: "B",
			args: args{
				i: B{x: "x", xx: "xx", xxx: "xxx"},
			},
			want: "",
		},
		{
			name: "A+B",
			args: args{
				i: []any{
					B{x: "x", xx: "xx", xxx: "xxx"},
					A{Device: "device", IP: "ip", Type: "type", unexported: "123"},
				},
			},
			want: "device=device&ip=ip&Type=type",
		},
		{
			name: "p",
			args: args{
				i: &p,
			},
			want: "login_token=550W_MOSS&format=json&lang=en&error_on_empty=no&domain=example.com&record_id=2&sub_domain=s1&record_line=%E9%BB%98%E8%AE%A4&value=fe80%3A%3Aad29%3A79b2%3Af25b%3Aaec4%2536&ttl=600&type=AAAA",
		},
		{
			name: "map",
			args: args{
				i: []any{
					A{
						Device:     "d",
						IP:         "i",
						unexported: "123",
						Type:       "aaa",
					},
					"1233",
					[]string{
						"123", "123", "123",
					},
					map[string]any{
						"2": A{
							Device:     "321",
							IP:         "4325",
							unexported: "432",
							Type:       "trew",
						},
						"name": "321",
					},
				},
			},
			want: "device=d&ip=i&Type=aaa&=1233&=123&=123&=123&device=321&ip=4325&Type=trew&name=321", // the order of the key-value pairs is not guaranteed
		},
		{
			name: "map example",
			args: args{
				i: map[string]string{"device": "device", "ip": "ip", "Type": "type"},
			},
			want: "device=device&ip=ip&Type=type",
		},
		{
			name: "slice example",
			args: args{
				i: []string{"device", "ip", "type"},
			},
			want: "=device&=ip&=type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Util.Convert2XWWWFormUrlencoded(tt.args.i); got != tt.want {
				t.Errorf("Convert2XWWWFormUrlencoded(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
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

func TestTemp(t *testing.T) {
	type A struct {
		Device     string `xwwwformurlencoded:"device" json:"device"`
		IP         string `json:"ip"`
		Type       string
		unexported string
	}

	type B struct {
		X string
		x string
	}
	a := A{Device: "device", IP: "ip", Type: "type"}
	ab := struct {
		A
		B
	}{A: a, B: B{X: "123", x: "321"}}

	res := Util.Convert2XWWWFormUrlencoded(ab)
	t.Log(res)
}

func BenchmarkConvert2XWWWFORMURLENCODED(b *testing.B) {
	type A struct {
		Device     string `xwwwformurlencoded:"device" json:"device"`
		IP         string `json:"ip"`
		Type       string
		unexported string
	}

	type B struct {
		X string
		x string
	}
	a := A{Device: "device", IP: "ip", Type: "type"}
	ab := struct {
		A
		B
	}{A: a, B: B{X: "123", x: "321"}}

	for i := 0; i < b.N; i++ {
		s := Util.Convert2XWWWFormUrlencoded(ab)
		_ = s
	}
}

func BenchmarkURLEncode(b *testing.B) {
	type A struct {
		Device     string `xwwwformurlencoded:"device" json:"device"`
		IP         string `json:"ip"`
		Type       string
		unexported string
	}

	type B struct {
		X string
		x string
	}
	a := A{Device: "device", IP: "ip", Type: "type"}
	ab := struct {
		A
		B
	}{A: a, B: B{X: "123", x: "321"}}
	for i := 0; i < b.N; i++ {
		v := url.Values{}
		v.Add("device", ab.Device)
		v.Add("ip", ab.IP)
		v.Add("type", ab.Type)
		v.Add("X", ab.X)
		s := v.Encode()
		_ = s
	}
}
