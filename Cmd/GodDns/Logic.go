package main

import (
	"GodDns/Core"
	"GodDns/Device"
	log "GodDns/Log"
	"GodDns/Net"
	"GodDns/Util/Collections"
	"errors"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"github.com/robfig/cron/v3"
	"os"
	"strconv"
	"sync"
	"time"
)

// -----------------------------------------------------------------------------------------------------------------------------------------//

func ModeController(ps []Core.Parameters, GlobalDevice *Device.Device) error {
	switch runMode {
	case run:
		err := RunDDNS(ps)
		if err != nil {
			return err
		}
	case runApi:
		err := RunGetFromApi(ps)
		if err != nil {
			return err
		}
	case runAuto:
		if GlobalDevice == nil {
			panic("no global device")
		}
		err := RunAuto(*GlobalDevice, ps)
		if err != nil {
			return err
		}
	case runAutoOverride:
		if GlobalDevice == nil {
			panic("no global device")
		}
		err := RunOverride(*GlobalDevice, ps)
		if err != nil {
			return err
		}
	default:
		panic("unknown run mode")
	}
	return nil
}

func RunDDNS(parameters []Core.Parameters) error {
	log.Debugf("run ddns")
	// run ddns here

	// get from api
	if ApiName != "" {
		return RunGetFromApi(parameters)
	}

	// -A is not set
	return GenerateExecuteSave(parameters)
}

type d2i map[string]Collections.Pair[string, string]

func (d *d2i) Add(device string, ip string, t Net.Type) {
	switch t {
	case Net.A:
		(map[string]Collections.Pair[string, string])(*d)[device] = Collections.Pair[string, string]{
			First:  &ip,
			Second: (map[string]Collections.Pair[string, string])(*d)[device].Second,
		}
	case Net.AAAA:
		(map[string]Collections.Pair[string, string])(*d)[device] = Collections.Pair[string, string]{
			First:  (map[string]Collections.Pair[string, string])(*d)[device].First,
			Second: &ip,
		}
	default:
		panic("invalid type")
	}

}

var Device2Ips = make(d2i, 20)

