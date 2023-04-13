package main

import (
	"errors"
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	log "GodDns/Log"
	"GodDns/Net"
	"GodDns/core"
	"github.com/urfave/cli/v2"
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
		_ = core.MainGoroutinePool.Submit(func() {
			CheckVersionUpgrade(msg)
		})
		_, _ = fmt.Println(core.NowVersionInfo())

		_, _ = fmt.Println(func() string {
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
	defaultLocation, err = core.GetDefaultConfigurationLocation()
	if err != nil {
		defaultLocation = "./DDNS.conf"
	}
}

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

	onChangeScanTimeFlag = &cli.StringFlag{
		Name:        "on-change-scan-time",
		Aliases:     []string{"ocst", "OCST"},
		Usage:       "scan ip change per time",
		DefaultText: "1 minute",
		Category:    "TRIGGER",
		Action: func(context *cli.Context, s string) error {
			if !context.Bool("on-change") {
				return errors.New("on-change-scan-time should be used with on-change")
			}

			t, err := time.ParseDuration(s)
			if err != nil {
				return err
			}
			if t.Seconds() < MINTIMEGAP {
				return errors.New("time gap is too short, should >= 5 seconds")
			}
			core.UniversalConfig[core.OcScanTime] = t
			return nil
		},
	}

	timeFlag = &cli.StringFlag{
		Name:        "time",
		Aliases:     []string{"t", "T"},
		DefaultText: "disabled",
		Usage:       "run ddns per time",
		Action: func(context *cli.Context, s string) error {
			t, err := time.ParseDuration(s)
			if err != nil {
				return err
			}
			Time = uint64(t.Seconds())
			if Time < MINTIMEGAP {
				core.ReturnCode = 2
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
				core.ReturnCode = 2
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
				prof, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0o644)
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

	boxFlag = &cli.BoolFlag{
		Name:        "print-in-table",
		Aliases:     []string{"tb", "table", "pit"},
		DefaultText: "disabled",
		Usage:       "print result in table **may render incorrectly in some terminals**",
		Destination: &tab,
		Category:    "OUTPUT",
	}

	mdFlag = &cli.BoolFlag{
		Name:        "print-in-markdown",
		Aliases:     []string{"md", "markdown", "pim"},
		DefaultText: "disabled",
		Usage:       "print result in markdown",
		Destination: &md,
		Category:    "OUTPUT",
	}
)
