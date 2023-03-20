# GodDNS

```
   ______              __   ____     _   __        
  / ____/  ____   ____/ /  / __ \   / | / /  _____
 / / __   / __ \ / __  /  / / / /  /  |/ /  / ___/
/ /_/ /  / /_/ // /_/ /  / /_/ /  / /|  /  (__  ) 
\____/   \____/ \__,_/  /_____/  /_/ |_/  /____/  
                         .___  .___             
   ____   ____         __| _/__| _/____   ______
  / ___\ /  _ \  ___  / __ |/ __ |/    \ /  ___/
 / /_/  >  <_> ) --- / /_/ / /_/ |   |  \\___ \ 
 \___  / \____/      \____ \____ |___|  /____  |
/_____/                   \/    \/    \/     \/ 
                                               
                                                
```

![GitHub](https://img.shields.io/github/license/Equationzhao/GoDDNS) [![Codacy Badge](https://app.codacy.com/project/badge/Grade/18444501bfd44f919c3a4c87b4e8fcaf)](https://app.codacy.com/gh/Equationzhao/GoDDNS/dashboard?utm\_source=gh\&utm\_medium=referral\&utm\_content=\&utm\_campaign=Badge\_grade) [![CodeFactor](https://www.codefactor.io/repository/github/equationzhao/goddns/badge)](https://www.codefactor.io/repository/github/equationzhao/goddns) 

![GitHub last commit](https://img.shields.io/github/last-commit/Equationzhao/GoDDNS) ![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/Equationzhao/GoDDNS) [![Go](https://github.com/Equationzhao/GodDns/actions/workflows/go.yml/badge.svg)](https://github.com/Equationzhao/GodDns/actions/workflows/go.yml)

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FEquationzhao%2FGoDDNS.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FEquationzhao%2FGoDDNS?ref=badge\_large)

a DDNS tool written in go

## usage
```bash
USAGE:
   GodDns [global options] command [command options] [arguments...]
   GodDns run - run the DDNS service
   GodDns run auto - run ddns, use ip address of interface set in Device Section automatically
   GodDns run auto override - run ddns, override the ip address of interface set in each service Section
   GodDns generate - generate a default configuration file
```

```
COMMANDS:
   run, r, R       run the DDNS service 
   [--api ApiName, -i ApiName, -I ApiName  get ip address from provided ApiName, eg: ipify/identMe]
	   auto, a, A  run ddns, use ip address of interface set in Device Section automatically
   			override, o, O  run ddns, override the ip address of interface set in each service Section
OPTIONS:
   --time seconds                         run ddns per time(seconds) (default: 0)
   --retry times                          retry times (default: 3)
   --silent, -s, -S                       no message output (default: false)
   --log value, -l value, -L value        Trace/Debug/Info/Warn/Error (default: "info")
   --config file, -c file, -C file        set configuration file
   --help, -h, -H                         show help (default: false)

COMMANDS:
   generate, g, G  generate a default configuration file
   help, h         Shows a list of commands or help for one command

OPTIONS:
   --silent, -s, -S                 no message output (default: false)
   --log value, -l value, -L value  Trace/Debug/Info/Warn/Error
   --config file, -c file, -C file  set configuration file
   --help, -h, -H                   show help (default: false)

GLOBAL OPTIONS:
   --help, -h, -H     show help (default: false)
   --version, -v, -V  print the version info (default: false)

```

## Configuration

\[Device] # required

device=\[$YourDeviceName1$,$YourDeviceName2$,...]

\[Name#No] # Name of the Service (start with Upper case) followed by #No

key=value # key and value of the Service (start with lower case)

...

## TODO

* [ ] add more service
* [ ] check flag validation
* [x] add support to write comment to configuration
* [ ] to fix RunPerTime at main.go:664
* [ ] refactor see DDNS.Config.go:211
* [ ] todo replace net with netip
* [ ] todo refactor do not use hard code "Devices" at DDNS.Config:211
* [ ] new feature support multi-device for each service(like Device does)
* [ ] ? refactor Dnspod.Config.ReadConfig:62
* [ ] deal ips at Net.Ip:375

## ISSUES

1. may be bug when deleting element in loop, see main.go:408
2. error saving config in linux

## [![Repography logo](https://images.repography.com/logo.svg)](https://repography.com) Structure

[![Structure](https://images.repography.com/35290882/Equationzhao/GoDDNS/structure/Xvtsc2MXHRRRBOO98rPykluHsbjgiXVtv151YJjZe-g/eV5f7dIVTtGDBh-UK4EnRsrCo0rHTumqrtoK3Ih6Ap0\_table.svg)](https://github.com/Equationzhao/GoDDNS)
