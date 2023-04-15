package dnspodyunapi

import (
	"fmt"
	"strconv"

	"GodDns/core"
	log "GodDns/log"
	json "GodDns/util/json"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

const api = "dnspod.tencentcloudapi.com"

type Request struct {
	Parameters DnspodYun
	status     core.Status
}

// Target return target domain
func (r *Request) Target() string {
	return r.Parameters.SubDomain + "." + r.Parameters.Domain
}

func newStatus() *core.Status {
	return &core.Status{
		Name:   serviceName,
		MG:     core.NewDefaultMsgGroup(),
		Status: core.NotExecute,
	}
}

func (r *Request) Init(yun DnspodYun) {
	r.Parameters = yun
}

func (r *Request) ToParameters() core.Service {
	return &r.Parameters
}

func (r *Request) GetName() string {
	return serviceName
}

func (r *Request) MakeRequest() error {
	r.status = *newStatus()

	credential := common.NewCredential(
		r.Parameters.SecretID,
		r.Parameters.SecretKey,
	)
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = api
	// 实例化要请求产品的client对象,clientProfile是可选的
	client, _ := dnspod.NewClient(credential, "", cpf)

	requestRecord := dnspod.NewDescribeRecordListRequest()
	requestRecord.Domain = common.StringPtr(r.Parameters.Domain)
	requestRecord.Subdomain = common.StringPtr(r.Parameters.SubDomain)
	requestRecord.RecordType = common.StringPtr(r.Parameters.Type)
	requestRecord.RecordLine = common.StringPtr(r.Parameters.RecordLine)

	var responseRecordId *dnspod.DescribeRecordListResponse
	// 返回的resp是一个DescribeRecordListResponse的实例，与请求对象对应
	errChan := make(chan error, 1)
	_ = core.MainGoroutinePool.Submit(func() {
		_responseRecordId, err := client.DescribeRecordList(requestRecord)
		if _, ok := err.(*errors.TencentCloudSDKError); ok {
			log.Debug("an API error has returned ", log.String("error", err.Error()).String())
			r.status.Status = core.Failed
			r.status.MG.AddError(err.(*errors.TencentCloudSDKError).Message)
			errChan <- err
		} else if err != nil {
			errChan <- err
		}
		errChan <- nil
		responseRecordId = _responseRecordId
	})

	// 实例化一个请求对象,每个接口都会对应一个request对象
	requestDDNS := dnspod.NewModifyDynamicDNSRequest()

	requestDDNS.Domain = common.StringPtr(r.Parameters.Domain)
	requestDDNS.SubDomain = common.StringPtr(r.Parameters.SubDomain)
	requestDDNS.RecordLine = common.StringPtr(r.Parameters.RecordLine)
	requestDDNS.Value = common.StringPtr(r.Parameters.Type)
	requestDDNS.Ttl = common.Uint64Ptr(r.Parameters.TTL)

	var id uint64
	if err := <-errChan; err != nil {
		return err
	}

	if len(responseRecordId.Response.RecordList) <= 1 {
		id = *responseRecordId.Response.RecordList[0].RecordId
		r.Parameters.RecordId = strconv.FormatUint(id, 10)
	}
	requestDDNS.RecordId = common.Uint64Ptr(id)

	// 返回的resp是一个ModifyDynamicDNSResponse的实例，与请求对象对应
	responseDDNS, err := client.ModifyDynamicDNS(requestDDNS)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		log.Debug("an API error has returned ", log.String("error", err.Error()).String())
		r.status.Status = core.Failed
		r.status.MG.AddError(err.(*errors.TencentCloudSDKError).Message)
		return fmt.Errorf("an API error has returned: %w", err)
	} else if err != nil {
		return err
	}

	res := res{}
	err = json.UnmarshalString(responseDDNS.ToJsonString(), &res)
	if err != nil {
		log.Debug("error unmarshalling response ",
			log.String("Json result", responseDDNS.ToJsonString()).String(),
			log.String("error", err.Error()).String())
		r.status.Status = core.Failed

		return fmt.Errorf("error umarshaling response %v: %w", responseDDNS.ToJsonString(), err)
	}

	// set status
	r.status.Status = core.Success
	r.status.MG.AddInfo(fmt.Sprintf("operation success %s %s", r.Parameters.getTotalDomain(), r.Parameters.Value))
	return nil
}

func (r *Request) Status() core.Status {
	return r.status
}

type res struct {
	Response struct {
		RecordId int `json:"RecordId"`
		Error    struct {
			Code    string `json:"Code"`
			Message string `json:"Message"`
		} `json:"Error"`
		RequestId string `json:"RequestId"`
	} `json:"Response"`
}
