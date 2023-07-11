package app

import (
	"main/pkg/constants"
	"main/pkg/query_info"

	"github.com/prometheus/client_golang/prometheus"
)

type Querier struct {
	Version string
}

func NewQuerier(version string) *Querier {
	return &Querier{
		Version: version,
	}
}

func (u *Querier) Enabled() bool {
	return true
}

func (u *Querier) Name() string {
	return "app-querier"
}

func (u *Querier) Get() ([]prometheus.Collector, []query_info.QueryInfo) {
	appVersionGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "version",
			Help: "The app info and version.",
		},
		[]string{"version"},
	)

	appVersionGauge.
		With(prometheus.Labels{"version": u.Version}).
		Set(1)

	return []prometheus.Collector{appVersionGauge}, []query_info.QueryInfo{}
}
