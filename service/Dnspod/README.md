# Dnspod

## Steps

1. Read config file
2. Make request to Get record_id
3. Make request

## Config

```ini
[Dnspod#1]
 # get from https://console.dnspod.cn/account/token/token, 'ID,Token'
login_token=TOKEN
 # data format, json(recommended) or xml(not support yet)
format=json
 # language, en or zh(recommended)
lang=en
 # return error if the data doesn't exist,no(recommended) or yes
error_on_empty=no
 # domain name
domain=example.com
 # record id can be get by making http POST request with required Parameters to https://dnsapi.cn/Record.List, more at https://docs.dnspod.com/api/get-record-list/
record_id=0
 # record name like www., if you have multiple records to update, set like sub_domain=www,ftp,mail
sub_domain=sub
 # The record line.You can get the list from the API.The default value is '默认'
record_line=默认
 # IP address like 6.6.6.6
value=YOUR IP
 # Time-To-Live, 600(default)
ttl=600
 # A/AAAA/4/6
type=AAAA
```
