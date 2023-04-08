// Package Device implements a Device which implements both Parameters and Config interface
// And ConfigFactory to make a Config object of Device
package Device

import (
	"bytes"
	"strings"

	"GodDns/Util"
	"GodDns/core"

	"gopkg.in/ini.v1"
)

// ServiceName is the name of Device
const ServiceName = "Device"

// ConfigInstance is a Config of Device to read/write config
var ConfigInstance Device

// ConfigFactoryInstance is a ConfigFactory to make a Config of Device
var ConfigFactoryInstance ConfigFactory

func init() {
	core.ConfigFactoryList = append(core.ConfigFactoryList, ConfigFactoryInstance)
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
func (d Device) SaveConfig(no uint) (core.ConfigStr, error) {
	return d.GenerateConfigInfo(d, no)
}

// GenerateDefaultConfigInfo generates the default config of Device
// depends on GenerateConfigInfo
// returns a ConfigStr which contains the name and content of config and nil
// should not return error
func (d Device) GenerateDefaultConfigInfo() (core.ConfigStr, error) {
	return d.GenerateConfigInfo(Device{
		Devices: []string{"interface1", "interface2", "..."},
	}, 0)
}

// ReadConfig reads the config of Device
// returns a Device which contains the config and nil
// if section [Device] has no value named "device", return nil and an error
func (d Device) ReadConfig(sec ini.Section) ([]core.Parameters, error) { // todo
	deviceList, err := sec.GetKey("device")
	if err != nil {
		return nil, err
	}

	// convert to []string
	// [DeviceName1,DeviceName2,...] -> replace "," -> [DeviceName1 DeviceName2 ...]
	// -> trim "[]" -> DeviceName1 DeviceName2 ... -> Fields " " -> []string

	// remove [] and remove " "
	d.Devices = strings.Fields(strings.Trim(strings.ReplaceAll(deviceList.String(), ",", " "), "[]"))
	return []core.Parameters{d}, nil
}

// GenerateConfigInfo generates the config of Device
// returns a ConfigStr which contains the name and content of config and nil
// should not return error
func (d Device) GenerateConfigInfo(parameters core.Parameters, no uint) (core.ConfigStr, error) {
	buffer := bytes.NewBufferString(core.ConfigHead(parameters, no))
	buffer.WriteString(Util.Convert2KeyValue(core.Format, parameters))
	buffer.Write([]byte{'\n', '\n'})
	return core.ConfigStr{Name: ServiceName, Content: buffer.String()}, nil
}

// GetName returns the ServiceName "Device"
func (d Device) GetName() string {
	return ServiceName
}

// Config returns a Config of Device
func (d Device) Config() core.Config {
	return d
}

// ConfigFactory is a factory to make a Config of Device
type ConfigFactory struct{}

// GetName returns the ServiceName "Device"
func (d ConfigFactory) GetName() string {
	return ServiceName
}

// Get single instance
func (d ConfigFactory) Get() core.Config {
	return &ConfigInstance
}

// New instance
func (d ConfigFactory) New() *core.Config {
	var res core.Config = ConfigInstance
	return &res
}
