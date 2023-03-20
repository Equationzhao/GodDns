/*
 *     @Copyright
 *     @file: DDNS_test.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/20 下午11:29
 *     @last modified: 2023/3/20 下午11:27
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

	location, err := DDNS.GetDefaultConfigurationLocation()
	if err != nil {
		t.Error(err)
	}

	err = DDNS.ConfigureWriter(location, os.O_CREATE, config)
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

func TestVersion(t *testing.T) {
	t.Log(DDNS.NowVersion)
	t.Log(DDNS.NowVersionInfo())

	latest, _, err := DDNS.GetLatestVersionInfo()
	if err != nil {
		t.Error(err)
	}
	t.Logf("latest version: v%s", latest)

	if DDNS.Version.Compare(latest, DDNS.NowVersion) > 0 {
		t.Log("new version available")
	} else {
		t.Log("already latest version")
	}

}

func TestFeedback(t *testing.T) {
	str := DDNS.Feedback()
	t.Log(str)
}
