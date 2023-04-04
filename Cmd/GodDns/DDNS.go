package main

import (
	"GodDns/Core"
	"GodDns/Device"
	log "GodDns/Log"
	"GodDns/Net"
	"GodDns/Util/Collections"
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	"os"
	"strconv"
	"sync"
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
	return GenerateExecuteSave(parameters)
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
			if d, ok := parameter.(DDNS.Service); ok {
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

func RunAuto(GlobalDevice Device.Device, parameters []DDNS.Parameters) error {
	log.Info("get ip address automatically")
	// get ip addr automatically

	/*------------------------------------------------------------------------------------------*/

	devices := GlobalDevice.GetDevices()

	ip4 := Collections.MakePair[string, string]() // First is device, second is ip
	ip6 := Collections.MakePair[string, string]() // First is device, second is ip

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
		switch parameter.(DDNS.Service).GetType() {
		case "4":
			parameter.(DDNS.Service).SetValue(ip4.GetSecond())
			return err1
		case "6":
			parameter.(DDNS.Service).SetValue(ip6.GetSecond())
			return err2
		default:
			return fmt.Errorf("unknown type %s", parameter.(DDNS.Service).GetType())
		}
	}

	newParameters := make([]DDNS.Parameters, 0, len(parameters))
	for _, parameter := range parameters {
		if deviceOverridable, ok := parameter.(DDNS.DeviceOverridable); ok {
			// if parameter implements DeviceOverridable interface, set the ip address
			if err := set(parameter); err != nil {
				log.Errorf("error setting ip address: %s, skip service:%s", err.Error(), deviceOverridable.GetName())
				continue // skip
			}
			newParameters = append(newParameters, deviceOverridable)
		} else {
			newParameters = append(newParameters, parameter)
		}
	}
	parameters = newParameters

	return GenerateExecuteSave(parameters)

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

	for _, parameter := range parameters {
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
						log.Errorf("error setting ip address: %s, skip overriding service:%s", err.Error(), parameter.GetName())
						errCount++

						continue
					}
					log.Warnf("Devices of %s is not set, use default value %s", parameter.GetName(), parameter.(DDNS.DeviceOverridable).GetIP())
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
				parameter.(DDNS.DeviceOverridable).SetValue(ips[0])
			} else {
				// Service is not DeviceOverridable, use ip got from Devices Section
				err := set(GlobalDevice, parameter)
				if err != nil {
					errCount++
					log.Errorf("error setting ip address: %s, use default value:%s", err.Error(), parameter.(DDNS.Service).GetIP())
					continue
				}
				log.Debugf("Parameter %s is not DeviceOverridable, use default value %s", parameter.GetName(), parameter.(DDNS.Service).GetIP())
			}

		}
	} // loop ends

	log.Infof("finish overriding ip with %d error(s)", errCount)

	return GenerateExecuteSave(parameters)

}

func set(GlobalDevice Device.Device, ParameterToSet DDNS.Parameters) error {

	toSet := ParameterToSet.(DDNS.Service)
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
		return fmt.Errorf("unknown type %s", ParameterToSet.(DDNS.Service).GetType())
	}

	if ip.GetFirst() != "" && ip.GetSecond() != "" {
		ParameterToSet.(DDNS.Service).SetValue(ip.GetSecond())
		return nil
	} else {
		return err
	}

}

