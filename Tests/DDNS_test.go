package Tests_test

import (
	"GodDns/Core"
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

	_, err = os.Stat(location)
	if err != nil {
		err = DDNS.ConfigureWriter(location, os.O_CREATE, config)
		if err != nil {
			t.Error(err)
		}
	}

}

func TestStatus_AppendMsg(t *testing.T) {
	s := &DDNS.Status{
		Name:   "test",
		MG:     DDNS.NewDefaultMsgGroup(),
		Status: DDNS.Success,
	}

	s.AppendMsg(DDNS.NewStringMsg(DDNS.Info).AppendAssign("hello"))
	s.MG.AddInfo("world")

	if s.MG.GetInfo()[0].String() != "hello" {
		t.Error("AppendMsg failed")
	}

	if s.MG.GetInfo()[1].String() != "world" {
		t.Error("AppendMsg failed")
	}
	t.Log(s.MG)
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
