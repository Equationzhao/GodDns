package Net

import (
	"errors"
	"fmt"
	"math"
	"net"
	"net/netip"
	"regexp"

	"github.com/go-resty/resty/v2"
)

const (
	ipv4Regex = `^(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.|$)){4})`
	ipv6Regex = `
				^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:)
				{1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}
				(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|
				([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}
				(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|
				:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]+
				|::(ffff(:0{1,4})?:)?((25[0-5]|(2[0-4]|1?[0-9])?[0-9])\.){3}(25[0-5]|
				(2[0-4]|1?[0-9])?[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1?[0-9])?[0-9])\.)
				{3}(25[0-5]|(2[0-4]|1?[0-9])?[0-9]))$`
)

// var Ipv4Pattern = regexp.MustCompile(ipv4Regex)
// var Ipv6Pattern = regexp.MustCompile(ipv6Regex)

var IpPattern = regexp.MustCompile(ipv4Regex + "|" + ipv6Regex)

// Api is a type = function that return a string and an error
type Api struct {
	Get func(Type) (string, error)
}

// Apis contains a map of apis
type Apis struct {
	a map[string]Api
}

var getIPFromIdentMeApi = Api{
	Get: getIPFromIdentMe,
}

var getIPFromIpifyApi = Api{
	Get: getIPFromIpify,
}

// ApiMap is a default Apis, contains a map of apis
var ApiMap = Apis{
	a: map[string]Api{
		"ipify":   getIPFromIpifyApi,
		"identMe": getIPFromIdentMeApi,
	},
}

// GetApiName return the names of apis
func (a *Apis) GetApiName() []string {
	res := make([]string, 0, len(a.a))
	for s := range a.a {
		res = append(res, s)
	}
	return res
}

// Add2Apis add api to Map
func (a *Apis) Add2Apis(name string, f Api) {
	a.a[name] = f
}

// GetApi return the api function
func (a *Apis) GetApi(name string) (Api, error) {
	api, ok := a.a[name]
	if !ok {
		return Api{}, errors.New("not found")
	}
	return api, nil
}

// GetMap return the map of apis
func (a *Apis) GetMap() map[string]Api {
	return a.a
}

// getIPFromIpify get ip from ipify
func getIPFromIpify(Type uint8) (string, error) {
	var ApiUri string
	switch Type {
	case AAAA:
		ApiUri = "https://api6.ipify.org"
	case A:
		ApiUri = "https://api.ipify.org"
	default:
		return "", NewUnknownType(Type)
	}
	res, err := resty.New().R().Get(ApiUri)
	if err != nil || res.String() == "" {
		return "", err
	}
	return res.String(), nil
}

// getIPFromIdentMe get ip from ident.me
func getIPFromIdentMe(Type uint8) (string, error) {
	var ApiUri string
	switch Type {
	case AAAA:
		ApiUri = "https://v6.ident.me"
	case A:
		ApiUri = "https://v4.ident.me"
	default:
		return "", NewUnknownType(Type)
	}
	res, err := resty.New().R().Get(ApiUri)
	if err != nil || res.String() == "" {
		return "", err
	}
	return res.String(), nil
}

// -------------------------------------------------------- //

// GetIp return Ip list of corresponding interface and nil error when error occurs, return nil and error
func GetIp(nameToMatch string) ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var ips []string

	for _, i := range interfaces {
		if i.Name == nameToMatch {
			address, err := i.Addrs()
			if err != nil {
				return nil, err
			}
			for _, addr := range address {
				ips = append(ips, addr.(*net.IPNet).IP.String())
			}
			return ips, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

// GetIpByType get specific type ip of corresponding interface
// parameter: NameToMatch, Type
// NameToMatch: interface name
// Type: 4 for ipv4, 6 for ipv6 (use constant A and AAAA)
// return: ip list, error
// when error occurs, return nil and error
func GetIpByType(nameToMatch string, Type uint8) ([]string, error) {
	if Type != A && Type != AAAA {
		return nil, fmt.Errorf("invalid type: %d", Type)
	} else {
		ips, err := GetIp(nameToMatch)
		res := make([]string, 0, len(ips))
		if err != nil {
			return nil, err
		}

		for _, ip := range ips {
			if WhichType(ip) == Type {
				res = append(res, ip)
			}
		}

		return res, nil
	}
}

// ---------------------------------------------------------- //

const (
	A    uint8 = 4
	AAAA uint8 = 6
)

type Type = uint8

type UnknownType struct {
	Type uint8
}

func NewUnknownType(Type uint8) *UnknownType {
	return &UnknownType{Type: Type}
}

func (u UnknownType) Error() string {
	return fmt.Sprintf("unknown type: %d", u.Type)
}

// WhichType get the type of ip
// parameter: ip
// if ip is an ipv4 address return 4, if it's an ipv6 address return 6, else return 0
func WhichType(ip string) uint8 {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return 0
	}
	if netip.Addr.Is4(addr) {
		return 4
	} else if netip.Addr.Is6(addr) {
		return 6
	}

	return 0
}

// WhichTypeStr get the type of ip in string
// parameter: ip
func WhichTypeStr(ip string) string {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return ""
	}
	if netip.Addr.Is4(addr) {
		return "A"
	} else if netip.Addr.Is6(addr) {
		return "AAAA"
	}

	return ""
}

