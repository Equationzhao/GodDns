package Dnspod

import "GodDns/Core"

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
	InvalidProxy        = "-3" // proxy only
	NotUnderProxy       = "-4" // proxy only
	APIPermissionDenied = "-7"
	TemporarilyBaned    = "-8"
	LoginRegionLimited  = "85"
	FunctionClosed      = "-99"
	Success             = "1"
	PostOnly            = "2"
	UnknownError        = "3"

	// BadUserId                  = 6 //proxy only
	// BadUserOwner               = 7 //proxy only

	AccountLocked = "83"
)

// code2status
// convert the code to message and set status.Status
func code2status(code string) *DDNS.Status {
	var msg = newStatus()
	switch code {
	case Success:
		msg.MG.AddInfo("接口调用成功")
	case BanedDomain:
		msg.MG.AddError("域名被封禁")
	case BadDomainId:
		msg.MG.AddError("域名 ID 错误")
	case BadDomainOwner:
		msg.MG.AddError("域名不属于您")
	case BadDomain:
		msg.MG.AddError("域名不存在")
	case IncorrectRecordValue:
		msg.MG.AddError("记录值非法")
	case LockedDomain:
		msg.MG.AddError("域名被锁定")
	case InvalidSubdomain:
		msg.MG.AddError("子域名非法")
	case SubdomainLevelOverRange:
		msg.MG.AddError("子域名级数超出限制")
	case SubdomainUniversalParsingError:
		msg.MG.AddError("子域名通配符解析错误")
	case TypeAOverLimit:
		msg.MG.AddError("A 记录负载均衡超出限制")
	case TypeCNAMEOverLimit:
		msg.MG.AddError("CNAME 记录负载均衡超出限制")
	case RecordLineError:
		msg.MG.AddError("记录线路非法")
	case LoginError:
		msg.MG.AddError("用户未登录")
	case APIOverLimit:
		msg.MG.AddError("请求次数超过限制")
	case InvalidProxy:
		msg.MG.AddError("不是合法代理")
	case NotUnderProxy:
		msg.MG.AddError("不在代理名下")
	case APIPermissionDenied:
		msg.MG.AddError("无接口权限")
	case TemporarilyBaned:
		msg.MG.AddError("用户被封禁")
	case FunctionClosed:
		msg.MG.AddError("接口已关闭")

	case PostOnly:
		msg.MG.AddError("请求方法错误")
	case UnknownError:
		msg.MG.AddError("未知错误")
	// case BadUserId:
	//	msg.msg = "用户 ID 错误"
	// case BadUserOwner:
	//	msg.msg = "用户不属于您"
	case AccountLocked:
		msg.MG.AddError("账号被锁定")
	case LoginRegionLimited:
		msg.MG.AddError("用户登录地异常或该帐户开启了登录区域保护，当前IP不在允许的区域内。")
	default:
		msg.MG.AddError("未知错误")
	}

	if code == "1" {
		msg.Status = DDNS.Success
	} else {
		msg.Status = DDNS.Failed
	}

	return msg
}
