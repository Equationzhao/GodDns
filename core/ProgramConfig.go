// Package core is the core package of GodDns, it contains the basic interfaces and tools for DDNS service.
package core

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	log "GodDns/Log"
	"GodDns/Net"
	"GodDns/Util"
	"GodDns/Util/Collections"
	json "GodDns/Util/Json"

	"github.com/go-resty/resty/v2"
	"gopkg.in/ini.v1"
)

const URLPattern = `(http|https)://[\w\-_]+(\.[\w\-_]+)+([\w\-.,@?^=%&:/~+#]*[\w\-@?^=%&/~+#])?`

var GetProgramConfigLocation = getDefaultProgramConfigurationLocation()

const ProgramConfigFileName = "GodDns.ini"

func getDefaultProgramConfigurationLocation() func() (string, error) {
	defaultConfiguration, err := defaultConfigurationDirectory()
	if err != nil {
		return func() (string, error) {
			return "", err
		}
	}

	sep := string(filepath.Separator) // get system separator
	return func() (string, error) {
		return defaultConfiguration + sep + ProgramConfigFileName, err
	}
}

type proxies []url.URL

func (p proxies) Convert2KeyValue(format string) string {
	v := "["
	for _, proxy := range p {
		v += proxy.String() + " "
	}
	v += "]\n\n"
	return fmt.Sprintf(format, "Proxy", v)
}

// ProgramConfig  config for program
type ProgramConfig struct {
	// todo add config for program
	// 1. proxy
	// 2. custom apis
	// 3. custom services Vscode-like ?
	// ...
	proxy proxies
	ags   []ApiGenerator
}

func (p *ProgramConfig) Convert2KeyValue(format string) (content string) {
	buffer := bytes.NewBufferString("[settings]\n")
	buffer.WriteString(p.proxy.Convert2KeyValue(format))
	buffer.WriteByte('\n')
	for _, api := range p.ags {
		buffer.WriteString(api.Convert2KeyValue(format))
	}
	return buffer.String()
}

func (p *ProgramConfig) ConfigStr() ConfigStr {
	return ConfigStr{
		Name:    "ProgramSettings",
		Content: p.Convert2KeyValue(Format),
	}
}

// Setup  program
// 1. set proxy [not implemented]
// 2. add apis
// 3. ...
func (p *ProgramConfig) Setup() {
	// 1. set proxy
	for _, p := range p.proxy {
		Net.AddProxy(Net.GlobalProxies, p.String())
	}
	// 2. add apis
	for _, ag := range p.ags {
		api := ag.Generate()
		Net.ApiMap.Add2Apis(ag.apiName, api)
	}
}

var DefaultConfig = ProgramConfig{
	proxy: nil,
	ags:   nil,
}

// GenerateConfigFile generate config file
// if config file already exists, return error
func (p *ProgramConfig) GenerateConfigFile() error {
	location, err := GetProgramConfigLocation()
	if err != nil {
		return err
	}

	if IsConfigExist(location) {
		return errors.New("config file already exists")
	}

	// if !errors.Is(err, &os.PathError{}) {
	// 	panic("not a path error")
	// }

	return ConfigureWriter(location, os.O_CREATE|os.O_WRONLY, p.ConfigStr())
}

