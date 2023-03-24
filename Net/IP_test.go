/*
 *     @Copyright
 *     @file: IP_test.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/25 上午1:46
 *     @last modified: 2023/3/25 上午1:44
 *
 *
 *
 */

package Net

import (
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
