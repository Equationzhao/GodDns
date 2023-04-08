// Package example is a template for creating new service
package example

import "GodDns/core"

func init() {
	// add to factory list
	core.Add2FactoryList(configFactoryInstance)
}

const serviceName = "example"

var (
	configFactoryInstance ConfigFactory
	configInstance        Config
)

type (

	// Parameter should implement DDNS.Service at least
	// and implement DDNS.DeviceOverridable to support user-defined Net Interface name in this Service Section in config
	Parameter struct {
		core.DefaultMsgGroup `KeyValue:"-"`
		Token                string `KeyValue:"Token,this tag will affect the name displayed in config, all the string after the ',' will be displayed as comments above this key"`
		Domain               string
		SubDomain            string
		RecordID             string
		IpToSet              string
		Type                 string // "AAAA" or "A"
		// ... other parameters
	}

	// Request should implement DDNS.Request
	Request struct {
		Parameter
		status core.Status
		// ... any other fields
	}

	// Config should implement DDNS.Config
	Config struct{}

	// ConfigFactory should implement DDNS.ConfigFactory
	ConfigFactory struct{}
)
