// Package service contains all DDNS service
package service

// import order maters
import (
	_ "GodDns/service/dnspod"       // register Dnspod
	_ "GodDns/service/dnspodyunapi" // register DnspodYunApi
)

// import _ "GodDns/Service/example"
