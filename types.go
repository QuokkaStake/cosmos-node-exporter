package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Querier interface {
	Enabled() bool
	Get() ([]prometheus.Collector, []QueryInfo)
	Name() string
}

type Upgrade struct {
	Name          string
	BinaryPresent bool
}

type ReleaseInfo struct {
	Name    string `json:"name"`
	TagName string `json:"tag_name"`
	Message string `json:"message"`
}

type VersionInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type UpgradesPresent map[string]bool

func (u UpgradesPresent) HasUpgrade(upgrade string) bool {
	value, ok := u[upgrade]
	if !ok {
		return false
	}

	return value
}
