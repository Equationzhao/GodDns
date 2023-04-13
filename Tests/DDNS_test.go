package Tests_test

import (
	"fmt"
	"os"
	"testing"

	"GodDns/Service/Dnspod"
	"GodDns/core"
)

func TestConfigFileLocation(t *testing.T) {
	fmt.Println(core.GetConfigureLocation())
}

func TestCreateDefaultConfig(t *testing.T) {
	c := Dnspod.Config{}
	config, err := c.GenerateDefaultConfigInfo()
	if err != nil {
		t.Error(err)
	}

	location, err := core.GetDefaultConfigurationLocation()
	if err != nil {
		t.Error(err)
	}

	_, err = os.Stat(location)
	if err != nil {
		err = core.ConfigureWriter(location, os.O_CREATE, config)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestStatus_AppendMsg(t *testing.T) {
	s := &core.Status{
		Name:   "test",
		MG:     core.NewDefaultMsgGroup(),
		Status: core.Success,
	}

	s.AppendMsg(core.NewStringMsg(core.Info).AppendAssign("hello"))
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
	t.Log(core.NowVersion)
	t.Log(core.NowVersionInfo())

	latest, _, err := core.GetLatestVersionInfo()
	if err != nil {
		t.Error(err)
	}
	t.Logf("latest version: v%s", latest)

	if core.Version.Compare(latest, core.NowVersion) > 0 {
		t.Log("new version available")
	} else {
		t.Log("already latest version")
	}
}

func TestFeedback(t *testing.T) {
	str := core.Feedback()
	t.Log(str)
}
