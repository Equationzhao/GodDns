package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"GodDns/Device"
	log "GodDns/Log"
	"GodDns/Util/Collections"
	"GodDns/core"
	"github.com/urfave/cli/v2"
)

func GetApp(configFactoryList []core.ConfigFactory, parameters []*core.Parameters, GlobalDevice Device.Device) *cli.App {
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
						p := p
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
					mdFlag,
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
							mdFlag,
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
									mdFlag,
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

					render := func(configStr core.ConfigStr, out string) (string, error) {
						configStr.Content = "# " + strings.ReplaceAll(strings.ReplaceAll(configStr.Content, "#", "##"), "\n", "\n\n")
						outi, err := GetMDRenderer().Render(configStr.Content)
						if err != nil {
							// should not reach here
							return "", err
						}
						out += outi
						return out, nil
					}

					if !all {
						services := c.Args().Slice()
						if len(services) == 0 {
							return errors.New("at least one service/section name is required")
						}
						Collections.RemoveDuplicate(&services)
						var out string
						for _, service := range services {
							found := false
							for _, configFactory := range configFactoryList {
								if strings.EqualFold(configFactory.GetName(), service) {
									found = true
									configStr, erri := configFactory.Get().GenerateDefaultConfigInfo()
									if !md {
										if erri != nil {
											err = errors.Join(err, erri)
											break
										}
										out += configStr.Content
										break
									} else {
										outi, err2 := render(configStr, out)
										if err2 != nil {
											return err2
										}
										out += outi
										break
									}
								}
							}
							if !found {
								erri := fmt.Errorf("service/section %s not found", service)
								err = errors.Join(err, erri)
							}
						}
						_, _ = log.InfoPP.Fprintln(output, out)
					} else {
						var out string
						for _, configFactory := range configFactoryList {
							configStr, erri := configFactory.Get().GenerateDefaultConfigInfo()
							if !md {
								if erri != nil {
									err = errors.Join(err, erri)
								}
								out += configStr.Content
							} else {
								configStr.Content = "# " + strings.ReplaceAll(strings.ReplaceAll(configStr.Content, "#", "##"), "\n", "\n\n")
								outi, err := GetMDRenderer().Render(configStr.Content)
								if err != nil {
									return err
								}
								out += outi
							}
						}
						_, _ = log.InfoPP.Fprintln(output, out)
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
					mdFlag,
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
	return app
}
