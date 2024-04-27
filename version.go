package main

import (
	_ "embed"
	"encoding/json"
	"runtime/debug"
)

//go:embed version.json
var VersionBytes []byte

type Version struct {
	Version string `json:"version"`
}

func GetVersion1() (string, string) {
	version, _ := GetVersion()
	gitHash := GetGitHash()
	return version, gitHash
}

func GetVersion0() (*Version, error) {
	version := &Version{}
	err := json.Unmarshal(VersionBytes, version)
	if err != nil {
		return nil, err
	}
	return version, nil
}

func GetVersion() (string, error) {
	version, err := GetVersion0()
	if err != nil {
		return "", err
	}
	return version.Version, nil
}

func GetGitHash() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return ""
}
