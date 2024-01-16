package config

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Config struct {
	Log Log `yaml:"log"`
}

//type Log struct {
//	Level logrus.Level `default:"debug" yaml:"level"`
//
//	UseRotator bool
//	Rotator    Rotator `yaml:"rotator"`
//}

type Log struct {
	Level logrus.Level `yaml:"level"`

	RotatorEnabled bool    `yaml:"rotator_enabled"`
	Rotator        Rotator `yaml:"rotator"`

	StdErrEnabled bool `yaml:"stderr_enabled"`
}

//type Rotator struct {
//	Filename string `default:"/var/log/pind/pind.log" yaml:"file_name"`
//	MaxSize  int    `default:"10" yaml:"max_size"`
//}

type Rotator struct {
	Filename   string `yaml:"file_name"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	LocalTime  bool   `yaml:"locale_time"`
}

func getDefaultConfig() *Config {
	config := &Config{}

	rotator := Rotator{
		Filename:   "/var/log/pind/pind.log",
		MaxSize:    10, // mb
		MaxBackups: 5,
		MaxAge:     28,
		LocalTime:  true,
	}

	config.Log.Level = logrus.DebugLevel
	config.Log.RotatorEnabled = false
	config.Log.Rotator = rotator
	config.Log.StdErrEnabled = true

	return config
}

func Load(confPath0 string) (*Config, error) {
	bytes0, err := os.ReadFile(confPath0)
	if err != nil {
		log.Printf("config Load err = %v\n", err)
		return nil, err
	}

	config := getDefaultConfig()
	err = yaml.Unmarshal(bytes0, config)
	if err != nil {
		log.Printf("config yaml.Unmarshal err = %v\n", err)
		return nil, err
	}

	return config, err
}
