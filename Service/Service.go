/*
 *     @Copyright
 *     @file: Service.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/22 上午6:29
 *     @last modified: 2023/3/22 上午6:21
 *
 *
 *
 */

// Package Service contains all DDNS service
package Service

// import order maters
import _ "GodDns/Service/Dnspod"       // register Dnspod
import _ "GodDns/Service/DnspodYunApi" // register Cloudflare
