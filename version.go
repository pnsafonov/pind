package main

import (
	_ "embed"
	"runtime/debug"
	"strings"
)

const (
	dev = "dev"
)

// GetVersion1 - return numeric version like 1.0.2 or dev
func GetVersion1() string {
	version0 := GetVersion()
	if version0 == dev {
		return dev
	}

	l0 := len(version0)
	in0 := strings.Index(version0, "v")
	if in0 == -1 || in0 >= l0 {
		return version0
	}

	in1 := in0 + 1
	return version0[in1:]
}

func GetVersion0() (string, string) {
	version0 := GetVersion()
	gitHash := GetGitHash()
	return version0, gitHash
}

func GetVersion() string {
	return version
}

func GetGitHash() string {
	if commit != "none" {
		return commit
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return ""
}
