package Tests_test

import (
	"GodDns/Core"
	"GodDns/Service/Dnspod"
	"fmt"
	"os"
	"testing"
)

func TestConfigFileLocation(t *testing.T) {
	fmt.Println(Core.GetConfigureLocation())
}

func TestCreateDefaultConfig(t *testing.T) {
	c := Dnspod.Config{}
	config, err := c.GenerateDefaultConfigInfo()
	if err != nil {
		t.Error(err)
	}

	location, err := Core.GetDefaultConfigurationLocation()
	if err != nil {
		t.Error(err)
	}

	_, err = os.Stat(location)
	if err != nil {
		err = Core.ConfigureWriter(location, os.O_CREATE, config)
		if err != nil {
			t.Error(err)
		}
	}

}

func TestStatus_AppendMsg(t *testing.T) {
	s := &Core.Status{
		Name:   "test",
		MG:     Core.NewDefaultMsgGroup(),
		Status: Core.Success,
	}

	s.AppendMsg(Core.NewStringMsg(Core.Info).AppendAssign("hello"))
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
	t.Log(Core.NowVersion)
	t.Log(Core.NowVersionInfo())

	latest, _, err := Core.GetLatestVersionInfo()
	if err != nil {
		t.Error(err)
	}
	t.Logf("latest version: v%s", latest)

	if Core.Version.Compare(latest, Core.NowVersion) > 0 {
		t.Log("new version available")
	} else {
		t.Log("already latest version")
	}

}

func TestFeedback(t *testing.T) {
	str := Core.Feedback()
	t.Log(str)
}
