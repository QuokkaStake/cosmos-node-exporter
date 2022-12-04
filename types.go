package main

import (
	"time"

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

type QueryInfo struct {
	Action  string
	Success bool
}

type UpgradePlan struct {
	Name   string    `json:"name"`
	Time   time.Time `json:"time"`
	Height int64     `json:"height"`
	Info   string    `json:"info"`
}
