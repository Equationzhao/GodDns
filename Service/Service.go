// Package Service contains all DDNS service
package Service

// import order maters
import (
	_ "GodDns/Service/Dnspod"       // register Dnspod
	_ "GodDns/Service/DnspodYunApi" // register DnspodYunApi
)

// import _ "GodDns/Service/example"
