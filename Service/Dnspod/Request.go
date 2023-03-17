/*
 *     @Copyright
 *     @file: Request.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/18 上午3:34
 *     @last modified: 2023/3/18 上午3:34
 *
 *
 *
 */

package Dnspod

import (
	"fmt"
	"github.com/Equationzhao/GodDns/DDNS"
	"github.com/Equationzhao/GodDns/Util"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
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

// Status return DDNS.Status which contains execution result etc.
func (r *Request) Status() DDNS.Status {
	return r.status
}

// ToParameters return DDNS.Parameters
func (r *Request) ToParameters() DDNS.Parameters {
	return &r.parameters
}

// Run implements Cron.Job
func (r *Request) Run() {
	err := r.MakeRequest()
	logrus.Debugf("status:%+v,err:%s", r.Status(), err)
}

// GetName return "dnspod"
func (r *Request) GetName() string {
	return serviceName
}

// Init set parameter
func (r *Request) Init(parameters DDNS.Parameters) error {
	r.parameters.PublicParameter = parameters.(*Parameters).PublicParameter
	r.parameters.ExternalParameter = parameters.(*Parameters).ExternalParameter
	return nil
}

// MakeRequest  1.GetRecordId  2.DDNS
func (r *Request) MakeRequest() error {
	status, err := r.GetRecordId()
	if err != nil || status.Success != DDNS.Success {
		r.status.Success = DDNS.Success
		r.status.Msg = status.Msg
		return err
	}

	s := &struct {
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
	}{}

	content := Util.Convert2XWWWFormUrlencoded(&r.parameters)
	logrus.Debug(content)
	client := resty.New()
	response, err := client.R().SetResult(s).SetHeader("Content-Type", "application/x-www-form-urlencoded").SetBody([]byte(content)).Post(DDNSUrl)
	logrus.Tracef("response: %v", response)
	logrus.Debugf("result:%+v", s)

	//r.status = *code2msg(s.Status.Code).AppendMsg(" ", s.Status.Message, "at ", s.Status.CreatedAt, " ", r.parameters.getTotalDomain(), " ", s.Record.Value)
	r.status = *code2msg(s.Status.Code).AppendMsgF(" %s at %s %s %s", s.Status.Message, s.Status.CreatedAt, r.parameters.getTotalDomain(), s.Record.Value)

	if err != nil {
		return err
	} else {
		return nil
	}
}

// GetRecordId make request to Dnspod to get RecordId and set ExternalParameter.RecordId
func (r *Request) GetRecordId() (DDNS.Status, error) {

	s := &struct {
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
	}{}

	p := struct {
		LoginToken   string `json:"login_token,omitempty" xwwwformurlencoded:"login_token"`
		Format       string `json:"format,omitempty" xwwwformurlencoded:"format"`
		Lang         string `json:"lang,omitempty" xwwwformurlencoded:"lang"`
		ErrorOnEmpty string `json:"error_on_empty,omitempty" xwwwformurlencoded:"error_on_empty"`
		Domain       string `json:"domain,omitempty" xwwwformurlencoded:"domain"`
		Subdomain    string `json:"sub_domain,omitempty" xwwwformurlencoded:"sub_domain"`
		Type         string `json:"record_type,omitempty" xwwwformurlencoded:"record_type"`
	}{
		LoginToken:   r.parameters.PublicParameter.LoginToken,
		Format:       r.parameters.PublicParameter.Format,
		Lang:         r.parameters.PublicParameter.Lang,
		ErrorOnEmpty: r.parameters.PublicParameter.ErrorOnEmpty,
		Type:         r.parameters.ExternalParameter.Type,
		Domain:       r.parameters.ExternalParameter.Domain,
		Subdomain:    r.parameters.ExternalParameter.Subdomain}

	content := Util.Convert2XWWWFormUrlencoded(p)
	logrus.Debug(content)

	// make request to "https://dnsapi.cn/Record.List" to get record id
	client := resty.New()
	response, err := client.R().SetResult(s).SetHeader("Content-Type", "application/x-www-form-urlencoded").SetBody([]byte(content)).Post(RecordListUrl)
	logrus.Tracef("response: %v", response)
	logrus.Debugf("result:%+v", s)
	status := *code2msg(s.Status.Code).AppendMsg(s.Status)
	if err != nil {
		return status, err
	}

	if s.Status.Code != "1" {
		return status, fmt.Errorf("status code is not 1, code:%s", s.Status.Code)
	}

	id, err := strconv.Atoi(s.Records[0].Id)

	if err != nil {
		return status, err
	}

	r.parameters.ExternalParameter.RecordId = uint32(id)
	return status, nil
}
