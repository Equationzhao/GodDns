/*
 *
 *     @file: IP_test.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/28 下午3:58
 *     @last modified: 2023/3/26 下午11:18
 *
 *
 *
 */

package Net

import (
	"fmt"
	"net/netip"
	"testing"
)

func TestGetClientIp(t *testing.T) {
	// GetIP()
	s, err := GetIp("WLAN")
	if err != nil {
		t.Error(err)
	}

	t.Log(s)
}

func TestGetClientIpByType(t *testing.T) {
	s, err := GetIpByType("WLAN", A)
	if err != nil {
		t.Error(err)
	}
	t.Log(s)
}

func TestWhichType(t *testing.T) {

	ip := "192.168.1"
	t.Log(WhichType(ip))

	ip = "127.0.0.1"
	t.Log(WhichType(ip))

	ip = "::1"
	t.Log(WhichType(ip))

	ip = "0:0:0:0:0:0:0:0"
	t.Log(WhichType(ip))

	ip = "1234:1234:1234:1234:1234:1234:1234:1234"
	t.Log(WhichType(ip))
}

func TestGetIPFromIpify(t *testing.T) {
	s, err := getIPFromIpify(4)
	if err != nil {
		t.Error(err)
	}
	t.Log(s)

	s, err = getIPFromIpify(6)
	if err != nil {
		t.Error(err)
	}
	t.Log(s)
}

func TestGetIPFromIdentMe(t *testing.T) {
	s, err := getIPFromIdentMe(4)
	if err != nil {
		t.Error(err)
	}
	t.Log(s)

	s, err = getIPFromIdentMe(6)
	if err != nil {
		t.Error(err)
	}
	t.Log(s)
}

func TestTypeEqual(t *testing.T) {

	t.Log(TypeEqual(256, 256))

	t.Log(TypeEqual("A", "A"))
	t.Log(TypeEqual("A", A))
	t.Log(TypeEqual("A", 4))

	t.Log(TypeEqual(A, A))
	t.Log(TypeEqual(A, 4))
	t.Log(TypeEqual(A, "A"))

	t.Log(TypeEqual("AAAA", AAAA))
	t.Log(TypeEqual("AAAA", 6))
	t.Log(TypeEqual("AAAA", "AAAA"))

	t.Log(TypeEqual(AAAA, AAAA))
	t.Log(TypeEqual(AAAA, 6))
	t.Log(TypeEqual(AAAA, "AAAA"))

	t.Log(TypeEqual(4, 4))
	t.Log(TypeEqual(4, A))
	t.Log(TypeEqual(4, "A"))

	t.Log(TypeEqual(6, 6))
	t.Log(TypeEqual(6, AAAA))
	t.Log(TypeEqual(6, "AAAA"))

}

func TestAdd2APIMap(t *testing.T) {
	t.Log(ApiMap)
	ApiMap.Add2Apis("Test", Api{Get: func(u uint8) (string, error) {
		return "Test", nil
	}})
	t.Log(ApiMap)
	s, err := ApiMap.GetApi("ipify")
	if err != nil {
		t.Error(err)
	}
	ip, err := s.Get(4)

	if err != nil {
		t.Error(err)
	}
	t.Log(ip)
}

func TestHandler(t *testing.T) {
	ips := []string{"1.1.1.1", "2.2.2.2", "fe80::1111:2222:3333:4444", "fe80::1111:2222:3333:4445"}
	ips, _ = HandleIp(ips, func(ip string) (string, error) {
		if WhichType(ip) == A {
			return ip, nil
		}
		return "", nil
	})

	fmt.Println(ips)

	ips = []string{"127.0.0.1", "::1", "8.8.8.8"}
	ips, _ = HandleIp(ips, RemoveLoopback)
	fmt.Println(ips)

	ns3 := NewSelector(3)
	ips = []string{"127.0.0.1", "::1", "8.8.8.8"}
	ips, _ = HandleIp(ips, ns3)
	fmt.Println(ips)

	ips = []string{"127.0.0.1", "::1", "8.8.8.8"}
	ips, _ = HandleIp(ips, ns3)
	fmt.Println(ips)

	ips = []string{"127.0.0.1", "::1", "8.8.8.8", "2001:db8::68", "1.2.3.invalid"}
	ips, _ = HandleIp(ips, RemoveInvalid, ReserveGlobalUnicastOnly)
	fmt.Println(ips)

	ips = []string{"127.0.0.1", "::1", "8.8.8.8", "2001:db8::68", "1.2.3.invalid"}
	ips, _ = HandleIp(ips, RemoveInvalid, RemoveGlobalUnicast, ReserveGlobalUnicastOnly)
	fmt.Println(ips)

	ips = []string{"127.0.0.1", "::1", "8.8.8.8", "2001:db8::68", "1.2.3.invalid"}
	ips, _ = HandleIp(ips, RemoveInvalid, RemoveLoopback, RemovePrivate)
	fmt.Println(ips)
}

func isIpValid2(ip string) bool {
	if _, err := netip.ParseAddr(ip); err != nil {
		return false
	}
	return true
}

func BenchmarkIsIpValid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := IsIpValid("123.432.123.456")
		_ = b
	}
}

func BenchmarkIsIpValid2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := isIpValid2("123.432.123.456")
		_ = b
	}
}
