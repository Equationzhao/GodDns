// Package Service contains all DDNS service
package Service

// import order maters
import (
	_ "GodDns/service/Dnspod"       // register Dnspod
	_ "GodDns/service/DnspodYunApi" // register DnspodYunApi
)

// import _ "GodDns/Service/example"