func ReadConfig(configs []Core.ConfigFactory) ([]Core.Parameters, error) {
	parameters, fileErr, configErrs := Core.ConfigureReader(Core.GetConfigureLocation(), configs...)
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
func RunGetFromApi(parameters []Core.Parameters) error {

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
	_ = Core.MainGoroutinePool.Submit(func() {
		_ip4, err1 = api.Get(4)
		if err1 != nil {
			ip4Done <- false
		} else {
			ip4 = _ip4
			ip4Done <- true
		}
		close(ip4Done)
	})

	var err2 error
	_ = Core.MainGoroutinePool.Submit(func() {

		_ip6, err2 = api.Get(6)
		if err2 != nil {

			ip6Done <- false
		} else {
			ip6 = _ip6

			ip6Done <- true
		}
		close(ip6Done)
	})

	var ipv4Received, ipv6Received bool
	for {
		select {
		case temp := <-ip4Done:
			if !ipv4Received {
				if temp {
					log.Infof("ipv4 from %s: %s", ApiName, ip4)
					ipv4Received = true
				} else {
					log.Errorf("error getting ipv4, %s", err1)
					return errors.New("quit")
				}
			}

		case temp := <-ip6Done:
			if !ipv6Received {
				if temp {
					log.Infof("ipv6 from %s: %s", ApiName, ip6)
					ipv6Received = true
				} else {
					log.Errorf("error getting ipv6, %s", err2)
					return errors.New("quit")
				}
			}

		case <-time.After(10 * time.Second):
			log.Errorf("timeout getting ip address from %s", ApiName)
			return errors.New("quit")
		}

		if ipv4Received && ipv6Received {
			break
		}
	}

	for _, parameter := range parameters {
		if parameter.GetName() != Device.ServiceName {
			if d, ok := parameter.(Core.Service); ok {
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

	return GenerateExecuteSave(parameters)
}

func RunAuto(GlobalDevice Device.Device, parameters []Core.Parameters) error {
	log.Info("get ip address automatically")
	// get ip addr automatically

	/*------------------------------------------------------------------------------------------*/

	devices := GlobalDevice.GetDevices()

	// First is device, second is ip
	ip4 := Collections.MakePair[string, string]()
	// First is device, second is ip
	ip6 := Collections.MakePair[string, string]()

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
				ip4s, err1 = Net.HandleIp(ip4sTemp)
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
				ip6s, err2 = Net.HandleIp(ip6sTemp)
				ip6.Set(device, ip6s[0])
			}
		}

		if ip6s != nil && ip4s != nil {
			break
		}

	}
	Device2Ips.Add(ip4.GetFirst(), ip4.GetSecond(), Net.A)
	Device2Ips.Add(ip6.GetFirst(), ip6.GetSecond(), Net.AAAA)

	set := func(parameter Core.Service) error {
		switch parameter.GetType() {
		case "4":
			MainBinder.Bind(ip4.GetFirst(), &parameter)
			parameter.SetValue(ip4.GetSecond())

			return err1
		case "6":
			MainBinder.Bind(ip6.GetFirst(), &parameter)
			parameter.SetValue(ip6.GetSecond())

			return err2
		default:
			return fmt.Errorf("unknown type %s", parameter.GetType())
		}
	}

	newParameters := make([]Core.Parameters, 0, len(parameters))
	for _, parameter := range parameters {
		if service, ok := parameter.(Core.Service); ok {
			// if parameter implements DeviceOverridable interface, set the ip address
			if err := set(service); err != nil {
				log.Errorf("error setting ip address: %s, skip service:%s", err.Error(), service.GetName())
				continue // skip
			}
			newParameters = append(newParameters, service)
		} else {
			newParameters = append(newParameters, parameter)
		}
	}
	parameters = newParameters

	return GenerateExecuteSave(parameters)

}

// GetGlobalDevice get the global device
// if not found, fatal
func GetGlobalDevice(parameters []Core.Parameters) (Device.Device, error) {
	deviceInterface, err := Core.Find(parameters, Device.ServiceName)
	if err != nil {
		log.Errorf("Section [Devices] not found, check configuration at %s", Core.GetConfigureLocation())
		return Device.Device{}, fmt.Errorf("section [Devices] not found, check configuration at %s", Core.GetConfigureLocation())
	}

	GlobalDevice, ok := deviceInterface.(Device.Device)
	if !ok {
		log.Errorf("Section [Devices] is not a device, check configuration at %s", Core.GetConfigureLocation())
		return Device.Device{}, fmt.Errorf("section [Devices] is not a device, check configuration at %s", Core.GetConfigureLocation())
	}
	return GlobalDevice, nil
}

func RunOverride(GlobalDevice Device.Device, parameters []Core.Parameters) error {
	// override the ip address here
	// use the Key `Devices` and `Type` of the Service if exist
	log.Info("-O is set, override the ip address")
	var errCount uint16

	for _, parameter := range parameters {
		// skip the device parameter
		if parameter.GetName() != Device.ServiceName {
			// check if parameter implements DeviceOverridable interface
			if d, ok := parameter.(Core.DeviceOverridable); ok {
				log.Debugf("Parameter %s implements DeviceOverridable interface", parameter.GetName())

				var tempDeviceName string

				// if device is not set, use Type IP value of Global Devices
				if !d.IsDeviceSet() {
					err := set(GlobalDevice, parameter)
					if err != nil {
						log.Errorf("error setting ip address: %s, skip overriding service:%s", err.Error(), parameter.GetName())
						errCount++

						continue
					}
					log.Warnf("Devices of %s is not set, use default value %s", parameter.GetName(), parameter.(Core.DeviceOverridable).GetIP())
					continue // skip
				}

				if !d.IsTypeSet() {
					errCount++

					log.Errorf("error setting ip address: unknown type, skip service:%s", parameter.GetName())
					continue
				}

				TypeInt, _ := strconv.Atoi(d.GetType())
				tempDeviceName = d.GetDevice()
				ips, err := Net.GetIpByType(tempDeviceName, uint8(TypeInt))
				if err != nil {
					errCount++

					log.Errorf("error getting ip address: %s, skip service:%s", err.Error(), parameter.GetName())
					continue
				}

				log.Infof("override %s with %s", parameter.GetName(), ips[0])
				parameter.(Core.DeviceOverridable).SetValue(ips[0])
			} else {
				// Service is not DeviceOverridable, use ip got from Devices Section
				err := set(GlobalDevice, parameter)
				if err != nil {
					errCount++
					log.Errorf("error setting ip address: %s, use default value:%s", err.Error(), parameter.(Core.Service).GetIP())
					continue
				}
				log.Debugf("Parameter %s is not DeviceOverridable, use default value %s", parameter.GetName(), parameter.(Core.Service).GetIP())
			}

		}
	} // loop ends

	log.Infof("finish overriding ip with %d error(s)", errCount)

	return GenerateExecuteSave(parameters)

}

func set(GlobalDevice Device.Device, ParameterToSet Core.Parameters) error {

	toSet := ParameterToSet.(Core.Service)
	Type := toSet.GetType() // Type is "4" or "6" or ""

	ip := Collections.MakePair[string, string]() // First is device, second is ip

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
		return fmt.Errorf("unknown type %s", ParameterToSet.(Core.Service).GetType())
	}

	if ip.GetFirst() != "" && ip.GetSecond() != "" {
		ParameterToSet.(Core.Service).SetValue(ip.GetSecond())
		return nil
	} else {
		return err
	}

}

