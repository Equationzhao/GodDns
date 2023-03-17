/*
 *     @Copyright
 *     @file: main.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/18 上午3:43
 *     @last modified: 2023/3/18 上午3:42
 *
 *
 *
 */

package main

import (
	"GodDns/DDNS"
	"GodDns/Device"
	"GodDns/Log"
	"GodDns/Net"
	"GodDns/Util"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strconv"
	"time"
	//"github.com/hedzr/cmdr/flag"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	_ "GodDns/Service" // register all services
)

var output = os.Stdout

var (
	optionRun      = flag.Bool("R", false, "run ddns")
	Time           = flag.Uint64("t", 0, "run ddns per time, only work with -R")
	optionAuto     = flag.Bool("A", false, "get ip address automatically, only work with -R")
	optionOverride = flag.Bool("O", false, "override the ip address, only work with -R -A")
	optionGenerate = flag.Bool("G", false, "generate a default config file")
	LogLevel       = flag.String("log", "Info", "log level: Error, Warn, Info, Debug, Trace")
	NoStdout       = flag.Bool("S", false, "no stdout")
	ApiName        = flag.String("api", "", "api name")
	retryAttempt   = flag.Int("retry", 3, "retry attempt times")
	// todo RPC :=  flag.Bool("RPC", false, "run ddns with RPC") run RPC server
	config = flag.String("config", "", "config file path")
)

// InitLog
// initialize the log file with fileMode and log level
// print information to output
// return a function to close the log file
// if error occurs, return error
func InitLog(filename string, filePerm os.FileMode, loglevel string, output io.Writer) (func(), error) {

	var level logrus.Level
	switch loglevel {
	case "Error":
		level = logrus.ErrorLevel
	case "Warn":
		level = logrus.WarnLevel
	case "Info":
		level = logrus.InfoLevel
	case "Debug":
		level = logrus.DebugLevel
	case "Trace":
		level = logrus.TraceLevel
	default:
		logrus.Error("invalid log level")
	}

	// output to log file
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePerm)
	if err != nil {
		return nil, err
	}

	cleanUp := func() {
		err := file.Close()
		if err != nil {
			logrus.Error("failed to close log file ", err)
		}
	}

	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: time.DateTime,
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			filename := path.Base(frame.File)
			return frame.Function, filename
		},
	})
	logrus.SetLevel(level)

	_, _ = fmt.Fprintf(output, "init log file at %s\n", filename)
	Log.To(logrus.StandardLogger(), file)

	if output != nil {
		if _, ok := output.(*os.File); !ok || output.(*os.File) != nil {
			// output is not *os.File(nil)
			logrus.AddHook(Log.NewLogrusOriginally2writer(output))
		}
	}

	_, err = file.Write([]byte(fmt.Sprintf("---------start at %s---------\n", time.Now().Format(time.DateTime))))
	if err != nil {
		return cleanUp, err
	}

	return cleanUp, nil
}

// -R -A -O -t=xxx -S -log=xxx -retry=xxx -config=xxx
// -R -Api=xxx -t=xxx -S -log=xxx -retry=xxx -config=xxx
// -G -log=xxx -S -config=xxx
// test:
// -G -A -O
// -R -A -Api=xxx
//
//

func CheckFlagValidation() {
	validateLogLevel()
	validateRetryAttempt()
	validateConfig()
}

func validateRunAndGenerate() {
	// run=true generate=true -> invalid
	if *optionRun && *optionGenerate {
		logrus.Fatal("invalid options")
	}

	// run=false auto=true -> invalid
	if !*optionRun && *optionAuto {
		logrus.Fatal("invalid options")
	}
	// auto = true, run must be true now

	// auto=false override=true -> invalid
	if !*optionAuto && *optionOverride {
		logrus.Fatal("invalid options")
	}

	// auto=true, api=xxx -> invalid
	if *optionAuto && *ApiName != "" {
		logrus.Fatal("invalid options")
	}

	// todo implement here

}

