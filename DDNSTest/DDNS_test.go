/*
 *     @Copyright
 *     @file: DDNS_test.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/18 下午3:52
 *     @last modified: 2023/3/18 下午3:52
 *
 *
 *
 */

package DDNS_test

import (
	"GodDns/DDNS"
	"GodDns/Service/Dnspod"
	"fmt"
	"os"
	"testing"
)

func TestConfigFileLocation(t *testing.T) {
	fmt.Println(DDNS.GetConfigureLocation())
}

func TestCreateDefaultConfig(t *testing.T) {
	c := Dnspod.Config{}
	config, err := c.GenerateDefaultConfigInfo()
	if err != nil {
		t.Error(err)
	}
	err = DDNS.ConfigureWriter(DDNS.GetConfigureLocation(), os.O_CREATE, config)
	if err != nil {
		t.Error(err)
	}
}

func TestStatus_AppendMsg(t *testing.T) {
	s := &DDNS.Status{
		Name:    "test",
		Msg:     "hello",
		Success: DDNS.Success,
	}

	s2 := &DDNS.Status{
		Name:    "test",
		Msg:     "!",
		Success: DDNS.Success,
	}

	s.AppendMsg(" ", "world", s2.Msg)

	if s.Msg != "hello world!" {
		t.Error("AppendMsg failed")
	}
	t.Log(s.Msg)
}

func TestStatus_AppendMsgF(t *testing.T) {
	s := &DDNS.Status{
		Name:    "test",
		Msg:     "hello",
		Success: DDNS.Success,
	}

	s2 := &DDNS.Status{
		Name:    "test",
		Msg:     "!",
		Success: DDNS.Success,
	}
	s.AppendMsgF(" %s%s", "world", s2.Msg)

	if s.Msg != "hello world!" {
		t.Error("AppendMsg failed")
	}
	t.Log(s.Msg)
}
