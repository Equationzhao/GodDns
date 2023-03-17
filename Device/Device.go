/*
 *     @Copyright
 *     @file: Device.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/17 下午9:54
 *     @last modified: 2023/3/17 下午8:07
 *
 *
 *
 */

package Device

import (
	"DDNS/DDNS"
	"DDNS/Util"
	"strings"

	"gopkg.in/ini.v1"
)

const ServiceName = "Device"

var ConfigInstance Device
var ConfigFactoryInstance ConfigFactory

func init() {
	DDNS.ConfigFactoryList = append(DDNS.ConfigFactoryList, ConfigFactoryInstance)
}

type Device struct {
	Devices []string `KeyValue:"device"`
}

func (d Device) GetDevices() []string {
	return d.Devices
}

func (d Device) SaveConfig(No uint) (DDNS.ConfigStr, error) {
	return d.GenerateConfigInfo(d, No)
}

func (d Device) GenerateDefaultConfigInfo() (DDNS.ConfigStr, error) {
	return d.GenerateConfigInfo(Device{
		Devices: []string{"DeviceName"},
	}, 0)
}

func (d Device) ReadConfig(sec ini.Section) (DDNS.Parameters, error) { // todo
	deviceList, err := sec.GetKey("device")
	if err != nil {
		return nil, err
	}

	// convert to []string
	//[DeviceName1,DeviceName2,...] -> replace "," -> [DeviceName1 DeviceName2 ...] -> trim "[]" -> DeviceName1 DeviceName2 ... -> split " " -> []string
	d.Devices = strings.Split(strings.Trim(strings.Replace(deviceList.String(), ",", " ", -1), "[]"), " ") // remove [] and remove " "
	return d, nil
}

func (d Device) GenerateConfigInfo(parameters DDNS.Parameters, No uint) (DDNS.ConfigStr, error) {
	head := DDNS.ConfigHead(parameters, No)
	body := Util.Convert2KeyValue(DDNS.Format, parameters)
	tail := "\n\n"
	content := head + body + tail

	return DDNS.ConfigStr{
		Name:    ServiceName,
		Content: content,
	}, nil
}

func (d Device) GetName() string {
	return ServiceName
}

func (d Device) Config() DDNS.Config {
	return d
}

type ConfigFactory struct {
}

func (d ConfigFactory) GetName() string {
	return ServiceName
}

// Get single instance
func (d ConfigFactory) Get() DDNS.Config {
	return &ConfigInstance
}

// New instance
func (d ConfigFactory) New() *DDNS.Config {
	var res DDNS.Config = ConfigInstance
	return &res
}
