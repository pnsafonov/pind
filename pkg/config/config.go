package config

import (
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
)

const (
	ProcFilterTypeName     = "name"
	SelectionTypeSingle    = "single" // core per thread
	PinCoresAlgoTypeSingle = "single"
	IgnoreTypeName         = "name"
)

type Config struct {
	Log     *Log     `yaml:"log"`
	Service *Service `yaml:"service"`
}

type Log struct {
	Level logrus.Level `yaml:"level"`

	RotatorEnabled bool     `yaml:"rotator_enabled"`
	Rotator        *Rotator `yaml:"rotator"`

	StdErrEnabled bool `yaml:"stderr_enabled"`
}

type Rotator struct {
	Filename   string `yaml:"file_name"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	LocalTime  bool   `yaml:"locale_time"`
}

type Service struct {
	Interval     int           `yaml:"interval"` // ms
	Threshold    float64       `yaml:"threshold"`
	Filters0     []*ProcFilter `yaml:"filters0"`
	Filters1     []*ProcFilter `yaml:"filters1"`
	Pool         Pool          `yaml:"pool"`
	Selection    Selection     `yaml:"selection"`
	PinCoresAlgo *PinCoresAlgo `yaml:"pin_cores_algo"`
	Ignore       *Ignore       `yaml:"ignore"`
	HttpApi      *HttpApi      `yaml:"http_api"`
}

type Pool struct {
	Idle Intervals
	Load Intervals
}

type ProcFilter struct {
	Type     string   `yaml:"type"`
	Patterns []string `yaml:"patterns"`
}

type Selection struct {
	Type     string   `yaml:"type"`
	Patterns []string `yaml:"patterns"`
}

type Ignore struct {
	Type     string   `yaml:"type"`
	Patterns []string `yaml:"patterns"`
}

type PinCoresAlgo struct {
	Type        string `yaml:"type"`
	Selected    int    `yaml:"selected_cores_count"`
	NotSelected int    `yaml:"not_selected_cores_count"`
}

type HttpApi struct {
	Enabled bool   `yaml:"enabled"`
	Listen  string `yaml:"listen"`
}

func NewDefaultFilters0() []*ProcFilter {
	filter0 := &ProcFilter{
		Type:     ProcFilterTypeName,
		Patterns: []string{"/usr/bin/kvm"},
	}
	filter1 := &ProcFilter{
		Type:     ProcFilterTypeName,
		Patterns: []string{"/usr/bin/qemu-system-x86_64"},
	}
	filters := []*ProcFilter{filter0, filter1}
	return filters
}

func NewDefaultFilters1() []*ProcFilter {
	filter0 := &ProcFilter{
		Type:     ProcFilterTypeName,
		Patterns: []string{"deb-2"},
	}
	filter1 := &ProcFilter{
		Type:     ProcFilterTypeName,
		Patterns: []string{"deb-3"},
	}
	filters := []*ProcFilter{filter0, filter1}
	return filters
}

func NewDefaultConfig() *Config {
	config := &Config{}

	rotator := &Rotator{
		Filename:   "/var/log/pind/pind.log",
		MaxSize:    10, // mb
		MaxBackups: 5,
		MaxAge:     28,
		LocalTime:  true,
	}

	log0 := &Log{
		Level: logrus.InfoLevel,
		//Level:          logrus.DebugLevel,
		RotatorEnabled: false,
		Rotator:        rotator,
		StdErrEnabled:  true,
	}

	pool := Pool{
		Idle: Intervals{Values: []int{0, 1}},
		Load: Intervals{Values: []int{2, 3, 4, 5}},
		//Load: Intervals{Values: []int{2, 3, 4, 5, 6}},
	}

	selection := Selection{
		Type:     SelectionTypeSingle,
		Patterns: []string{"CPU", "/KVM"},
	}

	pinCoresAlgo := &PinCoresAlgo{
		Type:        PinCoresAlgoTypeSingle,
		Selected:    1, // 1 core per thread
		NotSelected: 2, // 2 cores for other threads
	}

	ignore := &Ignore{
		Type:     IgnoreTypeName,
		Patterns: []string{"iou-wrk-"},
	}

	httpApi := &HttpApi{
		Enabled: true,
		Listen:  "0.0.0.0:10331",
	}

	filters0 := NewDefaultFilters0()
	filters1 := NewDefaultFilters1()
	service := &Service{
		Interval:     1000,
		Threshold:    150,
		Filters0:     filters0,
		Filters1:     filters1,
		Pool:         pool,
		Selection:    selection,
		PinCoresAlgo: pinCoresAlgo,
		Ignore:       ignore,
		HttpApi:      httpApi,
	}

	config.Log = log0
	config.Service = service

	return config
}

func Load(confPath0 string) (*Config, error) {
	bytes0, err := os.ReadFile(confPath0)
	if err != nil {
		log.Errorf("Load, os.ReadFile err = %v\n", err)
		return nil, err
	}

	config := NewDefaultConfig()
	err = yaml.Unmarshal(bytes0, config)
	if err != nil {
		log.Errorf("Load, yaml.Unmarshal err = %v\n", err)
		return nil, err
	}

	return config, err
}

func ToString0(config *Config) (string, error) {
	bytes0, err := yaml.Marshal(config)
	if err != nil {
		log.Errorf("ConfigToString, yaml.Marshal err = %v\n", err)
		return "", err
	}
	return string(bytes0), nil
}
