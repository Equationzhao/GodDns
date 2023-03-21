/*
 *     @Copyright
 *     @file: Request.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/22 上午6:29
 *     @last modified: 2023/3/22 上午6:21
 *
 *
 *
 */

package DnspodYunApi

import (
	"GodDns/DDNS"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
	"strconv"
)

const api = "dnspod.tencentcloudapi.com"

type Request struct {
	Parameters DnspodYun
	status     DDNS.Status
}

func newStatus() *DDNS.Status {
	return &DDNS.Status{
		Name:    serviceName,
		Msg:     "",
		Success: DDNS.NotExecute,
	}
}

func (r *Request) Init(yun DnspodYun) {
	r.Parameters = yun
}

func (r *Request) Run() {
	err := r.MakeRequest()
	logrus.Debugf("status:%+v,err:%s", r.Status(), err)
}

func (r *Request) ToParameters() DDNS.Parameters {
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

	// 返回的resp是一个DescribeRecordListResponse的实例，与请求对象对应
	responseRecordId, err := client.DescribeRecordList(requestRecord)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		logrus.Debugf("an API error has returned: %s", err.Error())
		r.status.Success = DDNS.Failed
		r.status.Msg = err.(*errors.TencentCloudSDKError).Message
		return fmt.Errorf("an API error has returned: %w", err)
	} else if err != nil {
		return err
	}
	// 输出json格式的字符串回包
	var id uint64
	if len(responseRecordId.Response.RecordList) <= 1 {
		id = *responseRecordId.Response.RecordList[0].RecordId
		r.Parameters.RecordId = strconv.FormatUint(id, 10)
	}

	// 实例化一个请求对象,每个接口都会对应一个request对象
	requestDDNS := dnspod.NewModifyDynamicDNSRequest()

	requestDDNS.Domain = common.StringPtr(r.Parameters.Domain)
	requestDDNS.SubDomain = common.StringPtr(r.Parameters.SubDomain)
	requestDDNS.RecordId = common.Uint64Ptr(id)
	requestDDNS.RecordLine = common.StringPtr(r.Parameters.RecordLine)
	requestDDNS.Value = common.StringPtr(r.Parameters.Type)
	requestDDNS.Ttl = common.Uint64Ptr(r.Parameters.TTL)

	// 返回的resp是一个ModifyDynamicDNSResponse的实例，与请求对象对应
	responseDDNS, err := client.ModifyDynamicDNS(requestDDNS)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		logrus.Debugf("an API error has returned: %s", err.Error())
		r.status.Success = DDNS.Failed
		r.status.Msg = err.(*errors.TencentCloudSDKError).Message
		return fmt.Errorf("an API error has returned: %w", err)
	} else if err != nil {
		return err
	}

	res := res{}
	err = json.Unmarshal([]byte(responseDDNS.ToJsonString()), &res)
	if err != nil {
		logrus.Debugf("error umarshaling response %v: %s", responseDDNS.ToJsonString(), err.Error())
		r.status.Success = DDNS.Failed

		return fmt.Errorf("error umarshaling response %v: %w", responseDDNS.ToJsonString(), err)
	}

	// set status
	r.status.Success = DDNS.Success
	r.status.Msg = "operation success"
	r.status.AppendMsg(fmt.Sprintf(" %s %s", r.Parameters.getTotalDomain(), r.Parameters.Value))
	return nil
}

func (r *Request) Status() DDNS.Status {
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
