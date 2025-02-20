package config

import (
	"github.com/sirupsen/logrus"
	"reflect"
	"testing"
)

func TestLog0(t *testing.T) {
	path0 := "tests/pind0.yml"
	config, err := Load(path0, false)
	if err != nil {
		t.FailNow()
	}
	if config.Log.Level != logrus.WarnLevel {
		t.FailNow()
	}
}

func TestLog1(t *testing.T) {
	path0 := "tests/pind1.yml"
	config, err := Load(path0, false)
	if err != nil {
		t.FailNow()
	}
	if config.Log.Rotator.MaxSize != 2 {
		t.FailNow()
	}

	if config.Service.Pool.PinMode != PinModeNormal {
		t.FailNow()
	}
}

func TestLog2(t *testing.T) {
	path0 := "tests/pind2.yml"
	config, err := Load(path0, false)
	if err != nil {
		t.FailNow()
	}
	if config.Service.Interval != 1001 {
		t.FailNow()
	}

	if !reflect.DeepEqual(config.Service.Pool.Idle.Values, []int{124, 125, 126, 127}) {
		t.FailNow()
	}
	if !reflect.DeepEqual(config.Service.Pool.Load.Values, []int{32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47}) {
		t.FailNow()
	}

	if !reflect.DeepEqual(config.Service.Filters0[0].Patterns, []string{"/usr/bin/kvm"}) {
		t.FailNow()
	}
	if !reflect.DeepEqual(config.Service.Filters1[0].Patterns, []string{"qemu", "deb11-1"}) {
		t.FailNow()
	}
	if !reflect.DeepEqual(config.Service.Filters1[1].Patterns, []string{"qemu", "deb11-2"}) {
		t.FailNow()
	}

	if !reflect.DeepEqual(config.Service.FiltersAlwaysIdle[0].Patterns, []string{"node_exporter"}) {
		t.FailNow()
	}

	if config.Service.Pool.PinMode != PinModeDelayed {
		t.FailNow()
	}

}
