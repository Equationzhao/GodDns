// Package core
// basic interfaces and tools for DDNS service
package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	log "GodDns/Log"
	"gopkg.in/ini.v1"
)

func init() {
	ini.PrettyFormat = false // config style key=value without space
}

// Format define the key=value format of config file
const (
	Format     = "%s=%v"
	ConfigName = "DDNS.conf"
)

var configFileLocation string

// ConfigFactoryList is a list of ConfigFactory
var ConfigFactoryList []ConfigFactory

// Add2FactoryList add ConfigFactory to ConfigFactoryList
func Add2FactoryList(factory ...ConfigFactory) {
	ConfigFactoryList = append(ConfigFactoryList, factory...)
}

// GetDefaultConfigurationLocation get default configuration location
var GetDefaultConfigurationLocation = defaultConfigurationLocation()

func defaultConfigurationLocation() func() (string, error) {
	defaultConfiguration, err := defaultConfigurationDirectory()
	if err != nil {
		return func() (string, error) {
			return "", err
		}
	}

	sep := string(filepath.Separator) // get system separator
	defaultConfiguration += sep + ConfigName

	return func() (string, error) {
		return defaultConfiguration, err
	}
}

// defaultConfigurationDirectory get default configuration directory like /home/user/.config/GodDns
// make dir if not exist
func defaultConfigurationDirectory() (string, error) {
	sep := string(filepath.Separator)                                             // get system separator
	defaultConfiguration, err := os.UserConfigDir()                               // get user config dir
	err = errors.Join(err, os.MkdirAll(defaultConfiguration+sep+FullName, 0o777)) // create sub directory
	return defaultConfiguration + sep + FullName, err
}

// UpdateConfigureLocation update config file location
func UpdateConfigureLocation(newLocation string) {
	configFileLocation = newLocation
}

// Config interface for config that can be read and write config from/to file/parameters
type Config interface {
	GetName() string
	GenerateDefaultConfigInfo() (ConfigStr, error)
	ReadConfig(sec ini.Section) ([]Parameters, error)
	// GenerateConfigInfo [Name#No]\n + KeyValue(s) + \n\n
	GenerateConfigInfo(Parameters, uint) (ConfigStr, error)
}

// NameMatch customized name rule to match
// used in ConfigureReader
// if a service implements this interface
// the service will match section name by its NameMatch rule
type NameMatch interface {
	MatchName(string) bool
}

// ConfigFactory factory to create Config
type ConfigFactory interface {
	GetName() string
	Get() Config
	New() *Config
}

// ConfigStr config file content
// Name: config service name
// Content: config service content
type ConfigStr struct {
	Name    string
	Content string
}

// GetConfigureLocation  Get config file location
// Should call after GetDefaultConfigurationLocation or UpdateConfigureLocation
func GetConfigureLocation() string {
	return configFileLocation
}

// MissingKeyErr presents a key is not found in config file
type MissingKeyErr struct {
	KeyName     string
	SectionName string
}

type UnknownKeyErr struct {
	KeyName     string
	SectionName string
}

func (u UnknownKeyErr) Error() string {
	return fmt.Sprintf("unknown key %s in %s", u.KeyName, u.SectionName)
}

// NewUnknownKeyErr create a new UnknownKeyErr
func NewUnknownKeyErr(keyName string, sectionName string) *UnknownKeyErr {
	return &UnknownKeyErr{KeyName: keyName, SectionName: sectionName}
}

// NewMissKeyErr create a new MissingKeyErr
func NewMissKeyErr(keyName string, sectionName string) *MissingKeyErr {
	return &MissingKeyErr{KeyName: keyName, SectionName: sectionName}
}

// Error return error message
func (m MissingKeyErr) Error() string {
	return fmt.Sprintf("miss key %s in %s", m.KeyName, m.SectionName)
}

// ConfigureWriter Create key style config file
// structure :
// ServiceName -> [ServiceName]
// Key -> Key=value
// Any Service should use this function to create config file
func ConfigureWriter(filename string, flag int, config ...ConfigStr) error { // option: append/w
	log.Debugf("open file at %s", filename)

	configure, err := os.OpenFile(filename, flag, 0o777) // os.O_CREATE|os.O_WRONLY

	if err != nil {
		return err
	} else {
		log.Tracef("open config file at %s", filename)
		defer func(configure *os.File) {
			err := configure.Close()
			if err != nil {
				log.Error("failed to close configure ", log.String("error", err.Error()))
			}
		}(configure)
	}

	for _, service := range config {
		log.Tracef("write config for %s", service.Name)
		_, err = configure.WriteString(service.Content)
		if err != nil {
			return err
		}
	}
	_ = configure.Sync()
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
	return configReader(Filename, configs, ReadConfigErrs)
}

func configReader(Filename string, configs []ConfigFactory, ReadConfigErrs error) ([]Parameters, error, error) {
	cfg, err := ini.Load(Filename)

	if err != nil {
		return nil, fmt.Errorf("failed to read configure at %s: %w", Filename, err), nil
	} else {
		log.Infof("load config file at %s", Filename)
	}

	cfg.BlockMode = false // !make sure read only

	ps := make([]Parameters, 0, 5*len(configs))
	var errCount uint8 = 0
	secs := cfg.Sections()
	for _, sec := range secs {
		for _, c := range configs {
			var match bool
			if _, ok := c.Get().(NameMatch); ok {
				match = c.Get().(NameMatch).MatchName(sec.Name()) // customized pattern, you can compare NameI NameII NameIII... if you want
			} else {
				pattern := regexp.MustCompile(regexp.QuoteMeta(c.GetName()) + `(#\d+)?$`) // default pattern
				match = pattern.MatchString(sec.Name())
			}

			// Read corresponding service
			if match {
				log.Debugf("read config for %s", c.GetName())
				temp, err := c.Get().ReadConfig(*sec)
				if err != nil {
					errCount++
					msg := fmt.Errorf("failed to read config for %s : %s", c.GetName(), err.Error())
					ReadConfigErrs = errors.Join(ReadConfigErrs, msg)
					log.Debug(msg)
					continue // skip this service
				}
				log.Tracef("%s : %s", c.GetName(), temp)
				log.Debugf("succeed to read config for %s", c.GetName())
				ps = append(ps, temp...)
				break
			}
		}
	}

	if ReadConfigErrs != nil {
		log.Errorf("finish with %d error(s)", errCount)
	}
	return ps, nil, ReadConfigErrs
}

// ConfigHead generate config head, the section name
// [Name#No]
// if No == 0, [Name]
func ConfigHead(parameters Parameters, No uint) (head string) { // todo head comment
	if No != 0 {
		head = "[" + parameters.GetName() + "#" + strconv.Itoa(int(No)) + "]" + "\n"
	} else {
		head = "[" + parameters.GetName() + "]" + "\n"
	}
	return
}
