# Service

This package contains all the DDNS services.

```go
// in the init(), add configFactoryInstance to FactoryList
func init() {
	// add to factory list
	DDNS.Add2FactoryList(configFactoryInstance)
}


// implement DDNS.ServiceParameter or further DeviceOverridable
Parameter struct {
		Token     string `KeyValue:"Token,this tag will affect the name displayed in config, all the string after the ',' will be displayed as comments above this key"`
		Domain    string
		SubDomain string
		RecordID  string
		IpToSet   string
		Type      string // "AAAA" or "A"
		// ... other parameters
}

// implement DDNS.Request
Request struct {
		Parameter
		status DDNS.Status
		// ... any other fields
}

// implement DDNS.Config
Config struct {
}

// implement DDNS.ConfigFactory
	ConfigFactory struct {
}


```

## Structs

### ServiceParameter

This struct implements `DDNS.ServiceParameter` and `DeviceOverridable` interfaces. It contains the following fields:

- `Token` : the token used to authenticate with the DDNS provider.
- `Domain` : the domain name to update.
- `SubDomain` : the subdomain to update.
- `RecordID` : the record ID to update.
- `IpToSet` : the IP address to set.
- `Type` : the record type to update.
- `...` :any other parameters.

### Request

This struct implements `DDNS.Request` interface. It contains the following fields:

- `Parameter` (ServiceParameter): the service parameters.
- `status` (DDNS.Status): the status of the request.

### Config

This struct implements `DDNS.Config` interface. It contains the following fields:

- None

### ConfigFactory

This struct implements `DDNS.ConfigFactory` interface. It contains the following fields:

- None
