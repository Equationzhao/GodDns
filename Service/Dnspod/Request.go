/*
 *
 *     @file: Request.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/31 下午3:16
 *     @last modified: 2023/3/31 下午1:57
 *
 *
 *
 */

/*
 *
 *     @file: Request.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/30 下午11:29
 *     @last modified: 2023/3/30 下午11:23
 *
 *
 *
 */

package Dnspod

import (
	"GodDns/DDNS"
	"GodDns/Net"
	"GodDns/Util"
	"errors"
	"fmt"
	"strconv"
	"time"

	log "GodDns/Log"

	"github.com/go-resty/resty/v2"
)

const (
	// RecordListUrl url of getting Record list
	RecordListUrl = "https://dnsapi.cn/Record.List"
	// DDNSUrl  url of DDNS
	DDNSUrl = "https://dnsapi.cn/Record.Ddns"
)

// usage
// r:=Dnspod.Request
// r.Init(Parameters)
// r.MakeRequest()

// Request implements DDNS.Request
type Request struct {
	parameters Parameters
	status     DDNS.Status
}

func (r *Request) Target() string {
	return r.parameters.Subdomain + "." + r.parameters.Domain
}

// Status return DDNS.Status which contains execution result etc.
func (r *Request) Status() DDNS.Status {
	return r.status
}

func newStatus() *DDNS.Status {
	return &DDNS.Status{
		Name:   serviceName,
		Msg:    "",
		Status: DDNS.NotExecute,
	}
}

// ToParameters return DDNS.Parameters
func (r *Request) ToParameters() DDNS.Parameters {
	return &r.parameters
}

