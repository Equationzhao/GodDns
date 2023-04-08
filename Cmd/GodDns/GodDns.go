package main

import (
	"GodDns/Core"
	"GodDns/Device"
	log "GodDns/Log"
	"GodDns/Net"
	_ "GodDns/Service" // register all services
	"errors"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"runtime/debug"
	"runtime/pprof"
	"syscall"
	"time"
)

var output = os.Stdout

const MAXRETRY = 255
const MINTIMEGAP = 5
const MAXTIMES = 2628000
const DEFAULTRETRYATTEMPT = 3

const (
	run             = "run"
	runAuto         = "run-auto"
	runApi          = "run-api"
	runAutoOverride = "run-auto-override"
)

func init() {
	Core.MainGoroutinePool, _ = ants.NewPool(200, ants.WithNonblocking(true))
}

// global variables
var (
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
)

// Flags
var (
	silentFlag = &cli.BoolFlag{
		Name:    "no-output",
		Aliases: []string{"s", "S", "silent"},
		Value:   false,
		Usage:   "no message output",
		Action: func(context *cli.Context, silent bool) error {
			// set output
			if silent {
				output = nil
			}
			return nil
		},
		Category: "OUTPUT",
	}

	onChangeFlag = &cli.BoolFlag{
		Name:        "on-change",
		Aliases:     []string{"oc", "OC"},
		Value:       false,
		Usage:       "run ddns automatically when ip changed",
		Destination: &onChange,
		Category:    "TRIGGER",
	}

	timeFlag = &cli.Uint64Flag{
		Name:        "time",
		Aliases:     []string{"t", "T"},
		Value:       0,
		DefaultText: "disabled",
		Usage:       "run ddns per time(`seconds`)",
		Destination: &Time,
		Action: func(context *cli.Context, u uint64) error {
			if u < MINTIMEGAP {
				Core.ReturnCode = 2
				return fmt.Errorf("time gap is too short, should be more than %d seconds", MINTIMEGAP)
			}
			return nil
		},
		Category: "TRIGGER",
	}

	timesLimitationFlag = &cli.IntFlag{
		Name:        "times-limitation",
		Aliases:     []string{"tl", "TL"},
		Value:       0,
		DefaultText: "infinity",
		Usage:       "run ddns per time(seconds) up to `n` times",
		Destination: &TimesLimitation,
		Action: func(context *cli.Context, i int) error {
			t := context.Uint64("time")
			if t == 0 && !onChange {
				Core.ReturnCode = 2
				return errors.New("time limitation must be used with time flag or on-change flag")
			}
			return nil
		},
		Category: "TIMES",
	}

	retryFlag = &cli.UintFlag{
		Name:  "retry",
		Value: DEFAULTRETRYATTEMPT,
		Usage: "retry `times`",
		Action: func(context *cli.Context, u uint) error {
			if u > MAXRETRY {
				return fmt.Errorf("too many retry times, should be less than %d", MAXRETRY)
			}
			retryAttempt = uint8(u)
			return nil
		},
		Category: "RUN",
	}

	logFlag = &cli.StringFlag{
		Name:        "log",
		Aliases:     []string{"l", "L", "Log"},
		DefaultText: "Info",
		Value: func() string {
			debugEnv := os.Getenv("DEBUG")
			if debugEnv != "" {
				return "Debug"
			}
			return "Info"
		}(),
		Usage:       "`level`: Trace/Debug/Info/Warn/Error",
		Destination: &logLevel,
		Category:    "OUTPUT",
	}

	configFlag = &cli.StringFlag{
		Name:        "config",
		Aliases:     []string{"c", "C", "Config"},
		Value:       "",
		DefaultText: defaultLocation,
		Usage:       "set configuration `file`",
		Destination: &config,
		Category:    "CONFIG",
	}

	proxyFlag = &cli.StringFlag{
		Name:        "proxy",
		Aliases:     []string{"p", "P", "Proxy"},
		Value:       "",
		Usage:       "set proxy `url`",
		Destination: &proxy,
		Action: func(context *cli.Context, s string) error {
			if s != "" {
				if s == "enable" {
					if Net.GlobalProxies.GetProxyIter().Len() == 0 {
						return fmt.Errorf("no proxy url found, please set proxy url in config/env/flag first")
					}
					proxyEnable = true
					return nil
				} else if s == "disable" {
					return nil
				} else if Net.IsProxyValid(s) {
					proxyEnable = true
					Net.AddProxy2Top(Net.GlobalProxies, s)
					return nil
				} else {
					return errors.New("invalid proxy url")
				}
			}
			return errors.New("empty proxy url")
		},
		Category: "RUN",
	}

	parallelFlag = &cli.BoolFlag{
		Name:        "parallel",
		Aliases:     []string{"Parallel"},
		Value:       false,
		Usage:       "run ddns parallel",
		Destination: &parallelExecuting,
		Category:    "RUN",
	}

	cpuProfilingFlag = &cli.BoolFlag{
		Name:        "cpu-profile",
		Aliases:     []string{"cpuprofile", "cpu", "cp"},
		Value:       false,
		DefaultText: "disabled",
		Usage:       "enable cpu profiling",
		Category:    "PERFORMANCE",
		Action: func(context *cli.Context, b bool) error {
			if b {
				filename := "goddns-cpu-" + time.Now().Format("2006-01-02-15-04-05") + ".prof"
				prof, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
				if err != nil {
					return err
				}
				err = pprof.StartCPUProfile(prof)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	memProfilingFlag = &cli.BoolFlag{
		Name:        "mem-profile",
		Aliases:     []string{"memprofile", "mem", "mp"},
		Value:       false,
		DefaultText: "disabled",
		Usage:       "enable memory profiling",
		Category:    "PERFORMANCE",
		Destination: &memProfiling,
	}
)

func init() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:               "version",
		Aliases:            []string{"v", "V"},
		Usage:              "print the version info/upgrade info",
		DisableDefaultText: true,
	}

	cli.VersionPrinter = func(c *cli.Context) {
		msg := make(chan string, 2)
		_ = Core.MainGoroutinePool.Submit(func() {
			CheckVersionUpgrade(msg)
		})
		fmt.Println(Core.NowVersionInfo())

		fmt.Println(func() string {
			{
				info, err := os.Stat(os.Args[0])
				if err != nil {
					return ""
				}
				t := info.ModTime().Local()
				return fmt.Sprintf("compiled at %s", t.Format(time.RFC3339))
			}
		}())
		for i := 0; i < 2; i++ {
			select {
			case s := <-msg:
				if s != "" {
					_, _ = log.DebugPP.Println(s) // use debug pretty print for green color
				}
			case <-time.After(2 * time.Second):
				return
			}
		}

	}

	cli.HelpFlag = &cli.BoolFlag{
		Name:               "help",
		Aliases:            []string{"h", "H"},
		Usage:              "show help",
		DisableDefaultText: true,
	}

	var err error
	defaultLocation, err = Core.GetDefaultConfigurationLocation()
	if err != nil {
		defaultLocation = "./DDNS.conf"
	}

}

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
		_, err := log.InitLog("DDNS.log", 0666, l, output)
		if err != nil {
			log.Error("failed to init log file ", log.String("error", err.Error()))
			return err
		}
		isLogSet = true
		// cleanUp = clean
		return nil
	default:
		return errors.New("invalid log level")
	}
}

