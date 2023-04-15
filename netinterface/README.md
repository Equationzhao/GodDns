# Device (net interface)

## Description

Read device(net interface) setting from config

```go
// Device contains a slice of device
// implements Parameters and Config interface
type Device struct {
	Devices []string `KeyValue:"device"`
}
```

## Config

```ini
[Device]
device=eth0 or device=eth0,eth1 or device=[eth0,eth1] or device=[eth0 eth1] # space and comma are both ok
```