func validateLogLevel() {
	if *LogLevel != "Error" && *LogLevel != "Warn" && *LogLevel != "Info" && *LogLevel != "Debug" && *LogLevel != "Trace" {
		logrus.Fatal("invalid log level")
	}
}

func validateRetryAttempt() {
	if *retryAttempt < 0 {
		logrus.Fatal("invalid retry attempt times")
	}
}

func validateConfig() {
	if *config != "" {
		DDNS.UpdateConfigureLocation(*config)
	} else {
		location, err := DDNS.GetDefaultConfigurationLocation()
		if err != nil {
			logrus.Fatalf("error opening folder:%s", err)
		}
		DDNS.UpdateConfigureLocation(location)
	}
}

// command line parameters
func main() {

	/*
	   -R options (-t time(seconds) -A/-I -O)  run ddns per time -A:get ip addr automatically
	   -G generate a default config file
	   -A get ip address automatically, only work with -R
	   -Api=xxx get ip address from ipify, only work with -R
	   -O override the ip address, only work with -R -A
	   -log=Error/Warn/Info/Debug/Trace
	   -retry=uint
	*/

	// Devices should be the first ,followed by other services
	configFactoryList := DDNS.ConfigFactoryList

	length := len(os.Args)
	if length == 1 {
		os.Args = append(os.Args, "-h")
	}

	flag.Parse()
	CheckFlagValidation()

	if *NoStdout {
		output = nil
	}

	if a, err := InitLog("ddns.log", 0664, *LogLevel, output); err != nil {
		logrus.Error("log file init failed ", err)
		defer a()
	}

	defer logrus.Infof("exit")

	// -r
	if *optionRun {
		RunDDNS(configFactoryList)
	} else if *optionGenerate {
		GenerateConfigure(configFactoryList)
	}

}

//-----------------------------------------------------------------------------------------------------------------------------------------//

func RunDDNS(configs []DDNS.ConfigFactory) {
	logrus.Debugf("-R is set, run ddns")
	// run ddns here
	parameters, fileErr, errs := DDNS.ConfigureReader(DDNS.GetConfigureLocation(), configs...)
	if fileErr != nil {
		logrus.Errorf("error reading config: %s", fileErr.Error())
		return
	}

	if errs != nil {
		logrus.Errorf("error reading config: %s", errs.Error())
		_, _ = fmt.Fprintf(output, "error reading config: %s", errs.Error())
	}

	if len(parameters) <= 1 {
		logrus.Info("no service left to run")
		return
	}

	var requests []DDNS.Request
	// -Api
	if *ApiName != "" {
		RunGetFromApi(parameters, requests)
		return
	}

	// -A
	if *optionAuto {
		RunAuto(parameters, requests)
	} else {
		// -A is not set
		requests = GenerateRequests(parameters, requests)

		if *Time != 0 {
			RunPerTime(Time, requests)
		}

		d, err := DDNS.Find(parameters, Device.ServiceName)
		Parameters2Save := make([]DDNS.Parameters, 0)
		if err == nil {
			Parameters2Save = append(Parameters2Save, d)
		}

		ExecuteRequests(requests...)
		Parameters2Save = append(Parameters2Save, Requests2Parameters(requests)...)
		SaveFromParameters(Parameters2Save...)
	}

}