// Run implements Cron.Job
func (r *Request) Run() {
	err := r.MakeRequest()
	log.Infof("status:%+v,err:%s", r.Status(), err)
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

func (r *Request) RequestThroughProxy() error {

	done := make(chan bool)
	status := newStatus()
	var err error
	go func(done chan bool) {
		*status, err = r.GetRecordIdByProxy(done)
	}(done)

	s := &resOfddns{}

	content := ""
	select {
	case <-done:
		if err != nil || status.Status != DDNS.Success {
			r.status.Name = serviceName
			r.status.Status = DDNS.Failed
			r.status.Msg = status.Msg
			return err
		}
		content = Util.Convert2XWWWFormUrlencoded(&r.parameters)
	case <-time.After(time.Second * 20):
		r.status.Status = DDNS.Timeout
		r.status.Msg = "GetRecordId timeout"
		return errors.New("GetRecordId timeout")
	}

	log.Debugf("content:%s", content)

	iter := Net.GlobalProxys.GetProxyIter()
	var response *resty.Response

	for iter.NotLast() {
		proxy := iter.Next()
		pool, err := DDNS.MainPoolMap.GetOrCreate(proxy, func() (resty.Client, error) {
			r := resty.New()
			r.SetProxy(proxy)
			return *r, nil
		})
		if err != nil {
			log.Error("error get client pool from map", log.String("error", err.Error()))
		} else {
			client := pool.Get()
			response, err = client.First.R().SetResult(s).SetHeader("Content-Type", "application/x-www-form-urlencoded").SetBody([]byte(content)).Post(DDNSUrl)
			log.Tracef("response: %v", response)
			log.Debugf("result:%+v", s)
			client.Release()
			if err != nil {
				log.Errorf("request error through proxy %s: %v", proxy, err)
				continue
			} else {
				break
			}
		}

	}
	r.status = *code2msg(s.Status.Code).AppendMsgF(" %s at %s %s %s", s.Status.Message, s.Status.CreatedAt, r.parameters.getTotalDomain(), s.Record.Value)

	if err != nil {
		return err
	} else {
		return nil
	}
}

// MakeRequest  1.GetRecordId  2.DDNS
func (r *Request) MakeRequest() error {
	done := make(chan bool)
	status := newStatus()
	var err error
	go func(done chan bool) {
		*status, err = r.GetRecordId(done)
	}(done)

	s := &resOfddns{}

	content := ""
	select {
	case <-done:
		if err != nil || status.Status != DDNS.Success {
			r.status.Name = serviceName
			r.status.Status = DDNS.Failed
			r.status.Msg = status.Msg
			return err
		}
		content = Util.Convert2XWWWFormUrlencoded(&r.parameters)
	case <-time.After(time.Second * 20):
		r.status.Status = DDNS.Timeout
		r.status.Msg = "GetRecordId timeout"
		return errors.New("GetRecordId timeout")
	}

	log.Debugf("content:%s", content)
	client := resty.New()
	response, err := client.R().SetResult(s).SetHeader("Content-Type", "application/x-www-form-urlencoded").SetBody([]byte(content)).Post(DDNSUrl)
	log.Tracef("response: %v", response)
	log.Debugf("result:%+v", s)

	// r.status = *code2msg(s.Status.Code).AppendMsg(" ", s.Status.Message, "at ", s.Status.CreatedAt, " ", r.parameters.getTotalDomain(), " ", s.Record.Value)
	r.status = *code2msg(s.Status.Code).AppendMsgF(" %s at %s %s %s", s.Status.Message, s.Status.CreatedAt, r.parameters.getTotalDomain(), s.Record.Value)

	if err != nil {
		return err
	} else {
		return nil
	}

}

// GetRecordId make request to Dnspod to get RecordId and set ExternalParameter.RecordId
func (r *Request) GetRecordId(done chan<- bool) (DDNS.Status, error) {
	defer func() {
		done <- true
	}()
	s := &resOfRecordId{}

	p := param2GetId{
		LoginToken:   r.parameters.LoginToken,
		Format:       r.parameters.Format,
		Lang:         r.parameters.Lang,
		ErrorOnEmpty: r.parameters.ErrorOnEmpty,
		Type:         r.parameters.Type,
		Domain:       r.parameters.Domain,
		Subdomain:    r.parameters.Subdomain,
	}

	content := Util.Convert2XWWWFormUrlencoded(p)
	log.Debugf("content:%s", content)

	// make request to "https://dnsapi.cn/Record.List" to get record id
	client := resty.New()

	response, err := client.R().SetResult(s).SetHeader("Content-Type", "application/x-www-form-urlencoded").SetBody(content).Post(RecordListUrl)
	log.Tracef("response: %v", response)
	log.Debugf("result:%+v", s)
	status := *code2msg(s.Status.Code).AppendMsgF(" %s at %s %s", s.Status.Message, s.Status.CreatedAt, r.parameters.getTotalDomain())
	if err != nil {
		return status, err
	}

	if s.Status.Code != "1" {
		return status, fmt.Errorf("status code is not 1, code:%s", s.Status.Code)
	}

	if len(s.Records) == 0 {
		return status, fmt.Errorf("no record found")
	}

	id, err := strconv.Atoi(s.Records[0].Id) // todo what if s.Records is empty

	if err != nil {
		return status, err
	}

	r.parameters.RecordId = uint32(id)
	return status, nil
}

func (r *Request) GetRecordIdByProxy(done chan<- bool) (DDNS.Status, error) {
	defer func() {
		done <- true
	}()
	s := &resOfRecordId{}

	p := param2GetId{
		LoginToken:   r.parameters.LoginToken,
		Format:       r.parameters.Format,
		Lang:         r.parameters.Lang,
		ErrorOnEmpty: r.parameters.ErrorOnEmpty,
		Type:         r.parameters.Type,
		Domain:       r.parameters.Domain,
		Subdomain:    r.parameters.Subdomain,
	}

	content := Util.Convert2XWWWFormUrlencoded(p)
	log.Debugf("content:%s", content)

	status := newStatus()

	// make request to "https://dnsapi.cn/Record.List" to get record id
	iter := Net.GlobalProxys.GetProxyIter()
	for iter.NotLast() {
		proxy := iter.Next()
		pool, err := DDNS.MainPoolMap.GetOrCreate(proxy, func() (resty.Client, error) {
			r := resty.New()
			r.SetProxy(proxy)
			return *r, nil
		})
		if err != nil {
			log.Error("error get client pool from map", log.String("error", err.Error()))
		} else {
			client := pool.Get()
			response, err := client.First.R().SetResult(s).SetHeader("Content-Type", "application/x-www-form-urlencoded").SetBody(content).Post(RecordListUrl)
			log.Tracef("response: %v", response)
			log.Debugf("result:%+v", s)

			if err != nil {
				log.Errorf("error getting record id by proxy %s", proxy)
				continue
			} else {
				break
			}
		}
	}
	status = code2msg(s.Status.Code).AppendMsgF(" %s at %s %s", s.Status.Message, s.Status.CreatedAt, r.parameters.getTotalDomain())
	if s.Status.Code != "1" {
		return *status, fmt.Errorf("status code is not 1, code:%s", s.Status.Code)
	}

	if len(s.Records) == 0 {
		return *status, fmt.Errorf("no record found")
	}

	id, err := strconv.Atoi(s.Records[0].Id)

	if err != nil {
		return *status, err
	}

	r.parameters.RecordId = uint32(id)
	return *status, nil
}

type param2GetId struct {
	LoginToken   string `json:"login_token,omitempty" xwwwformurlencoded:"login_token"`
	Format       string `json:"format,omitempty" xwwwformurlencoded:"format"`
	Lang         string `json:"lang,omitempty" xwwwformurlencoded:"lang"`
	ErrorOnEmpty string `json:"error_on_empty,omitempty" xwwwformurlencoded:"error_on_empty"`
	Domain       string `json:"domain,omitempty" xwwwformurlencoded:"domain"`
	Subdomain    string `json:"sub_domain,omitempty" xwwwformurlencoded:"sub_domain"`
	Type         string `json:"record_type,omitempty" xwwwformurlencoded:"record_type"`
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
