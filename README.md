# README

GoDDNS

a DDNS tool written in go

## Flags

	-R run ddns
	
		-A automatically get ip from device from Device Section
		
			-O override ip with device set from each service Section
	
		-api=ApiName get ip from ipify.org/ident.me Etc.
	
		-t=time(seconds) run ddns per time(seconds)
	
		-retry=times retry times when error occurs

	-S no output to stdout

	-G generate default configure

	-log = Trace/Debug/Info/Warn/Error log level

	-config= Config

## Configure
[Device] # required

device=[$YourDeviceName1$,$YourDeviceName2$,...]



[Name#No] # Name of the Service (start with Upper case) followed by #No

key=value # key and value of the Service (start with lower case)

...