func GenerateExecuteSave(parameters []Core.Parameters) error {
	requests := GenerateRequests(parameters)

	d, err := Core.Find(parameters, Device.ServiceName)
	Parameters2Save := make([]Core.Parameters, 0, len(parameters))
	if err == nil {
		Parameters2Save = append(Parameters2Save, d)
	}
	ExecuteRequests(requests...)
	DisplayAll(requests...)
	for _, request := range requests {
		// update info from request.parameters
		s2p := request.ToParameters()
		Parameters2Save = append(Parameters2Save, s2p)
	}

	err = SaveFromParameters(Parameters2Save...)

	if err != nil {
		return err
	}
	return nil
}

func Display(request Core.Request) {
	_, _ = log.InfoPP.Fprintln(output, fmt.Sprint("displaying message from Service ", request.GetName(), " at ", request.Target()))
	serviceInfo := request.Status().MG.GetMsgOf(Core.Info)
	if len(serviceInfo) > 0 {
		for _, i := range serviceInfo {
			_, _ = log.SuccessPP.Fprintln(output, i)
		}
	}

	serviceErr := request.Status().MG.GetMsgOf(Core.Error)
	serviceWarn := request.Status().MG.GetMsgOf(Core.Warn)
	if len(serviceErr) > 0 {
		for _, e := range serviceErr {
			_, _ = log.ErrPP.Fprintln(output, e)
		}
	}
	if len(serviceWarn) > 0 {
		for _, w := range serviceWarn {
			_, _ = log.WarnPP.Fprintln(output, w)
		}
	}
}

func DisplayAll(requests ...Core.Request) {
	for _, request := range requests {
		Display(request)
		_, _ = fmt.Fprintln(output)
	}
}

func GenerateConfigure(configFactoryList []Core.ConfigFactory) error {
	if Core.IsConfigExist(Core.GetConfigureLocation()) {
		log.Warnf("configure at %s already exist", Core.GetConfigureLocation())
		return errors.New("configure already exist")
	}
	log.Debugf("start generating default configure")
	err := GenerateDefaultConfigure(configFactoryList...)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Infof("generate a default config file at %s", Core.GetConfigureLocation())
	return nil
}