func RunGetFromApi(parameters []DDNS.Parameters, requests []DDNS.Request) {

	var api = Net.ApiMap.GetApi(*ApiName)

	logrus.Debugf("-I is set, get ip address from %s", *ApiName)
	ip4, err1 := api(4)
	if err1 != nil {
		logrus.Errorf("error getting ipv4 ,%s", err1)
	} else {
		logrus.Infof("ipv4 from %s: %s", *ApiName, ip4)
	}

	ip6, err2 := api(6)
	if err2 != nil {
		logrus.Errorf("error getting ipv4 ,%s", err2)
	} else {
		logrus.Infof("ipv6 from %s: %s", *ApiName, ip6)
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
						logrus.Errorf("unknown type %s", d.GetType())
					}
				}
			}
		}
	}

	requests = GenerateRequests(parameters, requests)

	if *Time != 0 {
		RunPerTime(Time, requests)
	}
	d, err := DDNS.Find(parameters, Device.ServiceName)
	Parameters2Save := make([]DDNS.Parameters, 0)
	if err == nil {
		Parameters2Save = append(Parameters2Save, d)
	}
	ExecuteRequests(requests...)
	Parameters2Save = append(Parameters2Save, Requests2Parameters(requests)...)
	SaveFromParameters(Parameters2Save...)

}

func RunAuto(parameters []DDNS.Parameters, requests []DDNS.Request) {
	logrus.Info("-A is set, get ip address automatically")
	// get ip addr automatically
	// the first parameter is Devices
	deviceInterface, err := DDNS.Find(parameters, Device.ServiceName)
	if err != nil {
		logrus.Fatalf("Section [Devices] not found, check configure at %s", DDNS.GetConfigureLocation())
	}
	GlobalDevice := deviceInterface.(Device.Device)

	/*------------------------------------------------------------------------------------------*/

	// -o
	if *optionOverride {
		RunOverride(parameters, GlobalDevice, requests)
		return
	}
	devices := GlobalDevice.GetDevices()

	ip4 := Util.Pair[string, string]{} // First is device, second is ip
	ip6 := Util.Pair[string, string]{} // First is device, second is ip

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
				logrus.Errorf("error getting ipv4 %s ,%s", device, err1)
			} else {
				logrus.Infof("ipv4 from %s: %s", device, ip4sTemp)
				ip4s = ip4sTemp
				ip4.Set(device, Net.DealWithIp(ip4s...))

			}
		}

		if ip6s == nil {
			ip6sTemp, err2Temp := Net.GetIpByType(device, Net.AAAA)
			if err2Temp != nil {
				err2 = errors.Join(err2, err2Temp)
				logrus.Errorf("error getting ipv6 %s ,%s", device, err2)
			} else {
				logrus.Infof("ipv6 from %s: %s", device, ip6sTemp)
				ip6s = ip6sTemp
				ip6.Set(device, Net.DealWithIp(ip6s...))
			}
		}

		if ip6s != nil && ip4s != nil {
			break
		}

	}

	set := func(parameter DDNS.Parameters) error {
		switch parameter.(DDNS.ServiceParameters).GetType() {
		case "4":
			parameter.(DDNS.ServiceParameters).SetValue(ip4.Second)
			return err1
		case "6":
			parameter.(DDNS.ServiceParameters).SetValue(ip6.Second)
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
				logrus.Errorf("error setting ip address: %s, skip service:%s", err.Error(), parameter.GetName())
				parameters = append(parameters[:i], parameters[i+1:]...) // ? may be bugs

			}
		}
	}

	requests = GenerateRequests(parameters, requests)

	if *Time != 0 {
		RunPerTime(Time, requests)
	}
	d, err := DDNS.Find(parameters, Device.ServiceName)
	Parameters2Save := make([]DDNS.Parameters, 0)
	if err == nil {
		Parameters2Save = append(Parameters2Save, d)
	}
	ExecuteRequests(requests...)
	Parameters2Save = append(Parameters2Save, Requests2Parameters(requests)...)
	SaveFromParameters(Parameters2Save...)

}

