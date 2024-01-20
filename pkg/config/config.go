package config

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

const (
	ProcFilterTypeName     = "name"
	SelectionTypeSingle    = "single" // core per thread
	PinCoresAlgoTypeSingle = "single"
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
	Filters      []*ProcFilter `yaml:"filters"`
	Pool         Pool          `yaml:"pool"`
	Selection    Selection     `yaml:"selection"`
	PinCoresAlgo *PinCoresAlgo `yaml:"pin_cores_algo"`
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

type PinCoresAlgo struct {
	Type        string `yaml:"type"`
	Selected    int    `yaml:"selection_cores_count"`
	NotSelected int    `yaml:"selection_cores_count"`
}

func NewDefaultFilters() []*ProcFilter {
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
		Level:          logrus.DebugLevel,
		RotatorEnabled: false,
		Rotator:        rotator,
		StdErrEnabled:  true,
	}

	pool := Pool{
		Idle: Intervals{Values: []int{0, 1}},
		Load: Intervals{Values: []int{2, 3, 4, 5}},
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

	filters := NewDefaultFilters()
	service := &Service{
		Interval:     1000,
		Threshold:    150,
		Filters:      filters,
		Pool:         pool,
		Selection:    selection,
		PinCoresAlgo: pinCoresAlgo,
	}

	config.Log = log0
	config.Service = service

	return config
}

func Load(confPath0 string) (*Config, error) {
	bytes0, err := os.ReadFile(confPath0)
	if err != nil {
		log.Printf("config Load err = %v\n", err)
		return nil, err
	}

	config := NewDefaultConfig()
	err = yaml.Unmarshal(bytes0, config)
	if err != nil {
		log.Printf("config yaml.Unmarshal err = %v\n", err)
		return nil, err
	}

	return config, err
}