type IntegerNumeric interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64
}

func moreThan[T IntegerNumeric](a T, b T) bool {
	return a > b
}

// TypeEqual compare t1 and t2
// if t1 = 4 or "A", t2 = 4 or "A" ,return true
// if t1 = 6 or "AAAA", t2 = 6 or "AAAA", return true
func TypeEqual(t1, t2 any) bool {
	// if t1/t2 is string, convert it to uint8
	// the compare uint8(t1) and uint8(t2)

	TypeStrToUint8 := func(v string) uint8 {
		switch v {
		case "A", "4":
			return A
		case "AAAA", "6":
			return AAAA
		default:
			return 0
		}
	}

	var v1, v2 uint8
	switch v := t1.(type) {
	case string:
		v1 = TypeStrToUint8(v)
	case uint8:
		if moreThan(v, math.MaxUint8) {
			return false
		}
		v1 = v
	case uint:
		if moreThan(v, math.MaxUint8) {
			return false
		}
		v1 = uint8(v)
	case uint16:
		if moreThan(v, math.MaxUint8) {
			return false
		}
		v1 = uint8(v)
	case uint32:
		if moreThan(v, math.MaxUint8) {
			return false
		}
		v1 = uint8(v)
	case uint64:
		if moreThan(v, math.MaxUint8) {
			return false
		}
		v1 = uint8(v)
	case int:
		if moreThan(v, math.MaxUint8) {
			return false
		}
		v1 = uint8(v)
	case int16:
		if moreThan(v, math.MaxUint8) {
			return false
		}
		v1 = uint8(v)
	case int32:
		if moreThan(v, math.MaxUint8) {
			return false
		}
		v1 = uint8(v)
	case int64:
		if moreThan(v, math.MaxUint8) {
			return false
		}
		v1 = uint8(v)
	default:
		return false
	}

	switch v := t2.(type) {
	case string:
		v2 = TypeStrToUint8(v)
	case uint8:
		if moreThan(v, math.MaxUint8) {
			return false
		}
		v2 = v
	case uint:
		if moreThan(v, math.MaxUint8) {
			return false
		}
		v2 = uint8(v)
	case uint16:
		if moreThan(v, math.MaxUint8) {
			return false
		}
		v2 = uint8(v)
	case uint32:
		if moreThan(v, math.MaxUint8) {
			return false
		}
		v2 = uint8(v)
	case uint64:
		if moreThan(v, math.MaxUint8) {
			return false
		}
		v2 = uint8(v)
	case int:
		if moreThan(v, math.MaxUint8) {
			return false
		}
		v2 = uint8(v)
	case int16:
		if moreThan(v, math.MaxUint8) {
			return false
		}
		v2 = uint8(v)
	case int32:
		if moreThan(v, math.MaxUint8) {
			return false
		}
		v2 = uint8(v)
	case int64:
		if moreThan(v, math.MaxUint8) {
			return false
		}
		v2 = uint8(v)
	default:
		return false
	}

	if v1 == 0 || v2 == 0 {
		return false
	}
	return v1 == v2
}

// Type2Num Convert type string to number string
// Receive “A”/“4” and “AAAA”/“6”, return “4” or “6”, if not match return ""
func Type2Num(Type string) string {
	switch Type {
	case "A", "4":
		return "4"
	case "AAAA", "6":
		return "6"
	default:
		return ""
	}
}

// Type2Str Convert type string to string
// Receive “A”/“4” and “AAAA”/“6”, return “A” or “AAAA”, if not match return ""
func Type2Str(Type string) string {
	switch Type {
	case "4", "A":
		return "A"
	case "6", "AAAA":
		return "AAAA"
	default:
		return ""
	}
}

// Type2Uint8 Convert type string to uint8
// Receive “A”/“4” and “AAAA”/“6”, return “A” or “AAAA”, if not match return ""
func Type2Uint8(Type string) uint8 {
	switch Type {
	case "A", "4":
		return A
	case "AAAA", "6":
		return AAAA
	default:
		return 0
	}
}

// IsTypeValid check if the type is valid
// "A" or "4" or "AAAA" or "6" is valid
// others is invalid
func IsTypeValid(Type string) bool {
	switch Type {
	case "A", "4", "AAAA", "6":
		return true
	default:
		return false
	}
}

func IsIpValid(ip string) bool {
	return net.ParseIP(ip) != nil
}

// ----------------------------------------------------------- //