func ExecuteRequests(requests ...Core.Request) {
	log.Info("start executing requests")
	var wg sync.WaitGroup

	deal := func(err error, request Core.Request) {

		if err != nil || (request).Status().Status != Core.Success {
			log.ErrorRaw(fmt.Sprintf("error executing request, %v", err))
			Retry(request, retryAttempt)
		}

		status := ""
		res := (request).Status()
		if res.Status == Core.Success {
			status = "Success"
			log.InfoRaw(fmt.Sprintf("name:%s, status:%s  msg:%s", res.Name, status, res.MG))
		} else if res.Status == Core.Failed {
			errMsg := fmt.Sprintf("error executing request, %v", err)

			log.ErrorRaw(errMsg)
			status = "Failed"
			log.InfoRaw(fmt.Sprintf("name:%s, status:%s, msg:%s", res.Name, status, res.MG))
			if retryAttempt != 0 {
				log.ErrorRaw(fmt.Sprintf("all retry failed, skip %s:%s", (request).GetName(), (request).Target()))
			}
		} else if res.Status == Core.NotExecute {
			log.Fatal("request not executed")
		}

	}

	defer ants.Release()
	if proxyEnable {
		for _, request := range requests {
			request := request
			wg.Add(1)

			_ = Core.MainGoroutinePool.Submit(func() {
				var err error
				defer wg.Done()
				log.Tracef("request: %s", request.GetName())
				throughProxy, ok := request.(Core.ThroughProxy)
				if ok {
					err = throughProxy.RequestThroughProxy()
				} else {
					err = request.MakeRequest()
				}
				deal(err, request)
			})

			if !parallelExecuting {
				wg.Wait()
			}
		}
		wg.Wait()

	} else {
		for _, request := range requests {
			request := request
			wg.Add(1)
			_ = Core.MainGoroutinePool.Submit(func() {
				defer wg.Done()
				var err error
				log.Tracef("request: %s", request.GetName())
				err = request.MakeRequest()

				deal(err, request)
			})
			if !parallelExecuting {
				wg.Wait()
			}
		}
		wg.Wait()
	}

	log.Info("all requests finished")
}

func Retry(request Core.Request, i uint8) {
	for j := uint8(1); j <= i; j++ {
		errMsg := fmt.Sprintf("retrying %s:%s, attempt %d", request.GetName(), request.Target(), j)
		log.WarnRaw(errMsg)
		request.Status().MG.AddError(fmt.Sprintf("retrying %s:%s, attempt %d", request.GetName(), request.Target(), j))

		if proxyEnable {
			throughProxy, ok := request.(Core.ThroughProxy)
			if ok {
				err := throughProxy.RequestThroughProxy()
				if err != nil {
					throughProxy.Status().MG.AddError(fmt.Sprintf("error executing request, %v", err))
					log.ErrorRaw(fmt.Sprintf("error: %s", err.Error()))
				} else {
					return
				}
			}
		} else {
			err := request.MakeRequest()
			if err != nil {
				request.Status().MG.AddError(fmt.Sprintf("error: %s", err.Error()))
				log.ErrorRaw(fmt.Sprintf("error: %s", err.Error()))
			} else {
				return
			}
		}

	}
}

