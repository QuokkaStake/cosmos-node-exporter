package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Querier interface {
	Enabled() bool
	Get() []prometheus.Collector
	Name() string
}

type Upgrade struct {
	Name          string
	BinaryPresent bool
}