// todo return config setting command `GodDns config -service=cloudflare`
func main() {

	defer func() {
		if memProfiling {
			filename := "goddns-mem-" + time.Now().Format("2006-01-02-15-04-05") + ".prof"
			prof, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				log.Error(err)
			}
			err = pprof.WriteHeapProfile(prof)
			if err != nil {
				log.Error(err)
			}
		}

		os.Exit(Core.ReturnCode)
	}()
	defer pprof.StopCPUProfile()
	defer Core.CatchPanic(output)

	var parameters []Core.Parameters
	var GlobalDevice Device.Device
	configFactoryList := Core.ConfigFactoryList

	location, err := Core.GetProgramConfigLocation()
	if err != nil {
		_, _ = log.ErrPP.Fprintln(output, "error loading program config: ", err, " use default config")
	} else {
		if Core.IsConfigExist(location) {
			programConfig, fatal, warn := Core.LoadProgramConfig(location)
			if fatal != nil {
				// default setup
				_, _ = log.ErrPP.Fprintln(output, "error loading program config: ", err, " use default config")
				_, _ = log.ErrPP.Fprintln(output, fatal)
				Core.DefaultConfig.Setup()
			} else {
				if warn != nil {
					_, _ = log.WarnPP.Fprintln(output, warn)
				}
				programConfig.Setup()
			}
		} else {
			// create Config here
			_, _ = log.ErrPP.Fprintln(output, "no config at ", location, " try to generate a default config")
			err := Core.DefaultConfig.GenerateConfigFile()
			Core.DefaultConfig.Setup()
			if err != nil {
				_, _ = log.ErrPP.Fprintln(output, "failed to generate default program config at ", location)
			} else {
				_, _ = log.ErrPP.Fprintln(output, "generate default program config at ", location)
			}
		}
	}

	app := &cli.App{
		Name:     Core.FullName,
		Usage:    "A DDNS tool written in Go",
		Version:  Core.NowVersion.Info(),
		Compiled: time.Now(),
		Authors: []*cli.Author{
			{
				Name:  Core.Author,
				Email: Core.FeedbackEmail(),
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
						Core.UpdateConfigureLocation(config)
					} else {
						Core.UpdateConfigureLocation(defaultLocation)
					}

					parametersTemp, err := ReadConfig(configFactoryList)
					if err != nil {
						return err
					}
					parameters = parametersTemp

					if ApiName == "" {
						runMode = run
					} else {
						runMode = runApi
					}

					if Time != 0 {
						_ = RunDDNS(&parameters)
						RunPerTime(Time, nil, parameters)
						return nil
					}

					return ModeController(&parameters, nil)
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
							timesLimitationFlag,
							retryFlag,
							silentFlag,
							logFlag,
							configFlag,
							proxyFlag,
							cpuProfilingFlag,
							memProfilingFlag,
						},
						Action: func(context *cli.Context) error {

							err := checkLog(logLevel)
							if err != nil {
								return err
							}

							if config != "" {
								Core.UpdateConfigureLocation(config)
							} else {
								Core.UpdateConfigureLocation(defaultLocation)
							}

							parametersTemp, err := ReadConfig(configFactoryList)
							if err != nil {
								return err
							}
							parameters = parametersTemp
							GlobalDevice, err = GetGlobalDevice(parameters)
							if err != nil {
								return err
							}

							runMode = runAuto
							if onChange {
								OnChange(&parameters, &GlobalDevice)
								return nil
							}

							if Time != 0 {
								_ = RunAuto(GlobalDevice, &parameters)
								RunPerTime(Time, &GlobalDevice, parameters)
								return nil
							}

							return ModeController(&parameters, &GlobalDevice)
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
									timesLimitationFlag,
									retryFlag,
									silentFlag,
									logFlag,
									configFlag,
									proxyFlag,
									cpuProfilingFlag,
									memProfilingFlag,
								},
								Action: func(context *cli.Context) error {

									err := checkLog(logLevel)
									if err != nil {
										return err
									}

									if config != "" {
										Core.UpdateConfigureLocation(config)
									} else {
										Core.UpdateConfigureLocation(defaultLocation)
									}

									parametersTemp, err := ReadConfig(configFactoryList)
									if err != nil {
										return err
									}
									parameters = parametersTemp
									GlobalDevice, err = GetGlobalDevice(parameters)
									if err != nil {
										return err
									}

									runMode = runAutoOverride
									if onChange {
										OnChange(&parameters, &GlobalDevice)
										return nil
									}

									if Time != 0 {
										_ = RunOverride(GlobalDevice, &parameters)
										RunPerTime(Time, &GlobalDevice, parameters)
										return nil
									}

									return ModeController(&parameters, &GlobalDevice)
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
						Core.UpdateConfigureLocation(config)
					} else {
						Core.UpdateConfigureLocation(defaultLocation)
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
		},
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	_ = Core.MainGoroutinePool.Submit(func() {
		defer func() {
			if err := recover(); err != nil {
				Core.MainPanicHandler.Receive(err, debug.Stack())
				Core.PrintPanic(output, Core.Errchan)
			}
		}()
		Core.Errchan <- app.Run(os.Args)
	})

	select {
	case err = <-Core.Errchan:
		if err != nil {
			if Core.ReturnCode == 0 {
				Core.ReturnCode = 1
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