func GenerateRequests(parameters []Core.Parameters) []Core.Request {
	log.Info("start generating requests")
	var errCount uint8 = 0
	requests := make([]Core.Request, 0, len(parameters))
	for _, parameter := range parameters {
		if parameter.GetName() == Device.ServiceName {
			continue // skip
		}

		log.Infof("service: %s", parameter.GetName())
		request, err := parameter.(Core.Service).ToRequest()
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

func GenerateDefaultConfigure(ConfigFactories ...Core.ConfigFactory) error {
	var infos []Core.ConfigStr
	var err error
	for _, factory := range ConfigFactories {
		info, errTemp := factory.Get().GenerateDefaultConfigInfo()
		log.Tracef("config info: \n%s", info)
		if errTemp != nil {
			err = errors.Join(err, errTemp)
		}
		infos = append(infos, info)
	}
	errTemp := Core.ConfigureWriter(Core.GetConfigureLocation(), os.O_CREATE|os.O_WRONLY, infos...)
	err = errors.Join(err, errTemp)
	if err != nil {
		return err
	}
	log.Info("write default configure to ", Core.GetConfigureLocation())
	return nil
}

// RunPerTime run ddns per time
func RunPerTime(Time uint64, GlobalDevice *Device.Device, parameters []Core.Parameters) {

	log.Infof("run ddns per %d seconds", Time)

	cornLogfile, err := os.OpenFile("cron.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Debug(err)
	}

	logger := log.NewLogger(cornLogfile)
	logger = logger.WithGroup("cron:")
	c := cron.New(cron.WithLogger(cron.VerbosePrintfLogger(logger)))
	newServiceCronJob := NewServiceCronJob(GlobalDevice, parameters...)
	wg := new(sync.WaitGroup)
	newServiceCronJob.SetWg(wg)
	if TimesLimitation == 0 {
		TimesLimitation = MAXTIMES
	}
	newServiceCronJob.SetTimes(TimesLimitation)
	_, err = c.AddJob(fmt.Sprintf("@every %ds", Time), cron.NewChain(cron.Recover(logger), cron.DelayIfStillRunning(cron.DefaultLogger)).Then(newServiceCronJob))
	if err != nil {
		log.Errorf("error adding job : %s", err.Error())
	}

	c.Start()
	wg.Wait()
	log.Info("all jobs finished", log.Int("total execution time", TimesLimitation).String())

}

// SaveConfig save parameters to file with flag
func SaveConfig(FileName string, flag int, parameters ...Core.Parameters) error {
	var err error
	n := make(map[string]uint)
	ConfigStrings := make([]Core.ConfigStr, 0, len(parameters))
	for _, parameter := range parameters {
		var no uint
		if parameter.GetName() == Device.ServiceName {
			no = 0
		} else {
			n[parameter.GetName()]++
			no = n[parameter.GetName()]
		}
		ConStr, err_ := parameter.SaveConfig(no)

		if err_ != nil {
			err = errors.Join(err, err_)
		} else {
			ConfigStrings = append(ConfigStrings, ConStr)
		}
	}
	err = errors.Join(err, Core.ConfigureWriter(FileName, flag, ConfigStrings...))
	return err
}

func SaveFromParameters(parameters ...Core.Parameters) error {
	// todo Merge Parameters that differ only by subdomain
	err := SaveConfig(Core.GetConfigureLocation(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, parameters...)
	if err != nil {
		log.Errorf("error saving config: %s", err.Error())
		return fmt.Errorf("error saving config: %w", err)
	}
	log.Infof("save config to %s", Core.GetConfigureLocation())
	return nil
}

func CheckVersionUpgrade(msg chan<- string) {
	// start checking version upgrade
	hasUpgrades, v, url, err := Core.CheckUpdate()
	defer close(msg)
	defer func() {
		log.Tracef("check version upgrade finished")
	}()
	if err != nil {
		if errors.Is(err, Core.NoCompatibleVersionError) {
			// "no suitable version"
			if hasUpgrades {
				msg <- fmt.Sprintf("new version %s is available", v.Info())
				msg <- fmt.Sprintf("no compatible release for your operating system, consider building from source:%s ", Core.RepoURLs())
			} else {
				// "already the latest version"
				msg <- ""
				msg <- ""
				return
			}

		}
		// error checking version upgrade
		msg <- ""
		msg <- ""
		return
	}

	if hasUpgrades {
		msg <- fmt.Sprintf("new version %s is available", v.Info())
		msg <- fmt.Sprintf("download url: %s", url)
	} else {
		// "already the latest version"
		msg <- ""
		msg <- ""
		return
	}
}