func RunOverride(parameters []DDNS.Parameters, GlobalDevice Device.Device, requests []DDNS.Request) {
	// override the ip address here
	// use the Key `Devices` and `Type` of the Service if exist
	logrus.Info("-O is set, override the ip address")
	var errCount uint16

	for i, parameter := range parameters {
		//skip the device parameter
		if parameter.GetName() != Device.ServiceName {
			// check if parameter implements DeviceOverridable interface
			if d, ok := parameter.(DDNS.DeviceOverridable); ok {
				logrus.Debugf("Parameter %s implements DeviceOverridable interface", parameter.GetName())

				var tempDeviceName string

				// if device is not set, use Type IP value of Global Devices
				if !d.IsDeviceSet() {
					err := set(GlobalDevice, parameter)
					if err != nil {
						logrus.Errorf("error setting ip address: %s, skip service:%s", err.Error(), parameter.GetName())
						errCount++
						parameters = append(parameters[:i], parameters[i+1:]...)
						continue
					}

					logrus.Warnf("Devices of %s is not set, use default value %s", parameter.GetName(), parameter.(DDNS.DeviceOverridable).GetIP())
					continue // skip
				}

				if !d.IsTypeSet() {
					errCount++
					parameters = append(parameters[:i], parameters[i+1:]...)
					logrus.Errorf("error setting ip address: unknown type, skip service:%s", parameter.GetName())
					continue
				}

				TypeInt, _ := strconv.Atoi(d.GetType())
				tempDeviceName = d.GetDevice()
				ips, err := Net.GetIpByType(tempDeviceName, uint8(TypeInt))
				if err != nil {
					errCount++
					parameters = append(parameters[:i], parameters[i+1:]...)
					logrus.Errorf("error getting ip address: %s, skip service:%s", err.Error(), parameter.GetName())
					continue
				}
				//

				logrus.Infof("override %s with %s", parameter.GetName(), ips[0])
				parameter.(DDNS.DeviceOverridable).SetValue(ips[0])
			} else {
				// Service is not DeviceOverridable, use ip got from Devices Section
				err := set(GlobalDevice, parameter)
				if err != nil {
					errCount++
					parameters = append(parameters[:i], parameters[i+1:]...)
					continue
				}
				logrus.Debugf("Parameter %s is not DeviceOverridable, use default value %s", parameter.GetName(), parameter.(DDNS.ServiceParameters).GetIP())
			}

		}
	} //loop ends

	logrus.Infof("finish overriding ip with %d error(s)", errCount)

	requests = GenerateRequests(parameters, requests)

	if *Time != 0 {
		RunPerTime(Time, requests)
	}
	d, err := DDNS.Find(parameters, Device.ServiceName)
	Parameters2Save := make([]DDNS.Parameters, 0)
	if err == nil {
		Parameters2Save = append(Parameters2Save, d)
	}
	ExecuteRequests(requests...)
	Parameters2Save = append(Parameters2Save, Requests2Parameters(requests)...)
	SaveFromParameters(Parameters2Save...)

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
					logrus.Errorf("error getting ipv4 %s ,%s", device, err)
				} else {
					logrus.Infof("ipv4 from %s: %s", device, ip4sTemp)
					ips = ip4sTemp
					ip.Set(device, Net.DealWithIp(ips...))
				}
			}
		}

	case "6":
		for _, device := range devices {
			if ips == nil {
				ip6sTemp, errTemp := Net.GetIpByType(device, Net.AAAA)
				if errTemp != nil {
					err = errors.Join(err, errTemp)
					logrus.Errorf("error getting ipv6 %s ,%s", device, err)
				} else {
					logrus.Infof("ipv6 from %s: %s", device, ip6sTemp)
					ips = ip6sTemp
					ip.Set(device, Net.DealWithIp(ips...))
				}
			}
		}

	default:
		return fmt.Errorf("unknown type %s", ParameterToSet.(DDNS.ServiceParameters).GetType())
	}

	if ip.First != "" && ip.Second != "" {
		ParameterToSet.(DDNS.ServiceParameters).SetValue(ip.Second)
		return nil
	} else {
		return err
	}

}

func GenerateConfigure(configFactoryList []DDNS.ConfigFactory) {
	if DDNS.IsConfigureExist() {
		logrus.Warnf("configure at %s already exist", DDNS.GetConfigureLocation())
		return
	}
	logrus.Debugf("start generating default configure")
	err := GenerateDefaultConfigure(configFactoryList...)
	if err != nil {
		logrus.Error(err.Error())
		return
	}
	logrus.Info("generate a default config file at", DDNS.GetConfigureLocation())

}