func IsConfigExist(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

// LoadProgramConfig load program config from file
func LoadProgramConfig(file string) (programConfig *ProgramConfig, Fatal error, Warn error) {
	cfg, Fatal := ini.Load(file)
	if Fatal != nil {
		return &ProgramConfig{}, Fatal, nil
	}
	cfg.BlockMode = false

	res := &ProgramConfig{}

	// load from env
	// proxy from env: ALL_PROXY, HTTP_PROXY, HTTPS_PROXY
	envProxy, err := loadProxy(fmt.Sprintf("[%s %s %s]", os.Getenv("ALL_PROXY"), os.Getenv("HTTP_PROXY"), os.Getenv("HTTPS_PROXY")))
	res.proxy = envProxy
	if err != nil {
		Warn = errors.Join(Warn, err)
	}

	// no color from env: NO_COLOR
	NoColor := os.Getenv("NO_COLOR")
	if NoColor != "" {
		// disable coloring
		log.InfoPP.SetColoringEnabled(false)
		log.WarnPP.SetColoringEnabled(false)
		log.DebugPP.SetColoringEnabled(false)
		log.ErrPP.SetColoringEnabled(false)
	}

	// load from file
	for _, section := range cfg.Sections() {
		switch section.Name() {
		case "DEFAULT", "default", "Default":
			continue
		case "Settings", "settings", "SETTINGS":
			for _, k := range section.Keys() {
				switch k.Name() {
				case "Proxy", "proxy", "PROXY":
					// var err error
					proxy, err := loadProxy(k.Value())
					res.proxy = proxy
					if err != nil {
						Warn = errors.Join(Warn, err)
					}
				default:
					Warn = errors.Join(Warn, NewUnknownKeyErr(k.Name(), section.Name()))
				}
			}
		default:
			// load apis
			if strings.HasPrefix(section.Name(), "Api.") ||
				strings.HasPrefix(section.Name(), "api.") ||
				strings.HasPrefix(section.Name(), "API.") {
				if len(section.Name()) == 4 {
					Warn = errors.Join(Warn, fmt.Errorf("invalid api name: `%s`", section.Name()))
					continue
				}
				api, err := LoadApiFromConfig(section)
				if err != nil {
					Warn = errors.Join(Warn, err)
				} else {
					res.ags = append(res.ags, api)
				}
			} else {
				Warn = errors.Join(Warn, fmt.Errorf("unknown section: %s", section.Name()))
			}
		}
	}

	return res, nil, Warn
}

// LoadApiFromConfig load api from config
// load string like "[http://localhost:10809 https://example.com:12345 socks5://127.0.0.1:10808 ]"
func loadProxy(proxy string) (res []url.URL, err error) {
	split := strings.Fields(strings.ReplaceAll(strings.Trim(proxy, "[]"), ",", " "))
	res = make([]url.URL, 0, len(split))
	// remove empty string
	for _, s := range split {
		if s != "" {
			if u, bad := url.Parse(s); bad == nil {
				res = append(res, *u)
			} else {
				err = errors.Join(err, fmt.Errorf("invalid proxy: %s", s))
			}
		} else {
			continue
		}
	}

	// remove duplicate
	Collections.RemoveDuplicate[url.URL](&res)

	return res, err
}

type responseHandler interface {
	HandleResponse(source string, toGet string) (target any, err error)
}

type methodHandler interface {
	Do(string) (string, error)
}

type GetHandler struct{}

func (g GetHandler) Do(s string) (string, error) {
	c := MainClientPool.Get().(*resty.Client)
	defer MainClientPool.Put(c)
	res, err := c.R().Get(s)
	if err != nil {
		return "", err
	}

	if res.String() == "" {
		return "", errors.New("empty response")
	}
	return res.String(), nil
}

type PostHandler struct{}

func (p PostHandler) Do(s string) (string, error) {
	c := MainClientPool.Get().(*resty.Client)
	defer MainClientPool.Put(c)
	res, err := c.R().Post(s)
	if err != nil {
		return "", err
	}
	if res.String() == "" {
		return "", errors.New("empty response")
	}
	return res.String(), nil
}

// ApiGenerator can generate api from config
type ApiGenerator struct {
	apiName         string
	method          string
	methodHandler   methodHandler
	a               string
	aaaa            string
	response        string
	resName         string
	responseHandler responseHandler
}

// Convert2KeyValue convert ApiGenerator to key-value format
func (a *ApiGenerator) Convert2KeyValue(format string) string {
	head := "[api." + a.apiName + "]\n"
	s := struct {
		A          string
		AAAA       string
		Response   string
		HTTPMethod string
		Value      string
	}{
		A:          a.a,
		AAAA:       a.aaaa,
		HTTPMethod: a.method,
		Response:   a.response,
		Value:      a.resName,
	}

	body := Util.Convert2KeyValue(format, s)

	tail := "\n\n"

	content := head + body + tail

	return content
}

// validateResponseType validate response type
// support json and txt for now
// set responseHandler
func (a *ApiGenerator) validateResponseType() error {
	var err error
	switch a.response {
	case "json", "JSON", "Json":
		a.response = "json"
		a.responseHandler = jsonHandler{}
	case "xml", "XML", "Xml":
		a.response = "xml"
	case "text", "TEXT", "Text":
		a.response = "text"
		a.responseHandler = txtHandler{}
	case "html", "HTML", "Html":
		a.response = "html"
	default:
		err = errors.New("unknown response type")
	}

	return err
}

func (a *ApiGenerator) validateURL() error {
	URLPattern := regexp.MustCompile(URLPattern)
	URLS := [2]string{a.a, a.aaaa}
	for _, URL := range URLS {
		if !URLPattern.MatchString(URL) {
			return errors.New("invalid URL")
		}
	}
	return nil
}

func (a *ApiGenerator) validateMethod() error {
	var err error
	switch a.method {
	case "GET", "get", "Get":
		a.method = "GET"
		a.methodHandler = GetHandler{}
	case "POST", "post", "Post":
		a.method = "POST"
		a.methodHandler = PostHandler{}
	default:
		err = errors.New("unsupported method")
	}
	return err
}

// Generate  api,
// PLZ check if the api is valid before call Generate(), use ApiGenerator.Validate()
func (a *ApiGenerator) Generate() Net.Api {
	copiedA := *a
	f := func(t Net.Type) (string, error) {
		URL := ""
		switch t {
		case Net.A:
			URL = copiedA.a
		case Net.AAAA:
			URL = copiedA.aaaa
		default:
			return "", Net.NewUnknownType(t)
		}
		res, err := copiedA.methodHandler.Do(URL)
		if err != nil {
			return "", err
		}
		target, err := copiedA.responseHandler.HandleResponse(res, copiedA.resName)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%v", target), nil
	}
	return Net.Api{
		Get: f,
	}
}

type jsonHandler struct{}

func (j jsonHandler) HandleResponse(source string, toGet string) (target any, err error) {
	// if response:
	// {
	// 	"code": 0,
	//  "data": {
	// 		"ipInfo": [
	// 			{
	// 				"value": "1.2.3.4",
	// 				"region": "CN",
	// 			},
	// 		]
	//  }
	// }
	// resName should be "data.ipInfo[0].value"

	// parse resName
	var result map[string]any

	err = json.UnmarshalString(source, &result)
	if err != nil {
		return "", err
	}

	parts := strings.Split(toGet, ".")
	for _, part := range parts {
		if strings.Contains(part, "[") {
			key := part[:strings.Index(part, "[")]
			index := part[strings.Index(part, "[")+1 : strings.Index(part, "]")]
			i, err := strconv.Atoi(index)
			if err != nil {
				return "", err
			}
			result = result[key].([]any)[i].(map[string]any)
		} else {
			resultTemp, ok := result[part].(map[string]any)
			if !ok {
				targetTemp := result[part]
				if targetTemp != nil {
					target = targetTemp
				} else {
					return "", errors.New("no such key")
				}
				break
			}
			result = resultTemp
		}
	}
	if err != nil {
		return "", err
	}
	return target, nil
}

type txtHandler struct{}

// HandleResponse handle text type response, the second param is the no of ip to get
// if response is:
// ip1 ... ip NO. ... ipn, return ip NO.
func (t txtHandler) HandleResponse(source string, no string) (target any, err error) {
	if matches := Net.IpPattern.FindAllString(source, -1); matches != nil {
		i, err := strconv.Atoi(no)
		if err != nil {
			return "", err
		}
		if i < len(matches) {
			return matches[i], nil
		}
		return source, nil
	}
	return "", errors.New("no valid ip matched")
}

// Validate check Validation of Api setting
func (a *ApiGenerator) Validate() (err error) {
	// all errors will be checked, joined
	return errors.Join(err, a.validateResponseType(), a.validateURL(), a.validateMethod())
}

// LoadApiFromConfig load api from config, add to Net.ApiMap
// return error if missing key
// return error if api setting is invalid
func LoadApiFromConfig(sec *ini.Section) (ApiGenerator, error) {
	Ag := ApiGenerator{}

	names := []string{"A", "AAAA", "HTTPMethod", "Response", "Value"}
	for _, name := range names {
		if !sec.HasKey(name) {
			return Ag, MissingKeyErr{
				KeyName:     name,
				SectionName: "ApiGenerator",
			}
		} else {
			switch name {
			case "A":
				Ag.a = sec.Key(name).Value()
			case "AAAA":
				Ag.aaaa = sec.Key(name).Value()
			case "HTTPMethod":
				Ag.method = sec.Key(name).Value()
			case "Response":
				Ag.response = sec.Key(name).Value()
			case "Value":
				Ag.resName = sec.Key(name).Value()
			}
		}
	}

	err := Ag.Validate()
	if err != nil {
		return ApiGenerator{}, err
	}

	prefixes := [3]string{"Api.", "api.", "API."}
	for _, prefix := range prefixes {
		if strings.HasPrefix(sec.Name(), prefix) {
			Ag.apiName = strings.Replace(sec.Name(), prefix, "", 1)
			break
		}
	}

	return Ag, nil
}
