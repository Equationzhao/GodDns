package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"

	"GodDns/core"

	"GodDns/Device"
	log "GodDns/Log"
	_ "GodDns/Service" // register all services

	"github.com/panjf2000/ants/v2"
	"github.com/urfave/cli/v2"
)

const (
	MAXRETRY            = 255
	MINTIMEGAP          = 5
	MAXTIMES            = 2628000
	DEFAULTRETRYATTEMPT = 3
)

const (
	run             = "run"
	runAuto         = "run-auto"
	runApi          = "run-api"
	runAutoOverride = "run-auto-override"
)

func init() {
	core.MainGoroutinePool, _ = ants.NewPool(200, ants.WithNonblocking(true))
}

// global variables
var (
	output            = os.Stdout
	Time              uint64
	TimesLimitation   int // 0 means no limitation
	ApiName           string
	retryAttempt      uint8 = DEFAULTRETRYATTEMPT
	config            string
	defaultLocation   string
	logLevel          string
	proxy             string
	proxyEnable       bool
	parallelExecuting bool
	runMode           string
	isLogSet          bool
	onChange          bool
	memProfiling      bool
	box               bool
)

func checkLog(l string) error {
	switch l {
	case "Trace", "trace", "TRACE":
		fallthrough
	case "Debug", "debug", "DEBUG":
		fallthrough
	case "Info", "info", "INFO":
		fallthrough
	case "Warn", "warn", "WARN":
		fallthrough
	case "Error", "error", "ERROR":
		_, err := log.InitLog("DDNS.log", 0o666, l, output)
		if err != nil {
			log.Error("failed to init log file ", log.String("error", err.Error()).String())
			return err
		}
		isLogSet = true
		return nil
	default:
		return errors.New("invalid log level")
	}
}

