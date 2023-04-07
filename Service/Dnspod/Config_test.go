package Dnspod

import (
	"GodDns/Core"
	"strings"
	"testing"

	"gopkg.in/ini.v1"
)

func TestConfig_CreateDefaultConfig(t *testing.T) {
	c := Config{}
	config, err := c.GenerateDefaultConfigInfo()
	if err != nil {
		t.Error(err)
	}
	t.Log(config.Content)
}

func TestConfig_GenerateConfigInfo(t *testing.T) {

	config := Config{}
	info, err := config.GenerateConfigInfo(&p, 1)

	if err != nil {
		t.Error(err)
	}
	t.Log(info.Content)
}

func TestConfig_ReadConfig(t *testing.T) {
	Filename, err := Core.GetDefaultConfigurationLocation()
	if err != nil {
		t.Error(err)
	}

	cfg, err := ini.Load(Filename)
	if err != nil {
		t.Error(err)
	}

	sec, err := cfg.GetSection("Dnspod#1")
	if err != nil {
		t.Error(err)
	}

	config, err := Config{}.ReadConfig(*sec)
	if err != nil {
		t.Error(err)
	}
	t.Log(config)

}

func TestConfigFactory_Get(t *testing.T) {
	a := configFactoryInstance.Get()
	b := configFactoryInstance.Get()
	a.(*Config).test = true
	t.Log(a.(*Config).test)
	b.(*Config).test = false
	t.Log(a.(*Config).test)
	if a != b {
		t.Error("ConfigFactory.Get() is not singleton")
	}
}

func TestSplit(t *testing.T) {
	subdomain := "a,b,c, g, h, i"
	s := strings.Split(strings.ReplaceAll(subdomain, ",", " "), " ")
	for _, str := range s {
		if str == "" {
			t.Log("empty")
		} else if str == " " {
			t.Log("space")
		} else {
			t.Log(str)
		}
	}
}
