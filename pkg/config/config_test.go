package config

import (
	"github.com/sirupsen/logrus"
	"testing"
)

func TestLog0(t *testing.T) {
	path0 := "tests/pind0.yml"
	config, err := Load(path0)
	if err != nil {
		t.FailNow()
	}
	if config.Log.Level != logrus.WarnLevel {
		t.FailNow()
	}
}

func TestLog1(t *testing.T) {
	path0 := "tests/pind1.yml"
	config, err := Load(path0)
	if err != nil {
		t.FailNow()
	}
	if config.Log.Rotator.MaxSize != 2 {
		t.FailNow()
	}
}
