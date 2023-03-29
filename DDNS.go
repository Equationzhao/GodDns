/*
 *
 *     @file: DDNS.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/29 下午11:24
 *     @last modified: 2023/3/29 下午8:47
 *
 *
 *
 */

/*
 *
 *     @file: DDNS.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/28 下午3:59
 *     @last modified: 2023/3/28 下午3:59
 *
 *
 *
 */

package main

import (
	"GodDns/DDNS"
	"GodDns/Device"
	log "GodDns/Log"
	"GodDns/Net"
	"GodDns/Util"
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	"os"
	"strconv"
	"time"
)

// -----------------------------------------------------------------------------------------------------------------------------------------//

func RunDDNS(parameters []DDNS.Parameters) error {
	log.Debugf("run ddns")
	// run ddns here

	// get from api
	if ApiName != "" {
		return RunGetFromApi(parameters)
	}

	// -A is not set
	requests := GenerateRequests(parameters)

	if Time != 0 {
		RunPerTime(Time, requests)
	}

	d, err := DDNS.Find(parameters, Device.ServiceName)
	Parameters2Save := make([]DDNS.Parameters, 0, len(parameters))
	if err == nil {
		Parameters2Save = append(Parameters2Save, d)
	}

	ExecuteRequests(requests...)
	for _, request := range requests {
		// update info from request.parameters
		Parameters2Save = append(Parameters2Save, request.ToParameters())
	}

	return SaveFromParameters(Parameters2Save...)
}

func ReadConfig(configs []DDNS.ConfigFactory) ([]DDNS.Parameters, error) {
	parameters, fileErr, configErrs := DDNS.ConfigureReader(DDNS.GetConfigureLocation(), configs...)
	if fileErr != nil {
		log.Errorf("error reading config: %s", fileErr.Error())

		return nil, fmt.Errorf("error reading config: %w", fileErr)
	}

	if configErrs != nil {
		log.Errorf("error reading config: %s", configErrs.Error())
	}

	if len(parameters) < 1 {
		log.Info("no service left to run")
		return nil, errors.New("no service left to run")
	}
	return parameters, nil
}

// RunGetFromApi get ip address from api
// require parameters contain Device.Device
func RunGetFromApi(parameters []DDNS.Parameters) error {

	// if api is api name , get api from map
	// if api is url , try to make request to url  http://example.com/api?ip=4 or http://example.com/api?ip=6

	var api Net.Api
	api, err := Net.ApiMap.GetApi(ApiName)
	if err != nil {
		log.Errorf("error getting api %s, %s", ApiName, err)
		// todo suggestion "do you mean xxx"
		return errors.New("") // return error with no message to avoid print error message again
	}

	log.Debugf("-I is set, get ip address from %s", ApiName)

	ip4Done := make(chan bool, 1)
	ip6Done := make(chan bool, 1)
	ip4 := ""
	_ip4 := ""
	ip6 := ""
	_ip6 := ""

	var err1 error
	go func() {
		_ip4, err1 = api.Get(4)
		if err1 != nil {
			ip4Done <- false
		} else {
			ip4 = _ip4
			ip4Done <- true
		}
		close(ip4Done)
	}()

	var err2 error
	go func() {

		_ip6, err2 = api.Get(6)
		if err2 != nil {

			ip6Done <- false
		} else {
			ip6 = _ip6

			ip6Done <- true
		}
		close(ip6Done)
	}()

	select {
	case temp := <-ip4Done:
		if temp {
			log.Infof("ipv4 from %s: %s", ApiName, ip4)
		} else {
			log.Errorf("error getting ipv4, %s", err1)
			return errors.New("quit")
		}

	case temp := <-ip6Done:
		if temp {
			log.Infof("ipv6 from %s: %s", ApiName, ip6)
		} else {
			log.Errorf("error getting ipv6, %s", err2)
			return errors.New("quit")
		}

	case <-time.After(10 * time.Second):
		log.Errorf("timeout getting ip address from %s", ApiName)
		return errors.New("quit")
	}

	for _, parameter := range parameters {
		if parameter.GetName() != Device.ServiceName {
			if d, ok := parameter.(DDNS.ServiceParameters); ok {
				if d.IsTypeSet() {
					if Net.TypeEqual(d.GetType(), Net.A) {
						d.SetValue(ip4)
					} else if Net.TypeEqual(d.GetType(), Net.AAAA) {
						d.SetValue(ip6)
					} else {
						log.Errorf("unknown type %s", d.GetType())
					}
				}
			}
		}
	}

	requests := GenerateRequests(parameters)

	if Time != 0 {
		RunPerTime(Time, requests)
	}
	d, err := DDNS.Find(parameters, Device.ServiceName)
	Parameters2Save := make([]DDNS.Parameters, 0, len(parameters))
	if err == nil {
		Parameters2Save = append(Parameters2Save, d)
	}
	ExecuteRequests(requests...)
	for _, request := range requests {
		// update info from request.parameters
		Parameters2Save = append(Parameters2Save, request.ToParameters())
	}
	return SaveFromParameters(Parameters2Save...)
}

