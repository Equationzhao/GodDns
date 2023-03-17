/*
 *     @Copyright
 *     @file: Config.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/17 下午8:04
 *     @last modified: 2023/3/17 下午8:04
 *
 *
 *
 */

package DDNS

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

const Format = "%s=%v"

var configFileLocation string // todo support customized location

var ConfigFactoryList []ConfigFactory

func init() {
	ini.PrettyFormat = false // config style key=value without space

}

var GetDefaultConfigurationLocation = defaultConfigurationLocation()

func defaultConfigurationLocation() func() (string, error) {

	sep := string(filepath.Separator) // get system separator
	// get user config dir
	defaultConfiguration, err := os.UserConfigDir()
	// create sub directory
	err = errors.Join(err, os.MkdirAll(defaultConfiguration+sep+"DDNS-go", 0664))

	defaultConfiguration += sep + "DDNS-go" + sep + "DDNS.conf"

	return func() (string, error) {
		return defaultConfiguration, err
	}
}

func UpdateConfigureLocation(newLocation string) {
	configFileLocation = newLocation
}

type Config interface {
	GetName() string
	GenerateDefaultConfigInfo() (ConfigStr, error)
	ReadConfig(sec ini.Section) (Parameters, error)
	// GenerateConfigInfo [Name#No]\n + KeyValue(s) + \n\n
	GenerateConfigInfo(Parameters, uint) (ConfigStr, error)
}

// NameMatch
// customized name rule to match
type NameMatch interface {
	MatchName(string) bool
}

// ConfigFactory
// factory to create Config
type ConfigFactory interface {
	GetName() string
	Get() Config
	New() *Config
}

type ConfigStr struct {
	Name    string
	Content string
}

func GetConfigureLocation() string {
	return configFileLocation
}

// ConfigureWriter
// Create key style config file
// structure :
// ServiceName -> [ServiceName]
// Key -> Key=value
// Any Service should use this function to create config file
// todo support comments
func ConfigureWriter(Filename string, flag int, config ...ConfigStr) error { // option: append/w
	logrus.Debugf("open file at %s", Filename)

	configure, err := os.OpenFile(Filename, flag, 0664) // os.O_CREATE|os.O_WRONLY

	if err != nil {
		return err
	} else {
		logrus.Trace("open config file at ", Filename)
		defer func(configure *os.File) {

			err := configure.Close()
			if err != nil {
				logrus.Error("failed to close configure ", err)
			}
		}(configure)
	}

	for _, service := range config {
		logrus.Trace("write config for ", service.Name)
		_, err = configure.Write([]byte(service.Content))
		if err != nil {
			return err
		}
	}
	return nil
}

/*
ConfigureReader

Read key-value style config file
structure :
[Devices]
device=[DeviceName1,DeviceName2,...]

[Dnspod#1]
Key=value

[Dnspod#2]
Key=value

[Cloudflare#1]
Key=value
...
*/
func ConfigureReader(Filename string, configs ...ConfigFactory) (ps []Parameters, LoadFileErr error, ReadConfigErrs error) {
	cfg, err := ini.Load(Filename)

	if err != nil {
		return nil, fmt.Errorf("failed to read configure at %s: %w", Filename, err), nil
	} else {
		logrus.Info("load config file at ", Filename)
	}

	cfg.BlockMode = false // !make sure read only
	defer func() { cfg.BlockMode = true }()

	var errCount uint8 = 0
	secs := cfg.Sections()
	for _, sec := range secs {
		for _, c := range configs {
			match := false
			if _, ok := c.Get().(NameMatch); ok {
				match = c.Get().(NameMatch).MatchName(sec.Name()) // todo get regexp from c.Get() if it is can Match names
			} else {
				pattern := regexp.MustCompile(regexp.QuoteMeta(c.GetName()) + `(#\d+)?$`) // default pattern
				match = pattern.MatchString(sec.Name())
			}

			// Read corresponding service
			if match {
				logrus.Debugf("read config for %s", c.GetName())
				temp, err := c.Get().ReadConfig(*sec)
				logrus.Debug(temp)
				if err != nil {
					errCount++
					ReadConfigErrs = errors.Join(ReadConfigErrs, fmt.Errorf("failed to read config for %s : %s", c.GetName(), err.Error()))
					logrus.Debugf("failed to read config for %s : %s , skip this service", c.GetName(), err.Error())
					continue // skip this service
				}
				logrus.Tracef("%s : %v", c.GetName(), temp)
				logrus.Debugf("succeed to read config for %s", c.GetName())
				ps = append(ps, temp)
			}
		}

	}

	if ReadConfigErrs != nil {
		logrus.Errorf("finish with %d error(s)", errCount)
	}
	return ps, nil, ReadConfigErrs
}

func IsConfigureExist() bool {
	_, err := os.Stat(GetConfigureLocation())

	return !errors.Is(err, os.ErrNotExist)
}

func SaveConfig(FileName string, flag int, parameters ...Parameters) error {
	var err error
	n := make(map[string]uint)
	ConStrs := make([]ConfigStr, 0)
	for _, parameter := range parameters {
		var no uint
		if parameter.GetName() == "Devices" { // todo refactor do not use hard code "Devices"
			no = 0
		} else {
			n[parameter.GetName()]++
			no = n[parameter.GetName()]
		}
		ConStr, err_ := parameter.SaveConfig(no)

		if err_ != nil {
			err = errors.Join(err, err_)
		} else {
			ConStrs = append(ConStrs, ConStr)
		}
	}
	err = errors.Join(err, ConfigureWriter(FileName, flag, ConStrs...))
	return err
}

func ConfigHead(parameters Parameters, No uint) (head string) {
	if No != 0 {
		head = "[" + parameters.GetName() + "#" + strconv.Itoa(int(No)) + "]" + "\n"
	} else {
		head = "[" + parameters.GetName() + "]" + "\n"
	}
	return
}
