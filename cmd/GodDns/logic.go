package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	// GodDns
	"GodDns/cmd/GodDns/tui"
	"GodDns/core"
	log "GodDns/log"
	"GodDns/netinterface"
	"GodDns/netutil"
	"GodDns/util"
	"GodDns/util/collections"

	tea "github.com/charmbracelet/bubbletea"
	// 3rd party
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/robfig/cron/v3"
)

// -----------------------------------------------------------------------------------------------------------------------------------------//

func ModeController(ps []*core.Parameters, GlobalDevice *netinterface.Device) error {
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

func RunDDNS(parameters []*core.Parameters) error {
	log.Debugf("run ddns")
	// run ddns here

	// todo check ip and type
	// if not match, error

	return GenerateExecuteSave(parameters)
}

// device name -> [ipv4, ipv6]
type d2i map[string]collections.Pair[string, string]

var Device2Ips = make(d2i, 20)

func (d *d2i) Add(device string, ip string, t netutil.Type) {
	switch t {
	case netutil.A:
		(map[string]collections.Pair[string, string])(*d)[device] = collections.Pair[string, string]{
			First:  &ip,
			Second: (map[string]collections.Pair[string, string])(*d)[device].Second,
		}
	case netutil.AAAA:
		(map[string]collections.Pair[string, string])(*d)[device] = collections.Pair[string, string]{
			First:  (map[string]collections.Pair[string, string])(*d)[device].First,
			Second: &ip,
		}
	default:
		panic("invalid type")
	}
}

func ReadConfig(configs []core.ConfigFactory) ([]core.Parameters, error) {
	parameters, fileErr, configErrs := core.ConfigureReader(core.GetConfigureLocation(), configs...)
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
func RunGetFromApi(parameters []*core.Parameters) error {
	// if api is api name , get api from map
	// if api is url , try to make request to url  http://example.com/api?ip=4 or http://example.com/api?ip=6

	var api netutil.Api
	api, err := netutil.ApiMap.GetApi(ApiName)
	if err != nil {
		log.Errorf("error getting api %s, %s", ApiName, err)
		// todo suggestion "do you mean xxx"
		return errors.New("") // return error with no message to avoid print error message again
	}

	log.Debugf("-I is set, get ip address from %s", ApiName)

	ip4Done, ip6Done := make(chan bool, 1), make(chan bool, 1)
	ip4, _ip4 := "", ""
	ip6, _ip6 := "", ""

	var err1 error
	_ = core.MainGoroutinePool.Submit(func() {
		_ip4, err1 = api.Get(netutil.A)
		if err1 != nil {
			ip4Done <- false
		} else {
			ip4 = _ip4
			ip4Done <- true
		}
		close(ip4Done)
	})

	var err2 error
	_ = core.MainGoroutinePool.Submit(func() {
		_ip6, err2 = api.Get(netutil.AAAA)
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
		if (*parameter).GetName() != netinterface.ServiceName {
			if d, ok := (*parameter).(core.Service); ok {
				if d.IsTypeSet() {
					if netutil.TypeEqual(d.GetType(), netutil.A) {
						d.SetValue(ip4)
					} else if netutil.TypeEqual(d.GetType(), netutil.AAAA) {
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

func RunAuto(GlobalDevice netinterface.Device, parameters []*core.Parameters) error {
	log.Info("get ip address automatically")
	// get ip addr automatically

	/*------------------------------------------------------------------------------------------*/

	devices := GlobalDevice.GetDevices()

	// First is device, second is ip
	ip4 := collections.Pair[string, string]{}
	// First is device, second is ip
	ip6 := collections.Pair[string, string]{}

	var err1, err2 error

	var ip4s []string = nil
	var ip6s []string = nil

	for _, device := range devices {
		if ip4s == nil {
			ip4sTemp, err1Temp := netutil.GetIpByType(device, netutil.A)
			if err1Temp != nil {
				err1 = errors.Join(err1, err1Temp)
				log.Errorf("error getting ipv4 %s ,%s", device, err1)
			} else {
				log.Infof("ipv4 from %s: %s", device, ip4sTemp)
				ip4s, err1 = netutil.HandleIp(ip4sTemp)
				device := device
				ip4.Move(&device, &ip4s[0])
			}
		}

		if ip6s == nil {
			ip6sTemp, err2Temp := netutil.GetIpByType(device, netutil.AAAA)
			if err2Temp != nil {
				err2 = errors.Join(err2, err2Temp)
				log.Errorf("error getting ipv6 %s ,%s", device, err2)
			} else {
				log.Infof("ipv6 from %s: %s", device, ip6sTemp)
				ip6s, err2 = netutil.HandleIp(ip6sTemp)
				device := device
				ip6.Move(&device, &ip6s[0])
			}
		}

		if ip6s != nil && ip4s != nil {
			break
		}
	}

	o4 := sync.Once{}
	o6 := sync.Once{}
	set := func(parameter *core.Parameters) error {
		switch (*parameter).(core.Service).GetType() {
		case "4":
			if ip4.First != nil {
				MainBinder.Bind(ip4.GetFirst(), parameter)
				(*parameter).(core.Service).SetValue(ip4.GetSecond())
				o4.Do(
					func() {
						Device2Ips.Add(ip4.GetFirst(), ip4.GetSecond(), netutil.A)
					})
			}
			return err1
		case "6":
			if ip6.First != nil {
				MainBinder.Bind(ip6.GetFirst(), parameter)
				(*parameter).(core.Service).SetValue(ip6.GetSecond())
				o6.Do(
					func() {
						Device2Ips.Add(ip6.GetFirst(), ip6.GetSecond(), netutil.AAAA)
					})
			}
			return err2
		default:
			return fmt.Errorf("unknown type %s", (*parameter).(core.Service).GetType())
		}
	}

	for _, parameter := range parameters {
		if _, ok := (*parameter).(core.Service); ok {
			// if parameter implements DeviceOverridable interface, set the ip address
			if err := set(parameter); err != nil {
				log.Errorf("error setting ip address: %s, skip service:%s", err.Error(), (*parameter).GetName())
			}
		}
	}

	return GenerateExecuteSave(parameters)
}

// GetGlobalDevice get the global device
// if not found, fatal
func GetGlobalDevice(parameters []*core.Parameters) (netinterface.Device, error) {
	for _, d := range parameters {
		if (*d).GetName() == netinterface.ServiceName {
			GlobalDevice, ok := (*d).(netinterface.Device)
			if !ok {
				panic("the parameter is not a Device but has a name equal to Device.ServiceName")
			}
			return GlobalDevice, nil
		}
	}

	log.Errorf("Section [Devices] not found, check configuration at %s", core.GetConfigureLocation())
	return netinterface.Device{},
		fmt.Errorf("section [Devices] not found, check configuration at %s", core.GetConfigureLocation())
}

func RunOverride(GlobalDevice netinterface.Device, parameters []*core.Parameters) error {
	// override the ip address here
	// use the Key `Devices` and `Type` of the Service if exist
	log.Info("-O is set, override the ip address")
	var errCount uint16

	for _, parameter := range parameters {
		// skip the device parameter
		if (*parameter).GetName() != netinterface.ServiceName {
			// check if parameter implements DeviceOverridable interface
			if d, ok := (*parameter).(core.DeviceOverridable); ok {
				log.Debugf("Parameter %s implements DeviceOverridable interface", d.GetName())

				var tempDeviceName string

				// if device is not set, use Type IP value of Global Devices
				if !d.IsDeviceSet() {
					err := set(GlobalDevice, parameter)
					d = (*parameter).(core.DeviceOverridable)
					if err != nil {
						log.Errorf("error setting ip address: %s, skip overriding service:%s", err.Error(), d.GetName())
						errCount++

						continue
					}
					log.Warnf("Devices of %s is not set, use default value %s", d.GetName(), d.GetIP())
					continue // skip
				}

				if !d.IsTypeSet() {
					errCount++

					log.Errorf("error setting ip address: unknown type, skip service:%s", d.GetName())
					continue
				}

				TypeInt, _ := strconv.Atoi(d.GetType())
				tempDeviceName = d.GetDevice()
				ips, err := netutil.GetIpByType(tempDeviceName, uint8(TypeInt))
				if err != nil {
					errCount++

					log.Errorf("error getting ip address: %s, skip service:%s", err.Error(), d.GetName())
					continue
				}

				log.Infof("override %s with %s", d.GetName(), ips[0])
				d.SetValue(ips[0])

				MainBinder.Bind(tempDeviceName, parameter)
			} else {
				// Service is not DeviceOverridable, use ip got from Devices Section
				err := set(GlobalDevice, parameter)
				if err != nil {
					errCount++
					log.Errorf("error setting ip address: %s, use default value:%s", err.Error(), (*parameter).(core.Service).GetIP())
					continue
				}
				log.Debugf("Parameter %s is not DeviceOverridable, use default value %s", (*parameter).GetName(), (*parameter).(core.Service).GetIP())
			}
		}
	} // loop ends

	log.Infof("finish overriding ip with %d error(s)", errCount)

	return GenerateExecuteSave(parameters)
}

func set(GlobalDevice netinterface.Device, ParameterToSet *core.Parameters) error {
	Type := (*ParameterToSet).(core.Service).GetType() // Type is "4" or "6" or ""

	ip := collections.Pair[string, string]{} // First is device, second is ip

	var err error
	devices := GlobalDevice.GetDevices()
	var ips []string = nil

	// if failed to get ip, then try to get ip of next device in the list, else break
	switch Type {
	case "4":
		for _, device := range devices {
			if ips == nil {
				ip4sTemp, errTemp := netutil.GetIpByType(device, netutil.A)
				if errTemp != nil {
					err = errors.Join(err, errTemp)
					log.Errorf("error getting ipv4 %s ,%s", device, err)
				} else {
					log.Infof("ipv4 from %s: %s", device, ip4sTemp)
					ips, errTemp = netutil.HandleIp(ip4sTemp, netutil.RemoveLoopback)
					if errTemp != nil {
						err = errors.Join(err, errTemp)
						log.Errorf("error handling ipv4 %s ,%s", device, err)
					}
					ip.Move(&device, &ips[0])
				}
			}
		}

	case "6":
		for _, device := range devices {
			if ips == nil {
				ip6sTemp, errTemp := netutil.GetIpByType(device, netutil.AAAA)
				if errTemp != nil {
					err = errors.Join(err, errTemp)
					log.Errorf("error getting ipv6 %s ,%s", device, err)
				} else {
					log.Infof("ipv6 from %s: %s", device, ip6sTemp)
					ips, errTemp = netutil.HandleIp(ip6sTemp, netutil.RemoveLoopback)
					if errTemp != nil {
						err = errors.Join(err, errTemp)
						log.Errorf("error handling ipv4 %s ,%s", device, err)
					}
					ip.Move(&device, &ips[0])
				}
			}
		}

	default:
		return fmt.Errorf("unknown type %s", (*ParameterToSet).(core.Service).GetType())
	}

	if ip.First != nil && ip.Second != nil {
		MainBinder.Bind(ip.GetFirst(), ParameterToSet)
		(*ParameterToSet).(core.Service).SetValue(ip.GetSecond())
		return nil
	} else {
		return err
	}
}

type NoRequestErr struct{}

func (n NoRequestErr) Error() string {
	return "no request generated"
}

func GenerateExecuteSave(parameters []*core.Parameters) error {
	requests := GenerateRequests(parameters)

	if requests == nil {
		return NoRequestErr{}
	}

	Parameters2Save := make([]core.Parameters, 0, len(parameters))
	for _, p := range parameters {
		if (*p).GetName() == netinterface.ServiceName {
			Parameters2Save = append(Parameters2Save, *p)
			break
		}
	}

	ExecuteRequests(requests...)
	DisplayAll(output, requests...)
	for _, request := range requests {
		// update info from request.parameters
		r2p := request.ToParameters()
		Parameters2Save = append(Parameters2Save, r2p)
	}

	for i := range parameters {
		*parameters[i] = Parameters2Save[i]
	}

	return SaveFromParameters(Parameters2Save...)
}

func Display(request core.Request, output io.Writer) {
	switch {
	case tab:
		PrintInTable(request, output)
	case md:
		PrintMD(request, output)
	default:
		PrintDefault(request, output)
	}
}

func DisplayAll(output io.Writer, requests ...core.Request) {
	for _, request := range requests {
		Display(request, output)
		_, _ = fmt.Fprintln(output)
	}
}

func PrintDefault(request core.Request, output io.Writer) {
	serviceInfo := request.Status().MG.GetMsgOf(core.Info)
	serviceErr := request.Status().MG.GetMsgOf(core.Error)
	serviceWarn := request.Status().MG.GetMsgOf(core.Warn)

	switch {
	case len(serviceErr) > 0:
		_, _ = log.ErrPP.Fprintln(output,
			fmt.Sprint("error executing request ", request.GetName(), " at ", request.Target()))
	case len(serviceWarn) > 0:
		_, _ = log.WarnPP.Fprintln(output,
			fmt.Sprint("warning executing request ", request.GetName(), " at ", request.Target()))
	default:
		_, _ = log.InfoPP.Fprintln(output,
			fmt.Sprint("displaying message from Service ", request.GetName(), " at ", request.Target()))
	}

	if len(serviceInfo) > 0 {
		for _, i := range serviceInfo {
			_, _ = log.SuccessPP.Fprintln(output, i)
		}
	}

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

func PrintInTable(request core.Request, output io.Writer) {
	t := GetTableObj(request)
	t.SetOutputMirror(output)
	t.Render()
}

func PrintMD(request core.Request, output io.Writer) {
	var content *strings.Builder
	if util.StrBuilderPoolEnable {
		content = util.StrBuilderPool.Get().(*strings.Builder)
		defer content.Reset()
		defer util.StrBuilderPool.Put(content)
	} else {
		content = &strings.Builder{}
	}
	defer util.StrBuilderPool.Put(content)
	content.WriteString("# ")
	content.WriteString(request.GetName())
	content.WriteString(" at ")
	content.WriteString(request.Target())
	content.WriteByte('\n')

	infoMsg := request.Status().MG.GetMsgOf(core.Info)
	if len(infoMsg) != 0 {
		content.WriteString("## 🚩 Info\n")
		content.WriteString(strings.Join(request.Status().MG.GetMsgOf(core.Info), "\n\n"))
	}

	errorsMsg := request.Status().MG.GetMsgOf(core.Error)
	if len(errorsMsg) != 0 {
		content.WriteString("\n## ❌  Error\n")
		e := request.Status().MG.GetMsgOf(core.Error)
		for _, i := range e {
			content.WriteString(i)
			content.WriteString("\n\n")
		}
	}

	warnMsg := request.Status().MG.GetMsgOf(core.Warn)
	if len(warnMsg) != 0 {
		content.WriteString("\n## ⚠️  Warn\n")
		e := request.Status().MG.GetMsgOf(core.Warn)
		for _, i := range e {
			content.WriteString(i)
			content.WriteString("\n\n")
		}
	}

	out, err := core.GetMDRenderer().Render(content.String())
	if err != nil {
		panic(err)
	}
	_, _ = log.InfoPP.Fprintln(output, out)
}

func GetTableObj(request core.Request) table.Writer {
	t := table.NewWriter()
	t.SetTitle(request.GetName())
	status := "OK"
	if request.Status().Status != core.Success {
		status = "Fail"
	}
	TitleColor := text.Colors{text.FgGreen}
	header := table.Row{"Service", "Status", "Target", "IP"}
	content := table.Row{
		request.GetName(),
		status,
		request.Target(),
		request.ToParameters().GetIP(),
	}

	infoMsg := request.Status().MG.GetMsgOf(core.Info)
	if len(infoMsg) != 0 {
		TitleColor = text.Colors{text.FgRed}
		header = append(header, "Info")
		content = append(content, strings.Join(request.Status().MG.GetMsgOf(core.Info), "\n"))
	}

	errorsMsg := request.Status().MG.GetMsgOf(core.Error)
	if len(errorsMsg) != 0 {
		TitleColor = text.Colors{text.FgRed}
		header = append(header, "Error")
		content = append(content, strings.Join(request.Status().MG.GetMsgOf(core.Error), "\n"))
	}

	warnMsg := request.Status().MG.GetMsgOf(core.Warn)
	if len(warnMsg) != 0 {
		header = append(header, "Warn")
		content = append(content, strings.Join(request.Status().MG.GetMsgOf(core.Warn), "\n"))
	}

	t.AppendHeader(header)
	t.AppendRow(content)

	t.SetStyle(
		table.Style{
			Name:  "result display",
			Box:   table.StyleBoxBold,
			Color: table.ColorOptionsDark,
			Format: table.FormatOptions{
				Header: text.FormatTitle,
				Row:    text.FormatLower,
			},
			Options: table.Options{
				DoNotColorBordersAndSeparators: true,
				DrawBorder:                     true,
				SeparateColumns:                true,
				SeparateFooter:                 false,
				SeparateHeader:                 false,
				SeparateRows:                   false,
			},
			Title: table.TitleOptions{
				Align:  text.AlignCenter,
				Colors: TitleColor,
				Format: text.FormatTitle,
			},
		})
	return t
}

func GenerateConfigure(configFactoryList []core.ConfigFactory) error {
	if core.IsConfigExist(core.GetConfigureLocation()) {
		log.Warnf("configure at %s already exist", core.GetConfigureLocation())
		return errors.New("configure already exist")
	}
	log.Debugf("start generating default configure")
	err := GenerateDefaultConfigure(configFactoryList...)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Infof("generate a default config file at %s", core.GetConfigureLocation())
	return nil
}

func ExecuteRequests(requests ...core.Request) {
	log.Info("start executing requests")
	var wg sync.WaitGroup

	deal := func(err error, request core.Request) {
		if err != nil || (request).Status().Status != core.Success {
			log.ErrorRaw(fmt.Sprintf("error executing request, %v", err))
			Retry(request, retryAttempt)
		}

		var status string
		res := (request).Status()
		switch res.Status {
		case core.Success:
			status = "Success"
			log.InfoRaw(fmt.Sprintf("name:%s, status:%s  msg:%s", res.Name, status, res.MG))
		case core.Failed:
			errMsg := fmt.Sprintf("error executing request, %v", err)
			log.ErrorRaw(errMsg)
			status = "Failed"
			log.InfoRaw(fmt.Sprintf("name:%s, status:%s, msg:%s", res.Name, status, res.MG))
			if retryAttempt != 0 {
				log.ErrorRaw(fmt.Sprintf("all retry failed, skip %s:%s", (request).GetName(), (request).Target()))
			}
		case core.NotExecute:
			log.Fatal("request not executed")
		}
	}

	msgSpinner := make(chan struct{})
	if output != io.Discard {
		_ = core.MainGoroutinePool.Submit(func() {
			tui.ShowSpinner(
				func() tea.Msg {
					<-msgSpinner
					return struct{}{}
				},
			)
		})
	}

	if proxyEnable {
		for _, request := range requests {
			request := request
			wg.Add(1)

			_ = core.MainGoroutinePool.Submit(func() {
				var err error
				defer wg.Done()
				log.Tracef("request: %s", request.GetName())
				throughProxy, ok := request.(core.ThroughProxy)
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
			_ = core.MainGoroutinePool.Submit(func() {
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
	if output != io.Discard {
		msgSpinner <- struct{}{}
		time.Sleep(time.Millisecond * 100) // wait for spinner stop
	}
	log.Info("all requests finished")
}

func Retry(request core.Request, i uint8) {
	for j := uint8(1); j <= i; j++ {
		errMsg := fmt.Sprintf("retrying %s:%s, attempt %d", request.GetName(), request.Target(), j)
		log.WarnRaw(errMsg)
		request.Status().MG.AddError(fmt.Sprintf("retrying %s:%s, attempt %d", request.GetName(), request.Target(), j))

		if proxyEnable {
			throughProxy, ok := request.(core.ThroughProxy)
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

func GenerateRequests(parameters []*core.Parameters) []core.Request {
	log.Info("start generating requests")
	var errCount uint8 = 0
	requests := make([]core.Request, 0, len(parameters))
	for _, parameter := range parameters {
		if (*parameter).GetName() == netinterface.ServiceName {
			continue // skip
		}

		log.Infof("service: %s", (*parameter).GetName())
		request, err := (*parameter).(core.Service).ToRequest()
		if err != nil {
			errCount++
			log.Errorf("error generating request for %s:%s ", (*parameter).GetName(), err.Error())
			continue
		}
		requests = append(requests, request)
	}

	if len(requests) != 0 {
		log.Infof("finish generating requests with %d error(s)", errCount)
		return requests
	} else {
		// no request generated
		return nil
	}
}

func GenerateDefaultConfigure(ConfigFactories ...core.ConfigFactory) error {
	infos := make([]core.ConfigStr, 0, len(ConfigFactories))
	var err error
	for _, factory := range ConfigFactories {
		info, errTemp := factory.Get().GenerateDefaultConfigInfo()
		log.Tracef("config info: \n%s", info)
		if errTemp != nil {
			err = errors.Join(err, errTemp)
		}
		infos = append(infos, info)
	}
	errTemp := core.ConfigureWriter(core.GetConfigureLocation(), os.O_CREATE|os.O_WRONLY, infos...)
	err = errors.Join(err, errTemp)
	if err != nil {
		return err
	}
	log.Info("write default configure to ", core.GetConfigureLocation())
	return nil
}

// RunPerTime run ddns per time
func RunPerTime(Time uint64, GlobalDevice *netinterface.Device, parameters []*core.Parameters) {
	log.Infof("run ddns per %d seconds", Time)

	cornLogfile, err := os.OpenFile("cron.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
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
	_, err = c.AddJob(fmt.Sprintf("@every %ds", Time),
		cron.NewChain(cron.Recover(logger),
			cron.DelayIfStillRunning(cron.DefaultLogger)).
			Then(newServiceCronJob))

	if err != nil {
		log.Errorf("error adding job : %s", err.Error())
	}

	c.Start()
	wg.Wait()
	log.Info("all jobs finished", log.Int("total execution time", TimesLimitation).String())
}

// SaveConfig save parameters to file with flag
func SaveConfig(FileName string, flag int, parameters ...core.Parameters) error {
	var err error
	n := make(map[string]uint)
	ConfigStrings := make([]core.ConfigStr, 0, len(parameters))
	for _, parameter := range parameters {
		var no uint
		if parameter.GetName() == netinterface.ServiceName {
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
	err = errors.Join(err, core.ConfigureWriter(FileName, flag, ConfigStrings...))
	return err
}

func SaveFromParameters(parameters ...core.Parameters) error {
	log.Debug(parameters)
	// todo Merge Parameters that differ only by subdomain
	err := SaveConfig(core.GetConfigureLocation(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, parameters...)
	if err != nil {
		log.Errorf("error saving config: %s", err.Error())
		return fmt.Errorf("error saving config: %w", err)
	}
	log.Infof("save config to %s", core.GetConfigureLocation())
	return nil
}

func CheckVersionUpgrade(msg chan<- string) {
	// start checking version upgrade
	hasUpgrades, v, url, err := core.CheckUpdate()
	defer close(msg)
	defer func() {
		log.Tracef("check version upgrade finished")
	}()
	if err != nil {
		if errors.Is(err, core.NoCompatibleVersionError) {
			// "no suitable version"
			if hasUpgrades {
				msg <- fmt.Sprintf("new version %s is available", v.Info())
				msg <- fmt.Sprintf(`no compatible release for your operating system, 
							consider building from source:%s `, core.RepoURLs())
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
