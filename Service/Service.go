/*
 *
 *     @file: Service.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/30 下午11:29
 *     @last modified: 2023/3/30 下午3:37
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
