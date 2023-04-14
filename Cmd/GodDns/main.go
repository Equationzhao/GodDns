package main

import (
	"errors"
	"os"
	"os/signal"
	"runtime/pprof"
	"sync"
	"syscall"
	"time"

	"GodDns/Device"
	log "GodDns/Log"
	_ "GodDns/Service" // register all services
	"GodDns/core"
	"github.com/charmbracelet/glamour"
	"github.com/panjf2000/ants/v2"
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
	core.MainGoroutinePool, _ = ants.NewPool(200, ants.WithNonblocking(false))
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
	tab               bool
	md                bool
)

var (
	mdRenderer sync.Once
	renderer   *glamour.TermRenderer
)

func GetMDRenderer() *glamour.TermRenderer {
	mdRenderer.Do(
		func() {
			renderer, _ = glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithEmoji(),
			)
		},
	)
	return renderer
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

	app := GetApp(configFactoryList, parameters, GlobalDevice)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	_ = core.MainGoroutinePool.Submit(func() {
		defer core.CatchPanic(output)
		core.Errchan <- app.Run(os.Args)
	})

	select {
	case err = <-core.Errchan:
		if err != nil {
			if core.ReturnCode == 0 {
				core.ReturnCode = 1
			}
			if isLogSet {
				log.Errorf("fatal:\n%s", err)
			} else {
				_, _ = log.ErrPP.Fprintln(output, "fatal:\n", err.Error())
			}
		}
	case <-interrupt:
		log.Warn("interrupted by user")
	}
}