func RunAuto(GlobalDevice Device.Device, parameters []DDNS.Parameters) error {
	log.Info("get ip address automatically")
	// get ip addr automatically

	/*------------------------------------------------------------------------------------------*/

	devices := GlobalDevice.GetDevices()

	ip4 := Util.MakePair[string, string]() // First is device, second is ip
	ip6 := Util.MakePair[string, string]() // First is device, second is ip

	var (
		err1, err2 error
	)

	var ip4s []string = nil
	var ip6s []string = nil
	for _, device := range devices {
		if ip4s == nil {
			ip4sTemp, err1Temp := Net.GetIpByType(device, Net.A)
			if err1Temp != nil {
				err1 = errors.Join(err1, err1Temp)
				log.Errorf("error getting ipv4 %s ,%s", device, err1)
			} else {
				log.Infof("ipv4 from %s: %s", device, ip4sTemp)
				ip4s, err1 = Net.HandleIp(ip4sTemp, Net.RemoveLoopback)
				ip4.Set(device, ip4s[0])

			}
		}

		if ip6s == nil {
			ip6sTemp, err2Temp := Net.GetIpByType(device, Net.AAAA)
			if err2Temp != nil {
				err2 = errors.Join(err2, err2Temp)
				log.Errorf("error getting ipv6 %s ,%s", device, err2)
			} else {
				log.Infof("ipv6 from %s: %s", device, ip6sTemp)
				ip6s, err2 = Net.HandleIp(ip6sTemp, Net.RemoveLoopback)
				ip6.Set(device, ip6s[0])
			}
		}

		if ip6s != nil && ip4s != nil {
			break
		}

	}

	set := func(parameter DDNS.Parameters) error {
		switch parameter.(DDNS.ServiceParameters).GetType() {
		case "4":
			parameter.(DDNS.ServiceParameters).SetValue(ip4.GetSecond())
			return err1
		case "6":
			parameter.(DDNS.ServiceParameters).SetValue(ip6.GetSecond())
			return err2
		default:
			return fmt.Errorf("unknown type %s", parameter.(DDNS.ServiceParameters).GetType())
		}
	}

	for i, parameter := range parameters {
		if _, ok := parameter.(DDNS.DeviceOverridable); ok {
			// if parameter implements DeviceOverridable interface, set the ip address
			err := set(parameter)
			if err != nil {
				log.Errorf("error setting ip address: %s, skip service:%s", err.Error(), parameter.GetName())
				parameters = append(parameters[:i], parameters[i+1:]...) // ? may be bugs

			}
		}
	}

	requests := GenerateRequests(parameters)

	if Time != 0 {
		RunPerTime(Time, requests)
	}
	d, err := DDNS.Find(parameters, Device.ServiceName)
	Parameters2Save := make([]DDNS.Parameters, 0, len(parameters))
	if err == nil {
		Parameters2Save = append(Parameters2Save, d)
	}
	ExecuteRequests(requests...)
	for _, request := range requests {
		// update info from request.parameters
		Parameters2Save = append(Parameters2Save, request.ToParameters())
	}
	return SaveFromParameters(Parameters2Save...)

}