func GenerateExecuteSave(parameters []DDNS.Parameters) error {
	requests := GenerateRequests(parameters)

	d, err := DDNS.Find(parameters, Device.ServiceName)
	Parameters2Save := make([]DDNS.Parameters, 0, len(parameters))
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

func Display(request DDNS.Request) {
	_, _ = log.InfoPP.Fprintln(output, fmt.Sprint("displaying message from Service ", request.GetName(), " at ", request.Target()))
	serviceInfo := request.Status().MG.GetMsgOf(DDNS.Info)
	if len(serviceInfo) > 0 {
		for _, i := range serviceInfo {
			_, _ = log.SuccessPP.Fprintln(output, i)
		}
	}

	serviceErr := request.Status().MG.GetMsgOf(DDNS.Error)
	serviceWarn := request.Status().MG.GetMsgOf(DDNS.Warn)
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

func DisplayAll(requests ...DDNS.Request) {
	for _, request := range requests {
		Display(request)
		_, _ = fmt.Fprintln(output)
	}
}

func GenerateConfigure(configFactoryList []DDNS.ConfigFactory) error {
	if DDNS.IsConfigExist(DDNS.GetConfigureLocation()) {
		log.Warnf("configure at %s already exist", DDNS.GetConfigureLocation())
		return errors.New("configure already exist")
	}
	log.Debugf("start generating default configure")
	err := GenerateDefaultConfigure(configFactoryList...)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Infof("generate a default config file at %s", DDNS.GetConfigureLocation())
	return nil
}

func ExecuteRequests(requests ...DDNS.Request) {
	log.Info("start executing requests")
	var wg sync.WaitGroup

	deal := func(err error, request DDNS.Request) {
		defer wg.Done()
		if err != nil || (request).Status().Status != DDNS.Success {
			log.ErrorRaw(fmt.Sprintf("error executing request, %v", err))
			// Retry(request, retryAttempt)
			for j := uint8(1); j <= retryAttempt; j++ {
				errMsg := fmt.Sprintf("retrying %s:%s, attempt %d", request.GetName(), request.Target(), j)
				log.WarnRaw(errMsg)
				request.Status().MG.AddError(fmt.Sprintf("retrying %s:%s, attempt %d", request.GetName(), request.Target(), j))

				if proxyEnable {
					throughProxy, ok := request.(DDNS.ThroughProxy)
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

		status := ""
		res := (request).Status()
		if res.Status == DDNS.Success {
			status = "Success"
			log.InfoRaw(fmt.Sprintf("name:%s, status:%s  msg:%s", res.Name, status, res.MG))
		} else if res.Status == DDNS.Failed {
			errMsg := fmt.Sprintf("error executing request, %v", err)

			log.ErrorRaw(errMsg)
			status = "Failed"
			log.InfoRaw(fmt.Sprintf("name:%s, status:%s, msg:%s", res.Name, status, res.MG))
			if retryAttempt != 0 {
				log.ErrorRaw(fmt.Sprintf("all retry failed, skip %s:%s", (request).GetName(), (request).Target()))
			}
		} else if res.Status == DDNS.NotExecute {
			log.Fatal("request not executed")
		}

	}

	if proxyEnable {
		for _, request := range requests {
			request := request
			wg.Add(1)
			go func() {
				var err error
				log.Tracef("request: %s", request.GetName())
				throughProxy, ok := request.(DDNS.ThroughProxy)
				if ok {
					err = throughProxy.RequestThroughProxy()
				} else {
					err = request.MakeRequest()
				}
				deal(err, request)
			}()
			if !parallelExecuting {
				wg.Wait()
			}
		}
		wg.Wait()

	} else {
		for _, request := range requests {
			request := request
			wg.Add(1)
			go func() {
				var err error
				log.Tracef("request: %s", request.GetName())
				err = request.MakeRequest()

				deal(err, request)
			}()
			if !parallelExecuting {
				wg.Wait()
			}
		}
		wg.Wait()
	}

	log.Info("all requests finished")
}

func Retry(request DDNS.Request, i uint8) {

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
		request, err := parameter.(DDNS.Service).ToRequest()
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
func RunPerTime(Time uint64, GlobalDevice *Device.Device, parameters []DDNS.Parameters) {

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
	newServiceCronJob.SetTimes(TimesLimitation)
	_, err = c.AddJob(fmt.Sprintf("@every %ds", Time), cron.NewChain(cron.Recover(logger), cron.DelayIfStillRunning(cron.DefaultLogger)).Then(newServiceCronJob))
	if err != nil {
		log.Errorf("error adding job : %s", err.Error())
	}

	c.Start()
	wg.Wait()
	log.Info("all jobs finished", log.Uint64("total execution time", TimesLimitation).String())

}

// SaveConfig save parameters to file with flag
func SaveConfig(FileName string, flag int, parameters ...DDNS.Parameters) error {
	var err error
	n := make(map[string]uint)
	ConfigStrings := make([]DDNS.ConfigStr, 0, len(parameters))
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
	err = errors.Join(err, DDNS.ConfigureWriter(FileName, flag, ConfigStrings...))
	return err
}

func SaveFromParameters(parameters ...DDNS.Parameters) error {
	// todo Merge Parameters that differ only by subdomain
	err := SaveConfig(DDNS.GetConfigureLocation(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, parameters...)
	if err != nil {
		log.Errorf("error saving config: %s", err.Error())
		return fmt.Errorf("error saving config: %w", err)
	}
	log.Infof("save config to %s", DDNS.GetConfigureLocation())
	return nil
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
			// "no suitable version"
			if hasUpgrades {
				msg <- fmt.Sprintf("new version %s is available", v.Info())
				msg <- fmt.Sprintf("no compatible release for your operating system, consider building from source:%s ", DDNS.RepoURLs())
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

type ServiceCronJob struct {
	ps           []DDNS.Parameters
	GlobalDevice *Device.Device
	wg           *sync.WaitGroup
	times        uint64
}

func (r *ServiceCronJob) SetTimes(times uint64) {
	r.times = times
	for i := uint64(0); i < times; i++ {
		r.wg.Add(1)
	}
}

func (r *ServiceCronJob) SetWg(wg *sync.WaitGroup) {
	r.wg = wg
}

func NewServiceCronJob(g *Device.Device, ps ...DDNS.Parameters) *ServiceCronJob {
	return &ServiceCronJob{ps: ps, GlobalDevice: g}
}

func (r *ServiceCronJob) Run() {
	if r.times == 0 {
		return
	}
	defer r.wg.Done()
	defer func() { r.times-- }()

	switch runMode {
	case run:
		err := RunDDNS(r.ps)
		if err != nil {
			log.Error("error running ddns: ", log.String("error", err.Error()))
		}
	case runApi:
		err := RunGetFromApi(r.ps)
		if err != nil {
			log.Error("error running api: ", log.String("error", err.Error()))
		}
	case runAuto:
		if r.GlobalDevice == nil {
			log.Error("error running auto: ", log.String("error", "no global device"))
			panic("no global device")
		}

		err := RunAuto(*r.GlobalDevice, r.ps)
		if err != nil {
			log.Error("error running auto: ", log.String("error", err.Error()))
		}
	case runAutoOverride:
		if r.GlobalDevice == nil {
			log.Error("error running auto: ", log.String("error", "no global device"))
			panic("no global device")
		}

		err := RunOverride(*r.GlobalDevice, r.ps)
		if err != nil {
			log.Error("error running override: ", log.String("error", err.Error()))
		}
	default:
		panic("unknown run mode")
	}

}
