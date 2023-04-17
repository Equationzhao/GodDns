# DDNS
Program config
## Config example
```ini
[settings]
Proxy = [socks://localhost:10808 http://127.0.0.1:10809]
ocst = 10s


# when Response=TEXT, Value is the no-th ip in the response
# 1.1.1.1 2.2.2.2 3.3.3.3
# Value=2 -> 2.2.2.2

[api.MyApi1] # ApiName is MyApi1
A=https://ip.3322.net
AAAA=http://myip.ipip.net/s
Response=TEXT
HTTPMethod=GET
Value=0




# when Response=JSON, Value is the key of the value you want to get
# {
#   "content":[
#       {
#           "address":".....",
#           "ip":"xxxxx"
#       }
#   ] 
# }
# Value=content[0].ip

[api.MyApi2] # ApiName is MyApi2
A=https://api.ipify.org?format=json
AAAA=https://api6.ipify.org?format=json
Response=JSON
HTTPMethod=GET
Value=ip
```


