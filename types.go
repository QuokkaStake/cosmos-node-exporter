package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Querier interface {
	Enabled() bool
	Get() []prometheus.Collector
	Name() string
}
