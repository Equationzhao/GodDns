// Package Device implements a Device which implements both Parameters and Config interface
// And ConfigFactory to make a Config object of Device
package Device

import (
	"GodDns/DDNS"
	"GodDns/Util"
	"bytes"
	"strings"

	"gopkg.in/ini.v1"
)

// ServiceName is the name of Device
const ServiceName = "Device"

// ConfigInstance is a Config of Device to read/write config
var ConfigInstance Device

// ConfigFactoryInstance is a ConfigFactory to make a Config of Device
var ConfigFactoryInstance ConfigFactory

func init() {
	DDNS.ConfigFactoryList = append(DDNS.ConfigFactoryList, ConfigFactoryInstance)
}

// Device contains a slice of device
// implements Parameters and Config interface
type Device struct {
	Devices []string `KeyValue:"device"`
}

// GetDevices returns the slice of device
func (d Device) GetDevices() []string {
	return d.Devices
}

// SaveConfig saves the config of Device
// returns a ConfigStr which contains the name and content of config and nil
// should not return error
func (d Device) SaveConfig(No uint) (DDNS.ConfigStr, error) {
	return d.GenerateConfigInfo(d, No)
}

// GenerateDefaultConfigInfo generates the default config of Device
// depends on GenerateConfigInfo
// returns a ConfigStr which contains the name and content of config and nil
// should not return error
func (d Device) GenerateDefaultConfigInfo() (DDNS.ConfigStr, error) {
	return d.GenerateConfigInfo(Device{
		Devices: []string{"interface1", "interface2", "..."},
	}, 0)
}

// ReadConfig reads the config of Device
// returns a Device which contains the config and nil
// if section [Device] has no value named "device", return nil and an error
func (d Device) ReadConfig(sec ini.Section) ([]DDNS.Parameters, error) { // todo
	deviceList, err := sec.GetKey("device")
	if err != nil {
		return nil, err
	}

	// convert to []string
	// [DeviceName1,DeviceName2,...] -> replace "," -> [DeviceName1 DeviceName2 ...] -> trim "[]" -> DeviceName1 DeviceName2 ... -> Fields " " -> []string
	d.Devices = strings.Fields(strings.Trim(strings.ReplaceAll(deviceList.String(), ",", " "), "[]")) // remove [] and remove " "

	ps := []DDNS.Parameters{d}
	return ps, nil
}

// GenerateConfigInfo generates the config of Device
// returns a ConfigStr which contains the name and content of config and nil
// should not return error
func (d Device) GenerateConfigInfo(parameters DDNS.Parameters, No uint) (DDNS.ConfigStr, error) {
	buffer := bytes.NewBufferString(DDNS.ConfigHead(parameters, No))
	buffer.WriteString(Util.Convert2KeyValue(DDNS.Format, parameters))
	buffer.Write([]byte{'\n', '\n'})
	return DDNS.ConfigStr{
		Name:    ServiceName,
		Content: buffer.String(),
	}, nil
}

// GetName returns the ServiceName "Device"
func (d Device) GetName() string {
	return ServiceName
}

// Config returns a Config of Device
func (d Device) Config() DDNS.Config {
	return d
}

// ConfigFactory is a factory to make a Config of Device
type ConfigFactory struct {
}

// GetName returns the ServiceName "Device"
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