// GetGlobalDevice get the global device
// if not found, fatal
func GetGlobalDevice(parameters []DDNS.Parameters) (Device.Device, error) {
	deviceInterface, err := DDNS.Find(parameters, Device.ServiceName)
	if err != nil {
		log.Errorf("Section [Devices] not found, check configuration at %s", DDNS.GetConfigureLocation())
		return Device.Device{}, fmt.Errorf("section [Devices] not found, check configuration at %s", DDNS.GetConfigureLocation())
	}

	GlobalDevice, ok := deviceInterface.(Device.Device)
	if !ok {
		log.Errorf("Section [Devices] is not a device, check configuration at %s", DDNS.GetConfigureLocation())
		return Device.Device{}, fmt.Errorf("section [Devices] is not a device, check configuration at %s", DDNS.GetConfigureLocation())
	}
	return GlobalDevice, nil
}

func RunOverride(GlobalDevice Device.Device, parameters []DDNS.Parameters) error {
	// override the ip address here
	// use the Key `Devices` and `Type` of the Service if exist
	log.Info("-O is set, override the ip address")
	var errCount uint16

	for i, parameter := range parameters {
		// skip the device parameter
		if parameter.GetName() != Device.ServiceName {
			// check if parameter implements DeviceOverridable interface
			if d, ok := parameter.(DDNS.DeviceOverridable); ok {
				log.Debugf("Parameter %s implements DeviceOverridable interface", parameter.GetName())

				var tempDeviceName string

				// if device is not set, use Type IP value of Global Devices
				if !d.IsDeviceSet() {
					err := set(GlobalDevice, parameter)
					if err != nil {
						log.Errorf("error setting ip address: %s, skip service:%s", err.Error(), parameter.GetName())
						errCount++
						parameters = append(parameters[:i], parameters[i+1:]...)
						continue
					}

					log.Warnf("Devices of %s is not set, use default value %s", parameter.GetName(), parameter.(DDNS.DeviceOverridable).GetIP())
					continue // skip
				}

				if !d.IsTypeSet() {
					errCount++
					parameters = append(parameters[:i], parameters[i+1:]...)
					log.Errorf("error setting ip address: unknown type, skip service:%s", parameter.GetName())
					continue
				}

				TypeInt, _ := strconv.Atoi(d.GetType())
				tempDeviceName = d.GetDevice()
				ips, err := Net.GetIpByType(tempDeviceName, uint8(TypeInt))
				if err != nil {
					errCount++
					parameters = append(parameters[:i], parameters[i+1:]...)
					log.Errorf("error getting ip address: %s, skip service:%s", err.Error(), parameter.GetName())
					continue
				}
				//

				log.Infof("override %s with %s", parameter.GetName(), ips[0])
				parameter.(DDNS.DeviceOverridable).SetValue(ips[0])
			} else {
				// Service is not DeviceOverridable, use ip got from Devices Section
				err := set(GlobalDevice, parameter)
				if err != nil {
					errCount++
					parameters = append(parameters[:i], parameters[i+1:]...)
					continue
				}
				log.Debugf("Parameter %s is not DeviceOverridable, use default value %s", parameter.GetName(), parameter.(DDNS.ServiceParameters).GetIP())
			}

		}
	} // loop ends

	log.Infof("finish overriding ip with %d error(s)", errCount)

	requests := GenerateRequests(parameters)

	if Time != 0 {
		RunPerTime(Time, requests)
	}
	d, err := DDNS.Find(parameters, Device.ServiceName)
	Parameters2Save := make([]DDNS.Parameters, 0, len(parameters))
	if err == nil {
		Parameters2Save = append(Parameters2Save, d)
	}
	ExecuteRequests(requests...)
	for _, request := range requests {
		// update info from request.parameters
		Parameters2Save = append(Parameters2Save, request.ToParameters())
	}
	return SaveFromParameters(Parameters2Save...)

}

