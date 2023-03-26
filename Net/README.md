# Net

## Description

This is a collection of network related tools.

## Tools

get ip(string) of interfaces
```go
ips, _ := Net.GetIP("eth0")
fmt.Println(ips)
output:
[fe80::1111:2222:3333:4444 1.1.1.1]
```

get the type of ip
```go
ipv4 := "1.1.1.1"
ipv6 := "fe80::1111:2222:3333:4444"
fmt.Println(Net.Which(ipv4),", ",Net.Which(ipv6))
output:
4
```

convert type string to number string 
```go
fmt.Println(Net.Type2Num("4"))
fmt.Println(Net.Type2Num("6"))
fmt.Println(Net.Type2Num("A"))
fmt.Println(Net.Type2Num("AAAA"))
fmt.Println(Net.Type2Num("5"))
fmt.Println(Net.Type2Num("AAA"))
output:
4
6
4
6
       // empty string
	   // 

```

convert number string type to string type
```go
fmt.Println(Net.Num2Type("4"))
fmt.Println(Net.Num2Type("6"))
fmt.Println(Net.Num2Type("A"))
fmt.Println(Net.Num2Type("AAAA"))
fmt.Println(Net.Num2Type("5"))
fmt.Println(Net.Num2Type("AAA"))
output:
A
AAAA
A
AAAA
       // empty string
       // 

```
convert string type to uint8 type
```go
fmt.Println(Net.Type2Uint8("4"))
fmt.Println(Net.Type2Uint8("6"))
fmt.Println(Net.Type2Uint8("A"))
fmt.Println(Net.Type2Uint8("AAAA"))
fmt.Println(Net.Type2Uint8("5"))
fmt.Println(Net.Type2Uint8("AAA"))
output:
4
6
4
6
0
0

```

## ip Handler

Handle ip slice, apply handler to each ip element
```go
ips := []string{"1.1.1.1","2.2.2.2","fe80::1111:2222:3333:4444","fe80::1111:2222:3333:4445"}
ips , _ = Net.HandleIP(ips, func(ip string)(string, error) {
    fmt.Println(ip)
    return ip, nil
})
fmt.Println(ips)
output:
1.1.1.1
2.2.2.2
fe80::1111:2222:3333:4444
fe80::1111:2222:3333:4445
```

Work as Filter
```go
ips := []string{"1.1.1.1", "2.2.2.2", "fe80::1111:2222:3333:4444", "fe80::1111:2222:3333:4445"}
ips, _ = Net.HandleIp(ips, func(ip string) (string, error) {
	if WhichType(ip) == Net.A {
		return ip, nil
	}
	return "", nil
})

fmt.Println(ips)
output:
[1.1.1.1 2.2.2.2]
```

Built-in Handler:
RemoveLoopback ReserveLoopbackOnly
RemoveGlobalUnicast ReserveGlobalUnicastOnly

// it's not recommended to use this, just to use ip[no] instead,
NewSelector(no uint64) if you want to select the second ip, you can use Net.NewSelector(1) // start from 0

```go
ips := []string{"127.0.0.1", "::1", "8.8.8.8"}
ips , _ = Net.HandleIp(ips, Net.RemoveLoopback)
fmt.Println(ips)
output:
[8.8.8.8]
```
handlers will be applied in order
```go
ips := []string{"127.0.0.1", "::1", "8.8.8.8"}
ips, _ = Net.HandleIp(ips, Net.RemoveGlobalUnicast, Net.ReserveGlobalUnicastOnly)
fmt.Println(ips)
output:
[] // no match
```
