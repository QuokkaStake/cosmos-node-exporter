package types

import (
	"main/pkg/query_info"

	"github.com/prometheus/client_golang/prometheus"
)

type Querier interface {
	Enabled() bool
	Get() ([]prometheus.Collector, []query_info.QueryInfo)
	Name() string
}

type Upgrade struct {
	Name          string
	BinaryPresent bool
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
