package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"GodDns/Device"
	"GodDns/Log"
	log "GodDns/Log"
	"GodDns/Net"
	"GodDns/core"

	"github.com/robfig/cron/v3"
)

type BindDeviceService map[string][]*core.Service

var MainBinder = make(BindDeviceService, 20)

func (b *BindDeviceService) Bind(Device string, Service *core.Service) (ok bool) {
	(*b)[Device] = append((*b)[Device], Service)
	return true
}

func OnChange(ps *[]core.Parameters, GlobalDevice *Device.Device) {
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

func StartIpChangeDaemon(ps *[]core.Parameters) {
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

	save := make(chan struct{})
	_ = core.MainGoroutinePool.Submit(func() {
		for {
			<-save
			err := SaveFromParameters(*ps...)
			if err != nil {
				_, _ = Log.ErrPP.Fprintln(output, err.Error())
			}
		}
	})

	for d, services := range MainBinder {
		changedSignal := make([]chan string, len(services))
		for i := range changedSignal {
			changedSignal[i] = make(chan string, 1)
		}
		serviceResult := make(chan result, len(services))
		wg.Add(len(services))
		_, _ = c.AddFunc("@every 10s", func() {
			defer core.CatchPanic(output)
			Log.Info("checking ip change for device", Log.String("device", d).String())
			res := [4]int{0, 0, 0, 0}
			for i := 0; i < 2; i++ {
				var t Net.Type
				switch i {
				case 0:
					t = Net.A
					ip, err := Net.GetIpByType(d, t)
					if err != nil {
						Log.Error("error getting ip", Log.String("error", err.Error()).String())
						continue
					}
					handledIp, err := Net.HandleIp(ip)
					if err != nil {
						Log.Error("error handle ip: ", Log.String("error", err.Error()).String())
						continue
					}
					if len(handledIp) == 0 {
						Log.Info("no ip left, please check ip handler or network", "device", d)
						continue
					}

					if OldIp, ok := Device2Ips[d]; ok {
						if OldIp.First == nil {
							continue
						}
						if OldIp.GetFirst() == handledIp[0] {
							Log.Info("ip not changed", "ip", OldIp.GetFirst())
							continue
						}
						Log.Info("ip changed", Log.String("old", OldIp.GetFirst()).String(), Log.String("new", handledIp[0]).String())
						*OldIp.First = handledIp[0]
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
						Log.Info(fmt.Sprintf("result for %s.%s", d, "ipv4"),
							Log.Int("done", res[done]).String(),
							Log.Int("unaffected", res[unaffected]).String(),
							Log.Int("error", res[errorOccur]).String(),
							Log.Int("timeout", res[timeout]).String())
					}
				case 1:
					t = Net.AAAA
					ip, err := Net.GetIpByType(d, t)
					if err != nil {
						Log.Error("error getting ip", Log.String("error", err.Error()).String())
						continue
					}
					handledIp, err := Net.HandleIp(ip)
					if err != nil {
						Log.Error("error handle ip: ", Log.String("error", err.Error()).String())
						continue
					}
					if len(handledIp) == 0 {
						Log.Info("no ip left, please check ip handler or network", "device", d)
						continue
					}

					if OldIp, ok := Device2Ips[d]; ok {
						if OldIp.Second == nil {
							break
						}
						if OldIp.GetSecond() == handledIp[0] {
							Log.Info("ip not changed", "ip", OldIp.GetSecond())
							continue
						}
						Log.Info("ip changed", Log.String("old", OldIp.GetSecond()).String(), Log.String("new", handledIp[0]).String())
						*OldIp.Second = handledIp[0]
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
						Log.Info(fmt.Sprintf("result for %s.%s", d, "ipv6"),
							Log.Int("done", res[done]).String(),
							Log.Int("unaffected", res[unaffected]).String(),
							Log.Int("error", res[errorOccur]).String(),
							Log.Int("timeout", res[timeout]).String())
					}
				default:
					panic("unknown type")
				}
			}

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
					if (*service).GetType() == _type {
						(*service).SetValue(newIp)
						request, err := (*service).ToRequest()
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
						Display(request)
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

	wg.Wait()

	_, _ = Log.DebugPP.Fprintln(output, "all services finished")
}
