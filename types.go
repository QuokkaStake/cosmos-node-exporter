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
}

type VersionInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type QueryInfo struct {
	Action  string
	Success bool
}

type UpgradePlan struct {
}
