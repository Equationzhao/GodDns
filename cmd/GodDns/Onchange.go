package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"GodDns/Device"
	"GodDns/core"
	"GodDns/log"
	log "GodDns/log"
	"GodDns/net"
	"GodDns/util/collections"
	"github.com/robfig/cron/v3"
)

type BindDeviceService map[string][]*core.Parameters // refactor?

var MainBinder = make(BindDeviceService, 20)

func (b *BindDeviceService) Bind(Device string, Service *core.Parameters) (ok bool) {
	(*b)[Device] = append((*b)[Device], Service)
	return true
}

func OnChange(ps []*core.Parameters, GlobalDevice *Device.Device) {
	defer core.CatchPanic(output)

	if GlobalDevice == nil {
		panic("no global device")
	}
	err := ModeController(ps, GlobalDevice)
	switch err {
	case nil:
		break
	case NoRequestErr{}:
		log.Error(err.Error())
		return
	default:
		log.Error("error running ddns: ", log.String("error", err.Error()))
	}

	StartIpChangeDaemon(ps)
}

func StartIpChangeDaemon(ps []*core.Parameters) {
	type result int
	const (
		done result = iota
		unaffected
		errorOccur
		timeout
	)
	cornLogfile, err := os.OpenFile("cron.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		log.Debug(err)
	}

	logger := log.NewLogger(cornLogfile)
	logger = logger.WithGroup("cron-oc:")
	c := cron.New(cron.WithChain(cron.Recover(logger),
		cron.DelayIfStillRunning(logger)),
		cron.WithLogger(cron.VerbosePrintfLogger(logger)))
	wg := sync.WaitGroup{}

	if TimesLimitation == 0 {
		TimesLimitation = MAXTIMES
	}

	save := make(chan struct{}, 10)

	scanGap, ok := core.UniversalConfig[core.OcScanTime].(time.Duration)
	if !ok {
		scanGap, _ = time.ParseDuration("10s")
	}

	sendSignal := func(OldIp collections.Pair[string, string], handledIp []string, changedSignal []chan string,
		services []*core.Parameters, serviceResult chan result, res *[4]int, timeout result, d string, done result,
		unaffected result, errorOccur result, isA bool,
	) {
		var typeToHandle string
		if isA {
			typeToHandle = "ipv4"
			*OldIp.First = handledIp[0]
		} else {
			typeToHandle = "ipv6"
			*OldIp.Second = handledIp[0]
		}

		for _, s := range changedSignal {
			s <- handledIp[0]
		}

		timeoutChan := time.After(30 * time.Second)
		total := len(services)
		countDone := make(chan struct{}, 1)
		_ = core.MainGoroutinePool.Submit(func() {
			defer func() {
				countDone <- struct{}{}
			}()

			for {
				select {
				case status := <-serviceResult:
					total--
					res[status]++
					if total == 0 {
						return
					}
				case <-timeoutChan:
					Log.Info("timeout")
					return
				}
			}
		})
		<-countDone
		res[timeout] += total

		Log.Info(fmt.Sprintf("result for %s.%s", d, typeToHandle),
			Log.Int("done", res[done]).String(),
			Log.Int("unaffected", res[unaffected]).String(),
			Log.Int("error", res[errorOccur]).String(),
			Log.Int("timeout", res[timeout]).String())
	}

	for d, services := range MainBinder {
		changedSignal := make([]chan string, len(services))
		for i := range changedSignal {
			changedSignal[i] = make(chan string, 1)
		}
		serviceResult := make(chan result, len(services))
		wg.Add(len(services))
		_, _ = c.AddFunc(fmt.Sprintf("@every %s", scanGap.String()), func() {
			defer core.CatchPanic(output)
			Log.Info("checking ip change ", Log.String("device", d).String())
			res := [4]int{0, 0, 0, 0}
			{
				var handledIp []string
				ip, err := Net.GetIpByType(d, Net.A)
				if err != nil {
					Log.Error("error getting ip", Log.String("error", err.Error()).String())
					goto AAAA
				}
				handledIp, err = Net.HandleIp(ip)
				if err != nil {
					Log.Error("error handle ip: ", Log.String("error", err.Error()).String())
					goto AAAA
				}
				if len(handledIp) == 0 {
					Log.Info("no ip left, please check ip handler or network", "device", d)
					goto AAAA
				}

				if OldIp, ok := Device2Ips[d]; ok {
					if OldIp.First == nil {
						goto AAAA
					}
					if OldIp.GetFirst() == handledIp[0] {
						Log.Info("ip not changed ", Log.String("ip", OldIp.GetFirst()).String())
						goto AAAA
					}
					Log.Info("ip changed", Log.String("old", OldIp.GetFirst()).String(), Log.String("new", handledIp[0]).String())
					sendSignal(OldIp, handledIp, changedSignal, services, serviceResult, &res,
						timeout, d, done, unaffected, errorOccur, true)
				}
			}
		AAAA:
			{
				ip, err := Net.GetIpByType(d, Net.AAAA)
				if err != nil {
					Log.Error("error getting ip ", Log.String("error", err.Error()).String())
				}
				handledIp, err := Net.HandleIp(ip)
				if err != nil {
					Log.Error("error handle ip ", Log.String("error", err.Error()).String())
				}
				if len(handledIp) == 0 {
					Log.Info("no ip left, please check ip handler or network ", Log.String("device", d).String())
					goto END
				}

				if OldIp, ok := Device2Ips[d]; ok {
					if OldIp.Second == nil {
						goto END
					}
					if OldIp.GetSecond() == handledIp[0] {
						Log.Info("ip not changed ", Log.String("ip", OldIp.GetSecond()).String())
						goto END
					}
					sendSignal(OldIp, handledIp, changedSignal, services, serviceResult, &res,
						timeout, d, done, unaffected, errorOccur, false)
				}
			}
		END:
			if res[done] != 0 {
				save <- struct{}{}
			}
		})

		for i, service := range services {
			times := TimesLimitation
			service := service
			i := i
			_ = core.MainGoroutinePool.Submit(func() {
				defer core.CatchPanic(output)
				defer wg.Done()
				for j := 0; j < times; j++ {
					newIp := <-changedSignal[i]
					_type := strconv.Itoa(int(Net.WhichType(newIp)))
					if (*service).(core.Service).GetType() == _type {
						(*service).(core.Service).SetValue(newIp)
						request, err := (*service).(core.Service).ToRequest()
						if err != nil {
							_, _ = Log.ErrPP.Fprintln(output, err.Error())
							continue
						}

						if proxyEnable {
							_, _ = Log.InfoPP.Fprintln(output, "try to request through proxy")
							err := request.(core.ThroughProxy).RequestThroughProxy()
							if err != nil {
								_, _ = Log.ErrPP.Fprintln(output, err.Error())
								Retry(request, retryAttempt)
							}
						} else {
							_, _ = Log.InfoPP.Fprintln(output, "make request")
							err := request.MakeRequest()
							if err != nil {
								_, _ = Log.ErrPP.Fprintln(output, err.Error())
								Retry(request, retryAttempt)
							}
						}

						switch request.Status().Status {
						case core.Success:
							serviceResult <- done
						case core.Failed:
							serviceResult <- errorOccur
						case core.NotExecute:
							serviceResult <- errorOccur
							// todo timeout
						}
						Display(request, output)
						*service = request.ToParameters()
					} else {
						serviceResult <- unaffected
						_, _ = Log.InfoPP.Fprintln(output, "ip type not match")
						j--
						continue
					}
				}
			})
		}
	}

	c.Start()

	_ = core.MainGoroutinePool.Submit(func() {
		for {
			<-save
			log.Debug("save from cron")

			log.Debug("services: ", Log.Any("all", ps).String())
			toSave := make([]core.Parameters, 0, len(ps))
			for _, p := range ps {
				toSave = append(toSave, *p)
			}
			err := SaveFromParameters(toSave...)
			if err != nil {
				_, _ = Log.ErrPP.Fprintln(output, err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	})

	wg.Wait()

	_, _ = Log.DebugPP.Fprintln(output, "all services finished")
}
