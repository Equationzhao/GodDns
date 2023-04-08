// Package Core
// basic interfaces and tools for DDNS service
package Core

import (
	"fmt"
)

// Parameters basic interface
type Parameters interface {
	GetName() string                       // return like "dnspod"
	SaveConfig(No uint) (ConfigStr, error) // todo LoadOptions with comment/without comment(default)/...
}

// Service is an interface a service must implement
type Service interface {
	Parameters
	Target() string
	ToRequest() (Request, error)
	SetValue(string)
	GetIP() string
	GetType() string // return "4" or "6" and "" if invalid type
	IsTypeSet() bool // IsTypeSet return true if the type is set correctly
}

// DeviceOverridable is an interface for service whose Ip value can be overridden by the specific Type ip of a device
type DeviceOverridable interface {
	Service
	// GetDevice return the device name
	GetDevice() string // todo change to GetDeviceList and return []string
	// IsDeviceSet return true if the device is set
	IsDeviceSet() bool
}

// Find finds the first parameter in the slice of parameters that has the same name as toFind.
// If the parameter is found, Find returns the parameter and nil error.
// If the parameter is not found, Find returns nil and an error.
func Find(parameters []Parameters, toFind string) (Parameters, error) {
	for _, d := range parameters {
		if d.GetName() == toFind {
			return d, nil
		}

	}
	return nil, fmt.Errorf("%s not found", toFind)
}
