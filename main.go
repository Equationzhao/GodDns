package main

import (
	"GodDns/DDNS"
	"GodDns/Device"
	log "GodDns/Log"
	"GodDns/Net"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/urfave/cli/v2"

	_ "GodDns/Service" // register all services
)

var output = os.Stdout
var returnCode = 0

const MAXRETRY = 255
const defaultRetryAttempt = 3
const MINTIMEGAP = 5

const (
	run             = "run"
	runAuto         = "run-auto"
	runApi          = "run-api"
	runAutoOverride = "run-auto-override"
)

// global variables
var (
	Time              uint64 = 0
	TimeLimitation    uint64 = 0 // 0 means no limitation
	ApiName                  = ""
	retryAttempt      uint8  = defaultRetryAttempt
	config                   = ""
	defaultLocation          = ""
	logLevel                 = ""
	proxy                    = ""
	proxyEnable              = false
	parallelExecuting        = false
	runMode                  = ""
	isLogSet                 = false
	// cleanUp         func()
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

	timeFlag = &cli.Uint64Flag{
		Name:        "time",
		Aliases:     []string{"t", "T"},
		Value:       0,
		Usage:       "run ddns per time(`seconds`)",
		Destination: &Time,
		Action: func(context *cli.Context, u uint64) error {
			if u < MINTIMEGAP {
				returnCode = 2
				return fmt.Errorf("time gap is too short, should be more than %d seconds", MINTIMEGAP)
			}
			return nil
		},
		Category: "TIME",
	}

	timeLimitationFlag = &cli.Uint64Flag{
		Name:        "time-limitation",
		Aliases:     []string{"tl", "TL"},
		Value:       0,
		DefaultText: "infinity",
		Usage:       "run ddns per time(seconds) up to `n` times",
		Destination: &TimeLimitation,
		Action: func(context *cli.Context, u uint64) error {
			t := context.Uint64("time")
			if t == 0 {
				returnCode = 2
				return errors.New("time limitation must be used with time flag")
			}
			return nil
		},
		Category: "TIME",
	}

	retryFlag = &cli.UintFlag{
		Name:  "retry",
		Value: defaultRetryAttempt,
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
					if Net.GlobalProxys.GetProxyIter().Len() == 0 {
						return fmt.Errorf("no proxy url found, please set proxy url in config/env/flag first")
					}
					proxyEnable = true
					return nil
				} else if s == "disable" {
					return nil
				} else if Net.IsProxyValid(s) {
					proxyEnable = true
					Net.AddProxy2Top(Net.GlobalProxys, s)
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
)

func init() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v", "V"},
		Usage:   "print the version info/upgrade info",
	}

	cli.VersionPrinter = func(c *cli.Context) {
		msg := make(chan string, 2)
		go CheckVersionUpgrade(msg)
		fmt.Println(DDNS.NowVersionInfo())

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
					_, _ = log.DebugPP.Println(s)
				}
			case <-time.After(2 * time.Second):
				return
			}
		}

	}

	cli.HelpFlag = &cli.BoolFlag{
		Name:    "help",
		Aliases: []string{"h", "H"},
		Usage:   "show help",
	}

	var err error
	defaultLocation, err = DDNS.GetDefaultConfigurationLocation()
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
		os.Exit(returnCode)
	}()

	defer func() {
		if err := recover(); err != nil {
			_, _ = fmt.Fprintln(output, "panic: ", err)
			_, _ = fmt.Fprintln(output, string(debug.Stack()))
			_, _ = fmt.Fprintln(output, "please report this issue to ", DDNS.IssueURL())
			_, _ = fmt.Fprintln(output, "or send email to ", DDNS.FeedbackEmail())
			returnCode = 1
		}
	}()

	var parameters []DDNS.Parameters
	var GlobalDevice Device.Device
	configFactoryList := DDNS.ConfigFactoryList

	location, err := DDNS.GetProgramConfigLocation()
	if err != nil {
		_, _ = log.ErrPP.Fprintln(output, "error loading program config: ", err, " use default config")
	} else {
		if DDNS.IsConfigExist(location) {
			programConfig, fatal, warn := DDNS.LoadProgramConfig(location)
			if fatal != nil {
				// default setup
				_, _ = log.ErrPP.Fprintln(output, "error loading program config: ", err, " use default config")
				_, _ = log.ErrPP.Fprintln(output, fatal)
				DDNS.DefaultConfig.Setup()
			} else {
				if warn != nil {
					_, _ = log.WarnPP.Fprintln(output, warn)
				}
				programConfig.Setup()
			}
		} else {
			// create Config here
			_, _ = log.ErrPP.Fprintln(output, "no config at ", location, " try to generate a default config")
			err := DDNS.DefaultConfig.GenerateConfigFile()
			DDNS.DefaultConfig.Setup()
			if err != nil {
				_, _ = log.ErrPP.Fprintln(output, "failed to generate default program config at ", location)
			} else {
				_, _ = log.ErrPP.Fprintln(output, "generate default program config at ", location)
			}
		}
	}

	app := &cli.App{
		Name:     DDNS.FullName,
		Usage:    "A DDNS tool written in Go",
		Version:  DDNS.NowVersion.Info(),
		Compiled: time.Now(),
		Authors: []*cli.Author{
			{
				Name:  DDNS.Author,
				Email: DDNS.FeedbackEmail(),
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
						DDNS.UpdateConfigureLocation(config)
					} else {
						DDNS.UpdateConfigureLocation(defaultLocation)
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
						_ = RunDDNS(parameters)
						RunPerTime(Time, nil, parameters)
						return nil
					}

					return RunDDNS(parameters)
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
					timeLimitationFlag,
					retryFlag,
					silentFlag,
					logFlag,
					configFlag,
					proxyFlag,
				},
				Subcommands: []*cli.Command{
					{
						Name:    "auto",
						Aliases: []string{"a", "A"},
						Usage:   "run ddns, use ip address of interface set in Device Section automatically",
						Action: func(context *cli.Context) error {

							err := checkLog(logLevel)
							if err != nil {
								return err
							}

							if config != "" {
								DDNS.UpdateConfigureLocation(config)
							} else {
								DDNS.UpdateConfigureLocation(defaultLocation)
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

							if Time != 0 {
								_ = RunAuto(GlobalDevice, parameters)
								RunPerTime(Time, &GlobalDevice, parameters)
								return nil
							}

							return RunAuto(GlobalDevice, parameters)
						},
						Flags: []cli.Flag{
							parallelFlag,
							timeFlag,
							timeLimitationFlag,
							retryFlag,
							silentFlag,
							logFlag,
							configFlag,
							proxyFlag,
						},
						Subcommands: []*cli.Command{
							{
								Name:    "override",
								Aliases: []string{"o", "O"},
								Usage:   "run ddns, override the ip address of interface set in each service Section",
								Flags: []cli.Flag{
									parallelFlag,
									timeFlag,
									timeLimitationFlag,
									retryFlag,
									silentFlag,
									logFlag,
									configFlag,
									proxyFlag,
								},
								Action: func(context *cli.Context) error {

									err := checkLog(logLevel)
									if err != nil {
										return err
									}

									if config != "" {
										DDNS.UpdateConfigureLocation(config)
									} else {
										DDNS.UpdateConfigureLocation(defaultLocation)
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

									if Time != 0 {
										_ = RunOverride(GlobalDevice, parameters)
										RunPerTime(Time, &GlobalDevice, parameters)
										return nil
									}

									return RunOverride(GlobalDevice, parameters)
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
						DDNS.UpdateConfigureLocation(config)
					} else {
						DDNS.UpdateConfigureLocation(defaultLocation)
					}
					return GenerateConfigure(configFactoryList)
				},
				Flags: []cli.Flag{
					silentFlag,
					logFlag,
					configFlag,
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		if returnCode == 0 {
			returnCode = 1
		}

		if isLogSet {
			log.Errorf("fatal: %s", err)
		} else {
			_, _ = log.ErrPP.Fprintf(output, "fatal: %s", err.Error())
		}

	}

}