func ExecuteRequests(requests ...DDNS.Request) {
	logrus.Info("start executing requests")

	for _, request := range requests {
		logrus.Tracef("request: %s", request.GetName())
		err := DDNS.ExecuteRequest(request)
		if err != nil || request.Status().Success != DDNS.Success {
			logrus.Errorf("error executing request, %s", err.Error())
			Retry(request, *retryAttempt)
		}

		status := ""
		res := request.Status()
		if res.Success == DDNS.Success {
			status = "Success"
		} else if res.Success == DDNS.Failed {
			logrus.Errorf("error executing request, %s", err.Error())
			status = "Failed"
		} else if res.Success == DDNS.NotExecute {
			logrus.Fatal("request not executed")
		}

		logrus.Infof("name:%s, status:%s, msg:%s", res.Name, status, res.Msg)
	}

}

func Retry(request DDNS.Request, i int) {
	for j := 1; j <= i; j++ {
		logrus.Warnf("retrying %s %d time", request.GetName(), j)
		err := DDNS.ExecuteRequest(request)
		if err != nil {
			logrus.Errorf("error: %s", err.Error())
		} else {
			return
		}
	}
	logrus.Errorf("all retry failed, skip %s", request.GetName())
}

func GenerateRequests(parameters []DDNS.Parameters, requests []DDNS.Request) []DDNS.Request {
	logrus.Info("start generating requests")
	var errCount uint8 = 0
	for _, parameter := range parameters {
		if parameter.GetName() == Device.ServiceName {
			continue //skip
		}

		logrus.Infof("service: %s", parameter.GetName())
		request, err := parameter.(DDNS.ServiceParameters).ToRequest()
		if err != nil {
			errCount++
			logrus.Errorf("error generating request for %s:%s ", parameter.GetName(), err.Error())
			continue
		}
		requests = append(requests, request)
	}
	logrus.Infof("finish generating requests with %d error(s)", errCount)
	return requests
}

func GenerateDefaultConfigure(ConfigFactories ...DDNS.ConfigFactory) error {
	var infos []DDNS.ConfigStr
	var err error
	for _, factory := range ConfigFactories {
		info, errTemp := factory.Get().GenerateDefaultConfigInfo()
		logrus.Tracef("config info: \n%s", info)
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
	logrus.Info("write default configure to", DDNS.GetConfigureLocation())
	return nil
}

// RunPerTime run ddns per time
// todo fix
// Expect: run 1 --time-- run 2 --time-- run 3 ...
// Actual   : --time-- run 1 --time-- run 2 --time-- run 3 ...
func RunPerTime(Time *uint64, requests []DDNS.Request) {

	logrus.Infof("run ddns per %d seconds", *Time)

	c := cron.New()
	for _, request := range requests {
		_, err := c.AddJob(fmt.Sprintf("@every %ds", *Time), cron.NewChain(cron.DelayIfStillRunning(cron.DefaultLogger)).Then(request))
		if err != nil {
			logrus.Errorf("error adding job %s: %s", request.GetName(), err.Error())
		}
	}

	c.Start()
	select {}

}

func SaveFromParameters(parameters ...DDNS.Parameters) {
	err := DDNS.SaveConfig(DDNS.GetConfigureLocation(), os.O_CREATE|os.O_TRUNC, parameters...)
	if err != nil {
		logrus.Errorf("error saving config: %s", err.Error())
	}
}

func Requests2Parameters(requests []DDNS.Request) []DDNS.Parameters {
	Parameters2Save := make([]DDNS.Parameters, 0)
	for _, request := range requests {
		if request.Status().Success == DDNS.Success {
			Parameters2Save = append(Parameters2Save, request.ToParameters())
		}
	}
	return Parameters2Save
}