func set(GlobalDevice Device.Device, ParameterToSet DDNS.Parameters) error {

	toSet := ParameterToSet.(DDNS.ServiceParameters)
	Type := toSet.GetType() // Type is "4" or "6" or ""

	ip := Util.Pair[string, string]{} // First is device, second is ip

	var err error
	devices := GlobalDevice.GetDevices()
	var ips []string = nil

	// if failed to get ip, then try to get ip of next device in the list, else break
	switch Type {
	case "4":
		for _, device := range devices {
			if ips == nil {
				ip4sTemp, errTemp := Net.GetIpByType(device, Net.A)
				if errTemp != nil {
					err = errors.Join(err, errTemp)
					log.Errorf("error getting ipv4 %s ,%s", device, err)
				} else {
					log.Infof("ipv4 from %s: %s", device, ip4sTemp)
					ips, errTemp = Net.HandleIp(ip4sTemp, Net.RemoveLoopback)
					if errTemp != nil {
						err = errors.Join(err, errTemp)
						log.Errorf("error handling ipv4 %s ,%s", device, err)
					}
					ip.Set(device, ips[0])
				}
			}
		}

	case "6":
		for _, device := range devices {
			if ips == nil {
				ip6sTemp, errTemp := Net.GetIpByType(device, Net.AAAA)
				if errTemp != nil {
					err = errors.Join(err, errTemp)
					log.Errorf("error getting ipv6 %s ,%s", device, err)
				} else {
					log.Infof("ipv6 from %s: %s", device, ip6sTemp)
					ips, errTemp = Net.HandleIp(ip6sTemp, Net.RemoveLoopback)
					if errTemp != nil {
						err = errors.Join(err, errTemp)
						log.Errorf("error handling ipv4 %s ,%s", device, err)
					}
					ip.Set(device, ips[0])
				}
			}
		}

	default:
		return fmt.Errorf("unknown type %s", ParameterToSet.(DDNS.ServiceParameters).GetType())
	}

	if ip.GetFirst() != "" && ip.GetSecond() != "" {
		ParameterToSet.(DDNS.ServiceParameters).SetValue(ip.GetSecond())
		return nil
	} else {
		return err
	}

}

func GenerateConfigure(configFactoryList []DDNS.ConfigFactory) error {
	if DDNS.IsConfigureExist() {
		log.Warnf("configure at %s already exist", DDNS.GetConfigureLocation())
		return errors.New("configure already exist")
	}
	log.Debugf("start generating default configure")
	err := GenerateDefaultConfigure(configFactoryList...)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Info("generate a default config file at ", DDNS.GetConfigureLocation())
	return nil
}

func ExecuteRequests(requests ...DDNS.Request) {
	log.Info("start executing requests")

	deal := func(err error, request DDNS.Request) {
		if err != nil || request.Status().Status != DDNS.Success {
			log.Errorf("error executing request, %s", err)
			Retry(request, retryAttempt)
		}

		status := ""
		res := request.Status()
		if res.Status == DDNS.Success {
			status = "Success"
			log.Infof("name:%s, status:%s  msg:%s", res.Name, status, res.Msg)
		} else if res.Status == DDNS.Failed {
			log.Errorf("error executing request, %s", err.Error())
			status = "Failed"
			log.Infof("name:%s, status:%s, msg:%s", res.Name, status, res.Msg)
			if retryAttempt != 0 {
				log.Errorf("all retry failed, skip %s", request.GetName())
			}
		} else if res.Status == DDNS.NotExecute {
			log.Fatal("request not executed")
		}
	}

	if proxyEnable {
		for _, request := range requests {
			var err error
			log.Tracef("request: %s", request.GetName())
			throughProxy, ok := request.(DDNS.ThroughProxy)
			if ok {
				err = throughProxy.RequestThroughProxy()
			} else {
				err = DDNS.ExecuteRequest(request)
			}
			deal(err, request)
		}

	} else {
		for _, request := range requests {
			log.Tracef("request: %s", request.GetName())
			err := DDNS.ExecuteRequest(request)
			if err != nil || request.Status().Status != DDNS.Success {
				log.Errorf("error executing request, %s", err.Error())
				Retry(request, retryAttempt)
			}
			deal(err, request)
		}
	}
}

func Retry(request DDNS.Request, i uint8) {
	for j := uint8(1); j <= i; j++ {
		log.Warnf("retrying %s %d time", request.GetName(), j)

		if proxyEnable {
			throughProxy, ok := request.(DDNS.ThroughProxy)
			if ok {
				err := throughProxy.RequestThroughProxy()
				if err != nil {
					log.Errorf("error: %s", err.Error())
				} else {
					return
				}
			}
		} else {
			err := DDNS.ExecuteRequest(request)
			if err != nil {
				log.Errorf("error: %s", err.Error())
			} else {
				return
			}
		}

	}
}