// basicHandler is a basic handler
// if CaseA or CaseAAAA is nil, return error
// if ip is invalid, return error
// if ip is valid but can't match any type, panic
func basicHandler(ip string, caseA IpHandler, caseAAAA IpHandler) (string, error) {
	if caseA == nil {
		return "", errors.New("bad handler for ipv4")
	}
	if caseAAAA == nil {
		return "", errors.New("bad handler for ipv6") // todo add handler info in error message
	}

	switch WhichType(ip) {
	case A:
		return caseA(ip)
	case AAAA:
		return caseAAAA(ip)
	default:
		if IsIpValid(ip) {
			panic("ip is valid but can't match any type")
		}
		return "", fmt.Errorf("ip  %s is invalid", ip)
	}
}

// IpHandler is a handler
// if ip is invalid, should return "" and error
type IpHandler func(ip string) (string, error)

func (i IpHandler) Msg() string {
	return "IpHandler"
}

func private(act bool) IpHandler {
	return func(ip string) (string, error) {
		addr, err := netip.ParseAddr(ip)
		if err != nil {
			return "", err
		}

		if addr.IsPrivate() {
			if act {
				// keep private
				return ip, nil
			} else {
				// remove private
				return "", nil
			}
		} else {
			// not private
			if act {
				// remove other
				return "", nil
			} else {
				// keep other
				return ip, nil
			}
		}
	}
}

var ReservePrivateOnly IpHandler = func(ip string) (string, error) {
	r := private(true)
	return basicHandler(ip, r, r)
}

var RemovePrivate IpHandler = func(ip string) (string, error) {
	r := private(false)
	return basicHandler(ip, r, r)
}

// loopback operate on ip
// if act is false, remove loopback ip
// if act is true, keep loopback ip, remove other ip
func loopback(act bool) IpHandler {
	return func(ip string) (string, error) {
		addr := net.ParseIP(ip)
		if addr.IsLoopback() {
			if act {
				// keep loopback
				return ip, nil
			} else {
				// remove loopback
				return "", nil
			}
		} else {
			// not loopback
			if act {
				// remove other
				return "", nil
			} else {
				// keep other
				return ip, nil
			}
		}
	}
}

// ReserveLoopbackOnly reserve loopback ip, remove other ip
var ReserveLoopbackOnly IpHandler = func(ip string) (string, error) {
	r := loopback(true)
	return basicHandler(ip, r, r)
}

// RemoveLoopback remove loopback ip, keep other ip
var RemoveLoopback IpHandler = func(ip string) (string, error) {
	r := loopback(false)
	return basicHandler(ip, r, r)
}

// globalUnicast operate on ip
// if act is false, remove globalUnicast ip
// if act is true, keep globalUnicast ip, remove other ip
func globalUnicast(act bool) IpHandler {
	return func(ip string) (string, error) {
		addr := net.ParseIP(ip)
		if addr.IsGlobalUnicast() {
			if act {
				// keep globalUnicast
				return ip, nil
			} else {
				// remove globalUnicast
				return "", nil
			}
		} else {
			// not globalUnicast
			if act {
				// remove other
				return "", nil
			} else {
				// keep other
				return ip, nil
			}
		}
	}
}

// ReserveGlobalUnicastOnly reserve globalUnicast ip and remove other ip
var ReserveGlobalUnicastOnly IpHandler = func(ip string) (string, error) {
	r := globalUnicast(true)
	return basicHandler(ip, func(ip string) (string, error) {
		return "", nil // return "" and error because GlobalUnicast ip is ipv6
	}, r)
}

// RemoveGlobalUnicast remove globalUnicast ip
var RemoveGlobalUnicast IpHandler = func(ip string) (string, error) {
	r := globalUnicast(false)
	return basicHandler(ip, func(ip string) (string, error) {
		return ip, nil
	}, r)
}

// RemoveInvalid remove invalid string
var RemoveInvalid IpHandler = func(ip string) (string, error) {
	if IsIpValid(ip) {
		return ip, nil
	} else {
		return "", nil
	}
}

type selector struct{}

func (s selector) _select(no uint64) IpHandler {
	count := uint64(0)
	hasPick := false
	return func(ip string) (string, error) {
		if hasPick {
			return "", nil
		} else {
			if count == no {
				hasPick = true
				return ip, nil
			} else {
				count++
				return "", nil
			}
		}
	}
}

// NewSelector return a selector
// select the no-th ip (start from 0)
func NewSelector(no uint64) IpHandler {
	return selector{}._select(no)
}

// HandleIp ip
// handle ip with handlers
// join errors together when error handling
func HandleIp(ips []string, handlers ...IpHandler) (res []string, errs error) {
	for _, ipHandler := range handlers {
		var temp []string = nil
		for _, ip := range ips {
			After, err := ipHandler(ip)
			if err != nil {
				errs = errors.Join(errs, err)
			} else if After != "" {
				temp = append(temp, After)
			}
		}
		ips = temp
	}

	return ips, errs
}
