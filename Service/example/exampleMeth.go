// Package example is a template for creating new service
package example

import (
	"GodDns/DDNS"
	log "GodDns/Log"
	"GodDns/Net"
	"GodDns/Util"
	"gopkg.in/ini.v1"
	"strings"
)

// /////////// Pointer methods /////////// //

func (p *Parameter) GetName() string {
	return serviceName
}

func (p *Parameter) SaveConfig(No uint) (DDNS.ConfigStr, error) {
	return configInstance.GenerateConfigInfo(p, No)
}

func (p *Parameter) ToRequest() (DDNS.Request, error) {
	request := Request{
		Parameter: *p,
	}
	return &request, nil
}

func (r *Request) Target() string {
	return r.SubDomain + "." + r.Domain
}

func (p *Parameter) SetValue(s string) {
	p.IpToSet = s
}

func (p *Parameter) GetIP() string {
	return p.IpToSet
}

// GetType Note that return value of this method should be "4" or "6"
func (p *Parameter) GetType() string {
	return Net.Type2Num(p.Type)
}

func (p *Parameter) IsTypeSet() bool {
	return p.Type == "AAAA" || p.Type == "A"
}

// Run will call MakeRequest and log the result
// implements Cron.Job
func (r *Request) Run() {
	err := r.MakeRequest()
	log.Infof("status:%+v,err:%s", r.Status(), err)
}

func (r *Request) ToParameters() DDNS.Parameters {
	return &r.Parameter
}

func (r *Request) MakeRequest() error {
	// prepare request
	// get necessary info like RecordId

	// make ddns request
	// update status.Status using DDNS.Status or DDNS.Fail or DDNS.NotExecute
	// update status.Msg & status.Name
	// log.Infof("relevant info...")

	panic("implement me")
}

func (r *Request) Status() DDNS.Status {
	return r.status
}

// ////////////////////////////////////////////////////// //

func (c Config) GetName() string {
	return serviceName
}

// GenerateDefaultConfigInfo generate default config info
// used to generate default config file
func (c Config) GenerateDefaultConfigInfo() (DDNS.ConfigStr, error) {
	// use GenerateConfigInfo to generate default config
	return c.GenerateConfigInfo(&Parameter{
		Token:     "defaultToken",
		Domain:    "defaultDomain",
		SubDomain: "defaultSubDomain",
		RecordID:  "defaultRecordID",
		IpToSet:   "defaultIp",
		Type:      "A/AAAA/4/6",
	}, 0) // pass a default Parameter
}

// ReadConfig read config file from ini.Section
func (c Config) ReadConfig(sec ini.Section) ([]DDNS.Parameters, error) {
	// parameters' field names or key names in config file(if you modify the name by setting tag "KeyValue")
	names := [6]string{"Token", "Domain", "SubDomain", "RecordID", "IpToSet", "Type"}

	p := Parameter{}
	var subdomains []string
	for _, name := range names {
		if !sec.HasKey(name) {
			return nil, DDNS.NewMissKeyErr(name, serviceName)
		} else {
			switch name {
			case "SubDomain":
				// support value like this subdomain = `sub1,sub2,sub3` or `sub1 sub2 sub3`
				subdomain := sec.Key(name).String()
				subdomains = strings.Split(strings.ReplaceAll(subdomain, ",", " "), " ")
				Util.RemoveDuplicate(&subdomains) // remove duplicate subdomains, remind to pass pointer
			case "Type":
				p.Type = Net.Type2Str(sec.Key(name).String()) // convert "4"/"6"/"A"/"AAAA" to "A"/"AAAA"
			// case UnexportedField:
			// Set field

			// case OtherSpecials:
			// do something special
			default:
				// any other !EXPORTED! field with no special treatment
				err := Util.SetVariable(&p, name, sec.Key(name).String())
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// if subdomain value-string contains multiple subdomains, generate multiple parameters
	ps := make([]DDNS.Parameters, 0, len(subdomains))
	for _, subdomain := range subdomains {
		if subdomain == "" {
			continue
		}
		ps = append(ps, &Parameter{
			Token:     p.Token,
			Domain:    p.Domain,
			SubDomain: subdomain,
			RecordID:  p.RecordID,
			IpToSet:   p.IpToSet,
			Type:      p.Type,
		})
	}

	return ps, nil
}

func (c Config) GenerateConfigInfo(parameters DDNS.Parameters, u uint) (DDNS.ConfigStr, error) {
	// the first line is Section name
	// if it's the first section, the name looks like [example]
	// if it's not the first section, the name looks like [example#1] [example#2] ...
	head := DDNS.ConfigHead(parameters, u)

	body := Util.Convert2KeyValue(DDNS.Format, parameters)
	// the default Convert will convert struct to key-value string like

	//	type B struct {
	//		X string
	//		x string
	//	}
	//
	//	type A struct {
	//		Device     string `KeyValue:"device,device name" json:"device"`
	//		IP         string `json:"ip,omitempty,string"`
	//		Type       string
	//		unexported string
	//		B          B
	//	}
	//		a := A{Device: "device", IP: "ip", Type: "type", B: B{X: "123", x: "321"}}
	//		fmt.Println(Convert2KeyValue("%s: %s", a))
	//		output:
	//	 # device name
	//		device: device
	//		ip: ip
	//		Type: type
	//		B: {123 321}

	// if you want to modify the key name or the order of key-value pairs, or recursively convert struct like
	//		device: device
	//		ip: ip
	//		Type: type
	//		X: 123
	// PLZ implement ConvertableKeyValue interface
	// eg:
	//		func (a A) Convert2KeyValue(format string) string {
	//			return Convert2KeyValue(format, a)+Convert2KeyValue(format, a.B)
	//		}

	tail := "\n\n"

	return DDNS.ConfigStr{
		Name:    serviceName,
		Content: head + body + tail,
	}, nil
}

func (c ConfigFactory) GetName() string {
	return serviceName
}

// Get return a singleton Config
func (c ConfigFactory) Get() DDNS.Config {
	return &configInstance
}

func (c ConfigFactory) New() *DDNS.Config {
	var config DDNS.Config = &Config{}
	return &config
}
