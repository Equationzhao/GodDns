/*
 *     @Copyright
 *     @file: main.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/25 上午1:46
 *     @last modified: 2023/3/25 上午1:45
 *
 *
 *
 */

package main

import (
	"GodDns/DDNS"
	"GodDns/Device"
	"GodDns/Log"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	_ "GodDns/Service" // register all services
)

var output = os.Stdout

const MAXRETRY = 255
const defaultRetryAttempt = 3

var (
	Time            uint64 = 0
	ApiName                = ""
	retryAttempt    uint8  = 0
	config                 = ""
	defaultLocation        = ""
	log                    = "Info"
	// cleanUp         func()
)

var (
	silentFlag = &cli.BoolFlag{
		Name:    "silent",
		Aliases: []string{"s", "S"},
		Value:   false,
		Usage:   "no message output",
		Action: func(context *cli.Context, silent bool) error {
			// set output
			if silent {
				output = nil
			}
			return nil
		},
	}

	TimeFlag = &cli.Uint64Flag{
		Name:        "time",
		Value:       0,
		Usage:       "run ddns per time(`seconds`)",
		Destination: &Time,
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
	}

	logFlag = &cli.StringFlag{
		Name:        "log",
		Aliases:     []string{"l", "L"},
		Value:       "Info",
		Usage:       "`level`: Trace/Debug/Info/Warn/Error",
		Destination: &log,
	}

	configFlag = &cli.StringFlag{
		Name:        "config",
		Aliases:     []string{"c", "C"},
		Value:       "",
		DefaultText: defaultLocation,
		Usage:       "set configuration `file`",
		Destination: &config,
	}
)

func init() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v", "V"},
		Usage:   "print the version info",
	}

	cli.VersionPrinter = func(c *cli.Context) {
		msg := make(chan string, 2)
		go CheckVersionUpgrade(msg)
		fmt.Println(DDNS.NowVersionInfo())
		for i := 0; i < 2; i++ {
			select {
			case s := <-msg:
				if s != "" {
					fmt.Println(s)
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
		_, err := InitLog("DDNS.log", 0666, l, output)
		if err != nil {
			logrus.Errorf("failed to init log file: %s", err)
			return err
		}
		// cleanUp = clean
		return nil
	default:
		return errors.New("invalid log level")
	}
}

// InitLog
// initialize the log file with fileMode and log level
// print information to output
// return a function to close the log file
// if error occurs, return error
func InitLog(filename string, filePerm os.FileMode, loglevel string, output io.Writer) (func(), error) {

	var level logrus.Level
	switch loglevel {
	// case "Panic", "panic", "PANIC":
	// 	level = logrus.PanicLevel
	// case "Fatal", "fatal", "FATAL":
	// 	level = logrus.FatalLevel
	case "Error", "error", "ERROR":
		level = logrus.ErrorLevel
	case "Warn", "warn", "WARN":
		level = logrus.WarnLevel
	case "Info", "info", "INFO":
		level = logrus.InfoLevel
	case "Debug", "debug", "DEBUG":
		level = logrus.DebugLevel
	case "Trace", "trace", "TRACE":
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
		fmt.Println("close log file")
		if err != nil {
			logrus.Error("failed to close log file ", err)
		}
	}

	logrus.SetLevel(level)
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: time.DateTime,
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			filename := path.Base(frame.File)
			if level >= logrus.DebugLevel {
				return frame.Function, filename
			}
			return "", ""
		},
	})

	Log.To(logrus.StandardLogger(), file)

	if output != nil {
		if _, ok := output.(*os.File); !ok || output.(*os.File) != nil {
			// output is not *os.File(nil)
			logrus.AddHook(Log.NewLogrusOriginally2writer(output))
		}
	}
	logrus.Infof("init log file at %s\n", filename)
	_, err = file.Write([]byte(fmt.Sprintf("---------start at %s---------\n", time.Now().Format(time.DateTime))))
	if err != nil {
		return cleanUp, err
	}

	return cleanUp, nil
}

// todo return non-zero value when error occurs
// todo return config setting command `GodDns config -service=cloudflare`
func main() {

	var parameters []DDNS.Parameters
	var GlobalDevice Device.Device
	configFactoryList := DDNS.ConfigFactoryList

	location, err := DDNS.GetProgramConfigLocation()
	if err != nil {
		_, _ = fmt.Fprintln(output, "error loading program config: ", err, " use default config")
	} else {
		programConfig, fatal, other := DDNS.LoadProgramConfig(location)
		if fatal != nil {
			// skip setup
			_, _ = fmt.Fprintln(output, fatal)

		} else {
			if other != nil {
				_, _ = fmt.Fprintln(output, other)
			}
			programConfig.Setup()
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
					err := checkLog(log)
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
						return errors.New("")
					}
					parameters = parametersTemp

					if !retryFlag.IsSet() {
						retryAttempt = defaultRetryAttempt
					}

					return RunDDNS(parameters)
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "api",
						Aliases: []string{"i", "I"},

						Usage: "get ip address from provided `ApiName`, eg: ipify/identMe",

						Destination: &ApiName,
					},
					TimeFlag,
					retryFlag,
					silentFlag,
					logFlag,
					configFlag,
				},
				Subcommands: []*cli.Command{
					{
						Name:    "auto",
						Aliases: []string{"a", "A"},
						Usage:   "run ddns, use ip address of interface set in Device Section automatically",
						Action: func(context *cli.Context) error {

							err := checkLog(log)
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
								return errors.New("")
							}
							parameters = parametersTemp
							GlobalDevice, err = GetGlobalDevice(parameters)
							if err != nil {
								return errors.New("")
							}

							if !retryFlag.IsSet() {
								retryAttempt = defaultRetryAttempt
							}

							return RunAuto(GlobalDevice, parameters)
						},
						Flags: []cli.Flag{
							TimeFlag,
							retryFlag,
							silentFlag,
							logFlag,
							configFlag,
						},
						Subcommands: []*cli.Command{
							{
								Name:    "override",
								Aliases: []string{"o", "O"},
								Usage:   "run ddns, override the ip address of interface set in each service Section",
								Flags: []cli.Flag{
									TimeFlag,
									retryFlag,
									silentFlag,
									logFlag,
									configFlag,
								},
								Action: func(context *cli.Context) error {

									err := checkLog(log)
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
										return errors.New("")
									}
									parameters = parametersTemp
									GlobalDevice, err = GetGlobalDevice(parameters)
									if err != nil {
										return errors.New("")
									}

									if !retryFlag.IsSet() {
										retryAttempt = defaultRetryAttempt
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
					err := checkLog(log)
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
		After: func(context *cli.Context) error {
			// if cleanUp != nil {
			// 	// bug cleanUp()
			// }
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}

}
