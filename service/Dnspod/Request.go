package Dnspod

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"GodDns/core"
	log "GodDns/log"
	"GodDns/net"
	json "GodDns/util/json"
	"github.com/go-resty/resty/v2"
)

const (
	// RecordListUrl url of getting Record list
	RecordListUrl = "https://dnsapi.cn/Record.List"
	// DDNSURL  url of DDNS
	DDNSURL = "https://dnsapi.cn/Record.Ddns"

	fatalStr = "Fatal"
)

type empty struct{}

// usage
// r:=Dnspod.Request
// r.Init(Parameters)
// r.MakeRequest()

// Request implements DDNS.Request
type Request struct {
	parameters Parameters
	status     core.Status
}

// Target return target domain
func (r *Request) Target() string {
	return r.parameters.Subdomain + "." + r.parameters.Domain
}

// Status return DDNS.Status which contains execution result etc.
func (r *Request) Status() core.Status {
	return r.status
}

func newStatus() *core.Status {
	return &core.Status{
		Name:   serviceName,
		Status: core.NotExecute,
		MG:     core.NewDefaultMsgGroup(),
	}
}

// ToParameters return DDNS.Parameters
func (r *Request) ToParameters() core.Service {
	return &r.parameters
}

// GetName return "dnspod"
func (r *Request) GetName() string {
	return serviceName
}

// Init set parameter
func (r *Request) Init(parameters Parameters) error {
	r.parameters = parameters

	return nil
}

func (r *Request) encodeURLWithoutIDContent() url.Values {
	v := url.Values{}
	v.Add("login_token", r.parameters.LoginToken)
	v.Add("format", r.parameters.Format)
	v.Add("lang", r.parameters.Lang)
	v.Add("error_on_empty", r.parameters.ErrorOnEmpty)
	v.Add("domain", r.parameters.Domain)
	v.Add("sub_domain", r.parameters.Subdomain)
	v.Add("record_line", r.parameters.RecordLine)

	v.Add("record_type", r.parameters.Type)
	return v
}

func (r *Request) encodeURLWithoutID() (content string) {
	content = r.encodeURLWithoutIDContent().Encode()
	return content
}

func (r *Request) encodeURL() (content string) {
	v := r.encodeURLWithoutIDContent()
	ttl := strconv.Itoa(int(r.parameters.TTL))
	v.Add("ttl", ttl)
	v.Add("value", r.parameters.Value)
	v.Add("record_id", r.parameters.RecordId)
	content = v.Encode()
	return content
}

func (r *Request) RequestThroughProxy() error {
	done := make(chan empty)
	status := newStatus()
	var err error
	_ = core.MainGoroutinePool.Submit(func() {
		*status, err = r.GetRecordIdByProxy()
		done <- empty{}
	})

	s := &resOfddns{}

	var content string
	select {
	case <-done:
		if err != nil || status.Status != core.Success {
			r.status.Name = serviceName
			r.status.Status = core.Failed
			for _, i := range status.MG.GetInfo() {
				r.status.MG.AddInfo(i.String())
			}

			for _, i := range status.MG.GetWarn() {
				r.status.MG.AddWarn(i.String())
			}

			for _, i := range status.MG.GetError() {
				r.status.MG.AddError(i.String())
			}

			r.status.MG.AddError(err.Error())
			return err
		}
		// content = Util.Convert2XWWWFormUrlencoded(&r.parameters)
		content = r.encodeURL()

	case <-time.After(time.Second * 20):
		r.status.Status = core.Timeout
		r.status.MG.AddError("GetRecordId timeout")
		return errors.New("GetRecordId timeout")
	}

	log.Debugf("content:%s", content)

	iter := Net.GlobalProxies.GetProxyIter()
	client := core.MainClientPool.Get().(*resty.Client)
	defer core.MainClientPool.Put(client)
	req := client.R()
	for iter.NotLast() {
		proxy := iter.Next()
		response, err := req.
			SetResult(s).
			SetHeader("Content-Type", "application/x-www-form-urlencoded").
			SetBody([]byte(content)).
			Post(DDNSURL)
		if err != nil {
			errMsg := fmt.Sprintf("request error through proxy %s: %v", proxy, err)
			r.status.MG.AddError(errMsg)
			log.Errorf(errMsg)
			continue
		} else {
			log.Debugf("result:%+v", string(response.Body()))
			_ = json.Unmarshal(response.Body(), s)
			log.Debugf("after marshall:%+v", s)
			break
		}
	}
	r.status = *code2status(s.Status.Code)
	if s.Status.Message == "" {
		s.Status.Message = fatalStr
	}
	resultMsg := fmt.Sprintf("%s at %s %s %s", s.Status.Message, s.Status.CreatedAt, r.parameters.getTotalDomain(), s.Record.Value)
	if r.status.Status == core.Success {
		r.status.MG.AddInfo(resultMsg)
	} else {
		r.status.MG.AddError(resultMsg)
	}

	if err != nil {
		return err
	} else {
		return nil
	}
}

