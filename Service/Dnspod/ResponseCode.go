/*
 *     @Copyright
 *     @file: ResponseCode.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/18 上午3:43
 *     @last modified: 2023/3/18 上午3:42
 *
 *
 *
 */

package Dnspod

import "GodDns/DDNS"

const (
	BanedDomain                    = "-15"
	BadDomainId                    = "6"
	BadDomainOwner                 = "7"
	BadDomain                      = "8"
	IncorrectRecordValue           = "17"
	LockedDomain                   = "21"
	InvalidSubdomain               = "22"
	SubdomainLevelOverRange        = "23"
	SubdomainUniversalParsingError = "24"
	TypeAOverLimit                 = "500025"
	TypeCNAMEOverLimit             = "500026"
	RecordLineError                = "26"

	LoginError          = "-1"
	APIOverLimit        = "-2"
	InvalidProxy        = "-3" //proxy only
	NotUnderProxy       = "-4" //proxy only
	APIPermissionDenied = "-7"
	TemporarilyBaned    = "-8"
	LoginRegionLimited  = "85"
	FunctionClosed      = "-99"
	Success             = "1"
	PostOnly            = "2"
	UnknownError        = "3"

	//BadUserId                  = 6 //proxy only
	//BadUserOwner               = 7 //proxy only

	AccountLocked = "83"
)

// code2msg
// convert the code to message and set status.Success
func code2msg(code string) *DDNS.Status {
	var msg = new(DDNS.Status)
	msg.Name = "dnspod"
	switch code {
	case Success:
		msg.Msg = "接口调用成功"
	case BanedDomain:
		msg.Msg = "域名被封禁"
	case BadDomainId:
		msg.Msg = "域名 ID 错误"
	case BadDomainOwner:
		msg.Msg = "域名不属于您"
	case BadDomain:
		msg.Msg = "域名不存在"
	case IncorrectRecordValue:
		msg.Msg = "记录值非法"
	case LockedDomain:
		msg.Msg = "域名被锁定"
	case InvalidSubdomain:
		msg.Msg = "子域名非法"
	case SubdomainLevelOverRange:
		msg.Msg = "子域名级数超出限制"
	case SubdomainUniversalParsingError:
		msg.Msg = "子域名通配符解析错误"
	case TypeAOverLimit:
		msg.Msg = "A 记录负载均衡超出限制"
	case TypeCNAMEOverLimit:
		msg.Msg = "CNAME 记录负载均衡超出限制"
	case RecordLineError:
		msg.Msg = "记录线路非法"
	case LoginError:
		msg.Msg = "用户未登录"
	case APIOverLimit:
		msg.Msg = "请求次数超过限制"
	case InvalidProxy:
		msg.Msg = "不是合法代理"
	case NotUnderProxy:
		msg.Msg = "不在代理名下"
	case APIPermissionDenied:
		msg.Msg = "无接口权限"
	case TemporarilyBaned:
		msg.Msg = "用户被封禁"
	case FunctionClosed:
		msg.Msg = "接口已关闭"

	case PostOnly:
		msg.Msg = "请求方法错误"
	case UnknownError:
		msg.Msg = "未知错误"
	//case BadUserId:
	//	msg.msg = "用户 ID 错误"
	//case BadUserOwner:
	//	msg.msg = "用户不属于您"
	case AccountLocked:
		msg.Msg = "账号被锁定"
	case LoginRegionLimited:
		msg.Msg = "用户登录地异常或该帐户开启了登录区域保护，当前IP不在允许的区域内。"
	default:
		msg.Msg = "未知错误"
	}

	if code == "1" {
		msg.Success = DDNS.Success
	} else {
		msg.Success = DDNS.Failed
	}

	return msg
}
