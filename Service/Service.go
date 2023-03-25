/*
 *     @Copyright
 *     @file: Service.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/26 上午4:04
 *     @last modified: 2023/3/26 上午4:02
 *
 *
 *
 */

// Package Service contains all DDNS service
package Service

// import order maters
import _ "GodDns/Service/Dnspod"       // register Dnspod
import _ "GodDns/Service/DnspodYunApi" // register DnspodYunApi

// import _ "GodDns/Service/example"
