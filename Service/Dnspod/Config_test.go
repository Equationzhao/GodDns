/*
 *     @Copyright
 *     @file: Config_test.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/20 下午11:29
 *     @last modified: 2023/3/20 下午11:27
 *
 *
 *
 */

package Dnspod

import (
	"GodDns/DDNS"
	"gopkg.in/ini.v1"
	"os"
	"testing"
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
	Filename, err := DDNS.GetDefaultConfigurationLocation()
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

func TestSave(t *testing.T) {
	err := DDNS.SaveConfig("test.conf", os.O_CREATE|os.O_APPEND, &p)
	if err != nil {
		t.Error(err)
	}

}