// MakeRequest  1.GetRecordId  2.DDNS
func (r *Request) MakeRequest() error {
	done := make(chan struct{})
	status := newStatus()
	var err error
	_ = core.MainGoroutinePool.Submit(func() {
		*status, err = r.GetRecordId()
		done <- empty{}
	})

	s := &resOfddns{}

	var content string
	select {
	case <-done:
		if err != nil || status.Status != core.Success {
			r.status.Name = serviceName
			r.status.Status = core.Failed
			for _, i := range status.MG.GetInfo() {
				r.status.MG.AddInfo(i.String())
			}

			for _, i := range status.MG.GetWarn() {
				r.status.MG.AddWarn(i.String())
			}

			for _, i := range status.MG.GetError() {
				r.status.MG.AddError(i.String())
			}
			r.status.MG.AddError(err.Error())
			return err
		}
		// content = Util.Convert2XWWWFormUrlencoded(&r.parameters)
		content = r.encodeURL()

	case <-time.After(time.Second * 20):
		r.status.Status = core.Timeout
		r.status.MG.AddError("GetRecordId timeout")
		return errors.New("GetRecordId timeout")
	}

	log.Debugf("content:%s", content)
	client := core.MainClientPool.Get().(*resty.Client)
	defer core.MainClientPool.Put(client)
	response, err := client.R().
		SetResult(s).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetBody([]byte(content)).
		Post(DDNSURL)
	log.Tracef("response: %v", response)
	log.Debugf("result:%+v", s)
	_ = json.Unmarshal(response.Body(), s)
	log.Debugf("after marshall:%+v", s)
	r.status = *code2status(s.Status.Code)
	if s.Status.Message == "" {
		s.Status.Message = fatalStr
	}
	resultMsg := fmt.Sprintf("%s at %s %s %s", s.Status.Message, s.Status.CreatedAt, r.parameters.getTotalDomain(), s.Record.Value)
	if r.status.Status == core.Success {
		r.status.MG.AddInfo(resultMsg)
	} else {
		r.status.MG.AddError(resultMsg)
	}
	if err != nil {
		return err
	} else {
		return nil
	}
}