func GenerateRequests(parameters []DDNS.Parameters) []DDNS.Request {
	log.Info("start generating requests")
	var errCount uint8 = 0
	requests := make([]DDNS.Request, 0, len(parameters))
	for _, parameter := range parameters {
		if parameter.GetName() == Device.ServiceName {
			continue // skip
		}

		log.Infof("service: %s", parameter.GetName())
		request, err := parameter.(DDNS.ServiceParameters).ToRequest()
		if err != nil {
			errCount++
			log.Errorf("error generating request for %s:%s ", parameter.GetName(), err.Error())
			continue
		}
		requests = append(requests, request)
	}
	log.Infof("finish generating requests with %d error(s)", errCount)
	return requests
}

func GenerateDefaultConfigure(ConfigFactories ...DDNS.ConfigFactory) error {
	var infos []DDNS.ConfigStr
	var err error
	for _, factory := range ConfigFactories {
		info, errTemp := factory.Get().GenerateDefaultConfigInfo()
		log.Tracef("config info: \n%s", info)
		if errTemp != nil {
			err = errors.Join(err, errTemp)
		}
		infos = append(infos, info)
	}
	errTemp := DDNS.ConfigureWriter(DDNS.GetConfigureLocation(), os.O_CREATE|os.O_WRONLY, infos...)
	err = errors.Join(err, errTemp)
	if err != nil {
		return err
	}
	log.Info("write default configure to ", DDNS.GetConfigureLocation())
	return nil
}

// RunPerTime run ddns per time
// todo fix && save config when success
// Expect: run 1 --time-- run 2 --time-- run 3 ...
// Actual   : --time-- run 1 --time-- run 2 --time-- run 3 ...
func RunPerTime(Time uint64, requests []DDNS.Request) {

	log.Infof("run ddns per %d seconds", Time)

	c := cron.New()
	for _, request := range requests {
		_, err := c.AddJob(fmt.Sprintf("@every %ds", Time), cron.NewChain(cron.DelayIfStillRunning(cron.DefaultLogger)).Then(request))
		if err != nil {
			log.Errorf("error adding job %s: %s", request.GetName(), err.Error())
		}
	}

	c.Start()
	select {}

}

func SaveFromParameters(parameters ...DDNS.Parameters) error {
	// todo Merge Parameters that differ only by subdomain
	err := DDNS.SaveConfig(DDNS.GetConfigureLocation(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, parameters...)
	if err != nil {
		log.Errorf("error saving config: %s", err.Error())
		return fmt.Errorf("error saving config: %w", err)
	}
	log.Infof("save config to %s", DDNS.GetConfigureLocation())
	return nil
}

// Requests2Parameters convert successful requests to parameters
// ![deprecated]
func Requests2Parameters(requests []DDNS.Request) []DDNS.Parameters {
	Parameters2Save := make([]DDNS.Parameters, 0, len(requests))
	for _, request := range requests {
		if request.Status().Status == DDNS.Success {
			Parameters2Save = append(Parameters2Save, request.ToParameters())
		}
	}
	return Parameters2Save
}

func CheckVersionUpgrade(msg chan<- string) {
	// start checking version upgrade
	hasUpgrades, v, url, err := DDNS.CheckUpdate()
	defer close(msg)
	defer func() {
		log.Tracef("check version upgrade finished")
	}()
	if err != nil {
		if errors.Is(err, DDNS.NoCompatibleVersionError) {
			// "no suitable version")
			msg <- fmt.Sprintf("new version %s is available\n", v.Info())
			msg <- fmt.Sprintf("no compatible release for your operating system, consider building from source:%s \n", DDNS.RepoURLs())
		}
		// error checking version upgrade
		msg <- ""
		msg <- ""
		return
	}

	if hasUpgrades {
		msg <- fmt.Sprintf("new version %s is available\n", v.Info())
		msg <- fmt.Sprintf("download url: %s", url)
	} else {
		// "already the latest version")
		return
	}
}
