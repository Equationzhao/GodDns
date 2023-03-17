# README

GoDDNS 

![GitHub](https://img.shields.io/github/license/Equationzhao/GoDDNS) [![Codacy Badge](https://app.codacy.com/project/badge/Grade/18444501bfd44f919c3a4c87b4e8fcaf)](https://app.codacy.com/gh/Equationzhao/GoDDNS/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade) ![GitHub last commit](https://img.shields.io/github/last-commit/Equationzhao/GoDDNS) ![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/Equationzhao/GoDDNS)

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FEquationzhao%2FGoDDNS.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FEquationzhao%2FGoDDNS?ref=badge_large)

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



## [![Repography logo](https://images.repography.com/logo.svg)](https://repography.com)  Structure
[![Structure](https://images.repography.com/35290882/Equationzhao/GoDDNS/structure/Xvtsc2MXHRRRBOO98rPykluHsbjgiXVtv151YJjZe-g/eV5f7dIVTtGDBh-UK4EnRsrCo0rHTumqrtoK3Ih6Ap0_table.svg)](https://github.com/Equationzhao/GoDDNS)

