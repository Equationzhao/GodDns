package DDNS

import (
	"GodDns/Net"
	"encoding/json"
	"net/url"
	"regexp"
	"testing"
)

func TestGetDefaultProgramConfigurationLocation(t *testing.T) {
	l := getDefaultProgramConfigurationLocation()
	t.Log(l())
}

func TestJsonHandler(t *testing.T) {
	s := struct {
		Code int `json:"code"`
		Data struct {
			IpInfo []struct {
				Value  string `json:"value"`
				Region string `json:"region"`
			} `json:"ipInfo"`
		} `json:"data"`
	}{
		Code: 0,
		Data: struct {
			IpInfo []struct {
				Value  string `json:"value"`
				Region string `json:"region"`
			} `json:"ipInfo"`
		}{},
	}

	s.Code = 123
	s.Data.IpInfo = append(s.Data.IpInfo, struct {
		Value  string `json:"value"`
		Region string `json:"region"`
	}{
		Value:  "1.2.3.4",
		Region: "CN",
	})

	bytes, err := json.Marshal(s)
	if err != nil {
		t.FailNow()
	}

	ip, err := jsonHandler{}.HandleResponse(string(bytes), "data.ipInfo[0].value")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if ip.(string) != "1.2.3.4" {
		t.Error("json handler failed")
	}

	region, err := jsonHandler{}.HandleResponse(string(bytes), "data.ipInfo[0].region")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if region.(string) != "CN" {
		t.Error("json handler failed")
	}

	code, err := jsonHandler{}.HandleResponse(string(bytes), "code")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if code.(float64) != 123 {
		t.Error("json handler failed")
	}

}

func TestURLParse(t *testing.T) {
	urls := "https://myip.ipip.net/s"
	re := regexp.MustCompile("(http|https)://[\\w\\-_]+(\\.[\\w\\-_]+)+([\\w\\-.,@?^=%&:/~+#]*[\\w\\-@?^=%&/~+#])?")
	if !re.MatchString(urls) {
		t.Error("url parse failed")
	}
}

func TestConfigStr(t *testing.T) {
	u1, _ := url.Parse("https://myip.ipip.net/s")
	u2, _ := url.Parse("https://speed.neu6.edu.cn/getIP.php")
	u3, _ := url.Parse("https://ip.3322.net")

	p := ProgramConfig{
		proxy: []url.URL{
			*u1,
			*u2,
			*u3,
		},
		ags: []ApiGenerator{
			{
				apiName:  "MyApi1",
				method:   "GET",
				a:        "https://speed.neu6.edu.cn/getIP.php",
				aaaa:     "https://myip.ipip.net/s",
				response: "TEXT",
				resName:  "0",
			},
			{
				apiName:  "MyApi2",
				method:   "POST",
				a:        "https://ip.3322.net",
				aaaa:     "https://speed.neu6.edu.cn/getIP.php",
				response: "JSON",
				resName:  "ip",
			},
		},
	}

	t.Log(p.ConfigStr().Content)

	t.Log(DefaultConfig.ConfigStr().Content)
}

func TestLoadProxy(t *testing.T) {

	ps, err := loadProxy("[https://ip.3322.net https://speed.neu6.edu.cn/getIP.php https://myip.ipip.net/s ]")
	if err != nil {
		t.Error(err)
	}

	for _, p := range ps {
		t.Log(p)
	}
}

func TestProgramConfigGenerateConfiguration(t *testing.T) {
	u1, _ := url.Parse("socks5://localhost:10808")
	u2, _ := url.Parse("https://localhost:10809")

	p := ProgramConfig{
		proxy: []url.URL{
			*u1,
			*u2,
		},
		ags: []ApiGenerator{
			{
				apiName:  "MyApi1",
				method:   "GET",
				a:        "https://speed.neu6.edu.cn/getIP.php",
				aaaa:     "https://myip.ipip.net/s",
				response: "TEXT",
				resName:  "0",
			},
			{
				apiName:  "MyApi2",
				method:   "GET",
				a:        "https://ip.3322.net",
				aaaa:     "https://speed.neu6.edu.cn/getIP.php",
				response: "JSON",
				resName:  "ip",
			},
		},
	}

	err := p.GenerateConfigFile()
	if err != nil {
		t.Error(err)
	}
}

func TestMyApiGet(t *testing.T) {
	location, err := GetProgramConfigLocation()
	if err != nil {
		t.Error(err)
	}

	config, fatal, other := LoadProgramConfig(location)
	if fatal != nil {
		t.Fatal(fatal)
	}
	if other != nil {
		t.Error(other)
	}

	config.Setup()

	names := [2]string{"MyApi1", "MyApi2"}
	for _, name := range names {
		api, err := Net.ApiMap.GetApi(name)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		ip4, err := api.Get(4)
		if err != nil {
			t.Error(err)
		}
		t.Log(ip4)

		ip6, err := api.Get(6)
		if err != nil {
			t.Error(err)
		}
		t.Log(ip6)
	}

}
