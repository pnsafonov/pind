package config

import "strings"

type PinMode int

const (
	PinModeNormal  = PinMode(0)
	PinModeDelayed = PinMode(1)
)

func (x *PinMode) UnmarshalYAML(unmarshal func(interface{}) error) error {
	str0 := ""
	err := unmarshal(&str0)
	if err != nil {
		return err
	}

	str1 := strings.TrimSpace(str0)
	str2 := strings.ToLower(str1)

	switch str2 {
	case "normal":
		*x = PinModeNormal
	case "delayed":
		*x = PinModeDelayed
	}

	return nil
}
