/*
 *     @Copyright
 *     @file: IP.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/20 下午11:29
 *     @last modified: 2023/3/20 下午11:27
 *
 *
 *
 */

package Net

import (
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/rdegges/go-ipify"
	"net" // todo replace with net ip
	"net/netip"
)

const (
	A              uint8 = 4
	AAAA           uint8 = 6
	DefaultTypeStr       = "A"
	DefaultType          = A
)

// Api is a type = function that return a string and an error
type Api = func(uint8) (string, error)

// Apis contains a map of apis
type Apis struct {
	a map[string]func(uint8) (string, error)
}

// ApiMap is a default Apis, contains a map of apis
var ApiMap = Apis{
	a: map[string]func(uint8) (string, error){
		"ipify":   getIPFromIpify,
		"identMe": getIPFromIdentMe,
	},
}

// GetApiName return the names of apis
func (a *Apis) GetApiName() []string {
	var res = make([]string, 0, len(a.a))
	for s := range a.a {
		res = append(res, s)
	}
	return res
}

func CreateApiFromURL(URL string) {

}

// Add2Apis add api to Map
func (a *Apis) Add2Apis(name string, f Api) {
	a.a[name] = f
}

// GetApi return the api function
func (a *Apis) GetApi(name string) (Api, error) {
	api, ok := a.a[name]
	if !ok {
		return nil, errors.New("not found")
	}
	return api, nil
}

// GetMap return the map of apis
func (a *Apis) GetMap() map[string]Api {
	return a.a
}

// GetIp return Ip list of corresponding interface and nil error when error occurs, return nil and error
func GetIp(NameToMatch string) ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var ips []string

	for _, i := range interfaces {

		if i.Name == NameToMatch {
			address, err := i.Addrs()
			if err != nil {
				return nil, err
			}
			for _, addr := range address {
				// ? how
				ips = append(ips, addr.(*net.IPNet).IP.String())
				// switch v := addr.(type) {
				// case *net.IPNet:
				//	ips = append(ips, v.IP.String())
				// }
			}
			return ips, nil
		}
	}
	return nil, fmt.Errorf("not found")
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

// GetIpByType get specific type ip of corresponding interface
// parameter: NameToMatch, Type
// NameToMatch: interface name
// Type: 4 for ipv4, 6 for ipv6 (use constant A and AAAA)
// return: ip list, error
// when error occurs, return nil and error
func GetIpByType(NameToMatch string, Type uint8) ([]string, error) {
	if Type != A && Type != AAAA {
		return nil, fmt.Errorf("invalid type: %d", Type)
	} else {
		ips, err := GetIp(NameToMatch)
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

// getIPFromIpify get ip from ipify
func getIPFromIpify(Type uint8) (string, error) {

	switch Type {
	case 6:
		ipify.API_URI = "https://api6.ipify.org"
	case 4:
		ipify.API_URI = "https://api.ipify.org"
	default:
		return "", fmt.Errorf("invalid type: %d", Type)
	}
	ip, err := ipify.GetIp()
	if err != nil {
		return "", err
	}
	return ip, nil
}

// getIPFromIdentMe get ip from ident.me
func getIPFromIdentMe(Type uint8) (string, error) {
	ApiUri := ""

	switch Type {
	case 6:
		ApiUri = "https://v6.ident.me"
	case 4:
		ApiUri = "https://v4.ident.me"
	default:
		return "", fmt.Errorf("invalid type: %d", Type)
	}
	r := resty.New()
	res, err := r.R().Get(ApiUri)
	if err != nil || res.String() == "" {
		return "", err
	}

	return res.String(), nil
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
		if moreThan(v, 255) {
			return false
		}
		v1 = v
	case uint:
		if moreThan(v, 255) {
			return false
		}
		v1 = uint8(v)
	case uint16:
		if moreThan(v, 255) {
			return false
		}
		v1 = uint8(v)
	case uint32:
		if moreThan(v, 255) {
			return false
		}
		v1 = uint8(v)
	case uint64:
		if moreThan(v, 255) {
			return false
		}
		v1 = uint8(v)
	case int:
		if moreThan(v, 255) {
			return false
		}
		v1 = uint8(v)
	case int16:
		if moreThan(v, 255) {
			return false
		}
		v1 = uint8(v)
	case int32:
		if moreThan(v, 255) {
			return false
		}
		v1 = uint8(v)
	case int64:
		if moreThan(v, 255) {
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
		if moreThan(v, 255) {
			return false
		}
		v2 = v
	case uint:
		if moreThan(v, 255) {
			return false
		}
		v2 = uint8(v)
	case uint16:
		if moreThan(v, 255) {
			return false
		}
		v2 = uint8(v)
	case uint32:
		if moreThan(v, 255) {
			return false
		}
		v2 = uint8(v)
	case uint64:
		if moreThan(v, 255) {
			return false
		}
		v2 = uint8(v)
	case int:
		if moreThan(v, 255) {
			return false
		}
		v2 = uint8(v)
	case int16:
		if moreThan(v, 255) {
			return false
		}
		v2 = uint8(v)
	case int32:
		if moreThan(v, 255) {
			return false
		}
		v2 = uint8(v)
	case int64:
		if moreThan(v, 255) {
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

// DealWithIp deal with ip
// like get specific ip ?
func DealWithIp(ip ...string) string {
	// todo deal with ip like getting specific ip
	return ip[0]
}
