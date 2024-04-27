package main

import (
	_ "embed"
	"runtime/debug"
)

var (
	version = "dev"
	commit  = commitNone
	date    = "unknown"
	builtBy = "manual"
)

const commitNone = "none"

func GetVersion0() (string, string) {
	version0 := GetVersion()
	gitHash := GetGitHash()
	return version0, gitHash
}

func GetVersion() string {
	return version
}

func GetGitHash() string {
	if commit != commitNone {
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
