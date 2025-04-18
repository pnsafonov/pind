package config

import (
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
)

const (
	Phys    = "phys"    // pin to physical cores, ignore hyper threading
	Logical = "logical" // default pool type, all cores types
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
	Interval          int           `yaml:"interval"` // ms
	Threshold         float64       `yaml:"threshold"`
	IdleOverwork      float64       `yaml:"idle_overwork"`
	Filters0          []*ProcFilter `yaml:"filters0"`
	Filters1          []*ProcFilter `yaml:"filters1"`
	FiltersAlwaysIdle []*ProcFilter `yaml:"filters_always_idle"`
	Pool              Pool          `yaml:"pool"`
	Selection         Selection     `yaml:"selection"`
	PinCoresAlgo      *PinCoresAlgo `yaml:"pin_cores_algo"`
	Ignore            *Ignore       `yaml:"ignore"`
	HttpApi           *HttpApi      `yaml:"http_api"`
	Monitoring        *Monitoring   `yaml:"monitoring"`
}

type Pool struct {
	Idle     Intervals
	Load     Intervals
	LoadType string  `yaml:"load_type"`
	PinMode  PinMode `yaml:"pin_mode"`
}

func (x *Pool) GetPinModeStr() string {
	if x.PinMode == PinModeNormal {
		return "normal"
	}
	return "delayed"
}

type ProcFilter struct {
	Patterns []string `yaml:"patterns" json:"patterns"`
}

type Selection struct {
	Patterns []string `yaml:"patterns" json:"patterns"`
}

type Ignore struct {
	Patterns []string `yaml:"patterns" json:"patterns"`
}

type PinCoresAlgo struct {
	Selected    int `yaml:"selected_cores_count"`
	NotSelected int `yaml:"not_selected_cores_count"`
}

type HttpApi struct {
	Enabled bool   `yaml:"enabled"`
	Listen  string `yaml:"listen"`
}

type Monitoring struct {
	Enabled bool   `yaml:"enabled"`
	Listen  string `yaml:"listen"`
}

func NewDefaultFilters0() []*ProcFilter {
	filter0 := &ProcFilter{
		Patterns: []string{"/usr/bin/kvm"},
	}
	filter1 := &ProcFilter{
		Patterns: []string{"/usr/bin/qemu-system-x86_64"},
	}
	filters := []*ProcFilter{filter0, filter1}
	return filters
}

func NewDefaultFilters1() []*ProcFilter {
	filter0 := &ProcFilter{
		Patterns: []string{"deb-2"},
	}
	filter1 := &ProcFilter{
		Patterns: []string{"deb-3"},
	}
	filters := []*ProcFilter{filter0, filter1}
	return filters
}

func NewDefaultFilters2(patternsAny []string) []*ProcFilter {
	l0 := len(patternsAny)
	result := make([]*ProcFilter, 0, l0)
	for _, pattern := range patternsAny {
		result = append(result, &ProcFilter{
			Patterns: []string{pattern},
		})
	}
	return result
}

func NewDefaultConfig(isService bool) *Config {
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
		RotatorEnabled: isService,
		Rotator:        rotator,
		StdErrEnabled:  true,
	}

	pool := Pool{
		//Idle: Intervals{Values: []int{0, 1}},
		//Load: Intervals{Values: []int{2, 3, 4, 5}},
		//Load: Intervals{Values: []int{2, 3, 4, 5, 6}},
		LoadType: Phys,
		PinMode:  PinModeNormal,
	}

	selection := Selection{
		Patterns: []string{"CPU", "/KVM"},
	}

	pinCoresAlgo := &PinCoresAlgo{
		Selected:    1, // 1 core per thread
		NotSelected: 2, // 2 cores for other threads
	}

	ignore := &Ignore{
		Patterns: []string{"iou-wrk-"},
	}

	httpApi := &HttpApi{
		Enabled: true,
		Listen:  "0.0.0.0:10331",
	}

	monitoring := &Monitoring{
		Enabled: true,
		Listen:  "0.0.0.0:9091",
	}

	//filters0 := NewDefaultFilters0()
	//filters1 := NewDefaultFilters1()
	var filters0 []*ProcFilter
	var filters1 []*ProcFilter
	service := &Service{
		Interval:     1000,
		Threshold:    150,
		IdleOverwork: 80,
		Filters0:     filters0,
		Filters1:     filters1,
		Pool:         pool,
		Selection:    selection,
		PinCoresAlgo: pinCoresAlgo,
		Ignore:       ignore,
		HttpApi:      httpApi,
		Monitoring:   monitoring,
	}

	config.Log = log0
	config.Service = service

	return config
}

func Load(confPath0 string, isService bool) (*Config, error) {
	bytes0, err := os.ReadFile(confPath0)
	if err != nil {
		log.Errorf("Load, os.ReadFile err = %v\n", err)
		return nil, err
	}

	config := NewDefaultConfig(isService)
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