func main() {
	defer func() {
		if memProfiling {
			filename := "goddns-mem-" + time.Now().Format("2006-01-02-15-04-05") + ".prof"
			prof, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0o644)
			if err != nil {
				log.Error(err)
			}
			err = pprof.WriteHeapProfile(prof)
			if err != nil {
				log.Error(err)
			}
		}

		os.Exit(core.ReturnCode)
	}()
	defer pprof.StopCPUProfile()
	defer core.CatchPanic(output)

	var parameters []*core.Parameters
	var GlobalDevice Device.Device
	configFactoryList := core.ConfigFactoryList

	location, err := core.GetProgramConfigLocation()
	if err != nil {
		_, _ = log.ErrPP.Fprintln(output, "error loading program config: ", err, " use default config")
	} else {
		if core.IsConfigExist(location) {
			programConfig, fatal, warn := core.LoadProgramConfig(location)
			if fatal != nil {
				// default setup
				_, _ = log.ErrPP.Fprintln(output, "error loading program config, use default config")
				_, _ = log.ErrPP.Fprintln(output, fatal.Error())
				core.DefaultConfig.Setup()
			} else {
				if warn != nil {
					_, _ = log.WarnPP.Fprintln(output, warn.Error())
				}
				programConfig.Setup()
			}
		} else {
			// create Config here
			_, _ = log.ErrPP.Fprintln(output, "no config at ", location, " try to generate a default config")
			err := core.DefaultConfig.GenerateConfigFile()
			core.DefaultConfig.Setup()
			if err != nil {
				_, _ = log.ErrPP.Fprintln(output, "failed to generate default program config at ", location)
			} else {
				_, _ = log.ErrPP.Fprintln(output, "generate default program config at ", location)
			}
		}
	}

	app := &cli.App{
		Name:     core.FullName,
		Usage:    "A DDNS tool written in Go",
		Version:  core.NowVersion.Info(),
		Compiled: time.Now(),
		Authors: []*cli.Author{
			{
				Name:  core.Author,
				Email: core.FeedbackEmail(),
			},
		},
		Suggest:              true,
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:    "run",
				Aliases: []string{"r", "R"},
				Usage:   "run the DDNS service",

				Action: func(context *cli.Context) error {
					err := checkLog(logLevel)
					if err != nil {
						return err
					}
					if config != "" {
						core.UpdateConfigureLocation(config)
					} else {
						core.UpdateConfigureLocation(defaultLocation)
					}

					parametersTemp, err := ReadConfig(configFactoryList)
					if err != nil {
						return err
					}

					for _, p := range parametersTemp {
						parameters = append(parameters, &p)
					}

					if ApiName == "" {
						runMode = run
					} else {
						runMode = runApi
					}

					if Time != 0 {
						_ = RunDDNS(parameters)
						RunPerTime(Time, nil, parameters)
						return nil
					}

					return ModeController(parameters, nil)
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "api",
						Aliases: []string{"i", "I"},

						Usage: "get ip address from provided `ApiName`, eg: ipify/identMe",

						Destination: &ApiName,

						Category: "RUN",
					},
					parallelFlag,
					timeFlag,
					timesLimitationFlag,
					retryFlag,
					silentFlag,
					logFlag,
					configFlag,
					proxyFlag,
					cpuProfilingFlag,
					memProfilingFlag,
					boxFlag,
				},
				Subcommands: []*cli.Command{
					{
						Name:    "auto",
						Aliases: []string{"a", "A"},
						Usage:   "run ddns, use ip address of interface set in Device Section automatically",
						Flags: []cli.Flag{
							parallelFlag,
							timeFlag,
							onChangeFlag,
							onChangeScanTimeFlag,
							timesLimitationFlag,
							retryFlag,
							silentFlag,
							logFlag,
							configFlag,
							proxyFlag,
							cpuProfilingFlag,
							memProfilingFlag,
							boxFlag,
						},
						Action: func(context *cli.Context) error {
							err := checkLog(logLevel)
							if err != nil {
								return err
							}

							if config != "" {
								core.UpdateConfigureLocation(config)
							} else {
								core.UpdateConfigureLocation(defaultLocation)
							}

							parametersTemp, err := ReadConfig(configFactoryList)
							if err != nil {
								return err
							}
							for _, p := range parametersTemp {
								p := p
								parameters = append(parameters, &p)
							}

							GlobalDevice, err = GetGlobalDevice(parameters)
							if err != nil {
								return err
							}

							runMode = runAuto
							if onChange {
								OnChange(parameters, &GlobalDevice)
								return nil
							}

							if Time != 0 {
								_ = RunAuto(GlobalDevice, parameters)
								RunPerTime(Time, &GlobalDevice, parameters)
								return nil
							}

							return ModeController(parameters, &GlobalDevice)
						},
						Subcommands: []*cli.Command{
							{
								Name:    "override",
								Aliases: []string{"o", "O"},
								Usage:   "run ddns, override the ip address of interface set in each service Section",
								Flags: []cli.Flag{
									parallelFlag,
									timeFlag,
									onChangeFlag,
									onChangeScanTimeFlag,
									timesLimitationFlag,
									retryFlag,
									silentFlag,
									logFlag,
									configFlag,
									proxyFlag,
									cpuProfilingFlag,
									memProfilingFlag,
									boxFlag,
								},
								Action: func(context *cli.Context) error {
									err := checkLog(logLevel)
									if err != nil {
										return err
									}

									if config != "" {
										core.UpdateConfigureLocation(config)
									} else {
										core.UpdateConfigureLocation(defaultLocation)
									}

									parametersTemp, err := ReadConfig(configFactoryList)
									if err != nil {
										return err
									}
									for _, p := range parametersTemp {
										p := p
										parameters = append(parameters, &p)
									}
									GlobalDevice, err = GetGlobalDevice(parameters)
									if err != nil {
										return err
									}

									runMode = runAutoOverride
									if onChange {
										OnChange(parameters, &GlobalDevice)
										return nil
									}

									if Time != 0 {
										_ = RunOverride(GlobalDevice, parameters)
										RunPerTime(Time, &GlobalDevice, parameters)
										return nil
									}

									return ModeController(parameters, &GlobalDevice)
								},
							},
						},
					},
				},
			},
			{
				Name:    "generate",
				Aliases: []string{"g", "G"},
				Usage:   "generate a default configuration file",
				Action: func(*cli.Context) error {
					err := checkLog(logLevel)
					if err != nil {
						return err
					}

					if config != "" {
						core.UpdateConfigureLocation(config)
					} else {
						core.UpdateConfigureLocation(defaultLocation)
					}
					return GenerateConfigure(configFactoryList)
				},
				Flags: []cli.Flag{
					silentFlag,
					logFlag,
					configFlag,
					cpuProfilingFlag,
					memProfilingFlag,
				},
			},
			{
				Name:    "show-config",
				Aliases: []string{"sc", "SC"},
				Usage:   "show the configuration of a service/section *case insensitive*",
				Action: func(c *cli.Context) error {
					err := checkLog(logLevel)
					if err != nil {
						return err
					}

					all := c.Bool("all")
					if !all {
						services := c.Args().Slice()
						if len(services) == 0 {
							return errors.New("at least one service/section name is required")
						}

						for _, service := range services {
							found := false
							for _, configFactory := range configFactoryList {
								if strings.EqualFold(configFactory.GetName(), service) {
									found = true
									configStr, erri := configFactory.Get().GenerateDefaultConfigInfo()
									if erri != nil {
										_, _ = log.ErrPP.Println(erri)
										err = errors.Join(err, erri)
										break
									}
									_, _ = log.InfoPP.Println(configStr.Content)
									break
								}
							}
							if !found {
								erri := fmt.Errorf("service/section %s not found", service)
								_, _ = log.ErrPP.Printf("%s\n\n\n", erri)
								err = errors.Join(err, erri)
							}
						}
					} else {
						for _, configFactory := range configFactoryList {
							configStr, erri := configFactory.Get().GenerateDefaultConfigInfo()
							if erri != nil {
								_, _ = log.ErrPP.Println(erri)
								err = errors.Join(err, erri)
							}
							_, _ = log.InfoPP.Println(configStr.Content)
						}
					}
					return err
				},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "all",
						Aliases: []string{"a", "A"},
						Value:   false,
						Usage:   "show all available services/sections configuration",
					},
					logFlag,
					cpuProfilingFlag,
					memProfilingFlag,
				},
				Subcommands: []*cli.Command{
					{
						Name:  "ls",
						Usage: "list all available services/sections",
						Action: func(c *cli.Context) error {
							err := checkLog(logLevel)
							if err != nil {
								return err
							}
							for _, configFactory := range configFactoryList {
								_, _ = log.InfoPP.Println(configFactory.GetName())
							}
							return nil
						},
						Flags: []cli.Flag{
							logFlag,
							cpuProfilingFlag,
							memProfilingFlag,
						},
					},
				},
			},
		},
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	_ = core.MainGoroutinePool.Submit(func() {
		defer func() {
			if err := recover(); err != nil {
				core.MainPanicHandler.Receive(err, debug.Stack())
				core.PrintPanic(output, core.Errchan)
			}
		}()
		core.Errchan <- app.Run(os.Args)
	})

	select {
	case err = <-core.Errchan:
		if err != nil {
			if core.ReturnCode == 0 {
				core.ReturnCode = 1
			}
			if isLogSet {
				log.Errorf("fatal: %s", err)
			} else {
				_, _ = log.ErrPP.Fprintln(output, "fatal: ", err.Error())
			}
		}
	case <-interrupt:
		log.Warn("interrupted by user")
	}
}