// GetRecordId make request to Dnspod to get RecordId and set ExternalParameter.RecordId
func (r *Request) GetRecordId() (core.Status, error) {
	if r.status.MG == nil {
		r.status.MG = core.NewDefaultMsgGroup()
	}

	s := &resOfRecordId{}

	content := r.encodeURLWithoutID()

	log.Debugf("content:%s", content)

	// make request to "https://dnsapi.cn/Record.List" to get record id
	client := core.MainClientPool.Get().(*resty.Client)
	defer core.MainClientPool.Put(client)
	_, err := client.R().
		SetResult(s).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetBody(content).
		Post(RecordListUrl)

	log.Debugf("after marshall:%s", s)
	status := *code2status(s.Status.Code)
	if err != nil {
		if s.Status.Message == "" {
			s.Status.Message = fatalStr
		}
		status.MG.AddError(fmt.Sprintf("%s at %s %s", s.Status.Message, s.Status.CreatedAt, r.parameters.getTotalDomain()))
		return status, err
	}

	if s.Status.Code != "1" {
		if s.Status.Code == "" {
			return status, errors.New("status code is empty")
		} else {
			status.MG.AddError(fmt.Sprintf("%s at %s %s", s.Status.Message, s.Status.CreatedAt, r.parameters.getTotalDomain()))
			return status, fmt.Errorf("status code:%s", s.Status.Code)
		}
	}

	if len(s.Records) == 0 {
		status.MG.AddError(fmt.Sprintf("%s at %s %s", s.Status.Message, s.Status.CreatedAt, r.parameters.getTotalDomain()))
		return status, fmt.Errorf("no record found")
	}

	status.MG.AddInfo(fmt.Sprintf("%s at %s %s", s.Status.Message, s.Status.CreatedAt, r.parameters.getTotalDomain()))
	r.parameters.RecordId = s.Records[0].Id
	return status, nil
}

func (r *Request) GetRecordIdByProxy() (core.Status, error) {
	if r.status.MG == nil {
		r.status.MG = core.NewDefaultMsgGroup()
	}
	s := &resOfRecordId{}

	content := r.encodeURLWithoutID()
	log.Debugf("content:%s", content)

	client := core.MainClientPool.Get().(*resty.Client)
	defer core.MainClientPool.Put(client)
	res := client.R().SetHeader("Content-Type", "application/x-www-form-urlencoded")
	// make request to "https://dnsapi.cn/Record.List" to get record id
	iter := Net.GlobalProxies.GetProxyIter()
	for iter.NotLast() {
		proxy := iter.Next()
		_, err := res.SetBody(content).SetResult(s).Post(RecordListUrl)
		log.Debugf("after marshall:%s", s)
		if err == nil {
			break
		}

		errMsg := fmt.Sprintf("error get record id by proxy %s, error:%s", proxy, err.Error())
		r.status.MG.AddError(errMsg)
		log.ErrorRaw(errMsg)
		continue
	}
	status := code2status(s.Status.Code) // " %s at %s %s", s.Status.Message, s.Status.CreatedAt, r.parameters.getTotalDomain()

	if s.Status.Code != "1" {
		if s.Status.Code == "" {
			return *status, errors.New("status code is empty")
		} else {
			status.MG.AddError(fmt.Sprintf("%s at %s %s", s.Status.Message, s.Status.CreatedAt, r.parameters.getTotalDomain()))
			return *status, fmt.Errorf("status code:%s", s.Status.Code)
		}
	}

	if len(s.Records) == 0 {
		status.MG.AddError(fmt.Sprintf("%s at %s %s", s.Status.Message, s.Status.CreatedAt, r.parameters.getTotalDomain()))
		return *status, fmt.Errorf("no record found")
	}

	status.MG.AddInfo(fmt.Sprintf("%s at %s %s", s.Status.Message, s.Status.CreatedAt, r.parameters.getTotalDomain()))
	r.parameters.RecordId = s.Records[0].Id
	return *status, nil
}

type resOfRecordId struct {
	Status struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		CreatedAt string `json:"created_at"`
	} `json:"status"`

	Records []struct {
		Id            string `json:"id"`
		Ttl           string `json:"ttl"`
		Value         string `json:"value"`
		Enabled       string `json:"enabled"`
		Status        string `json:"status"`
		UpdatedOn     string `json:"updated_on"`
		RecordTypeV1  string `json:"record_type_v1"`
		Name          string `json:"name"`
		Line          string `json:"line"`
		LineId        string `json:"line_id"`
		Type          string `json:"type"`
		Weight        any    `json:"weight"`
		MonitorStatus string `json:"monitor_status"`
		Remark        string `json:"remark"`
		UseAqb        string `json:"use_aqb"`
		Mx            string `json:"mx"`
	} `json:"records"`
}

type resOfddns struct {
	Status struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		CreatedAt string `json:"created_at"`
	} `json:"status"`
	Record struct {
		Id    int    `json:"id"`
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"record"`
}
