package cosmovisor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"main/pkg/constants"
	cosmovisorPkg "main/pkg/cosmovisor"
	"main/pkg/query_info"
)

type Querier struct {
	Logger     zerolog.Logger
	Cosmovisor *cosmovisorPkg.Cosmovisor
}

func NewQuerier(
	logger *zerolog.Logger,
	cosmovisor *cosmovisorPkg.Cosmovisor,
) *Querier {
	return &Querier{
		Logger:     logger.With().Str("component", "cosmovisor_querier").Logger(),
		Cosmovisor: cosmovisor,
	}
}

func (v *Querier) Enabled() bool {
	return v.Cosmovisor != nil
}

func (v *Querier) Name() string {
	return "cosmovisor-querier"
}

func (v *Querier) Get() ([]prometheus.Collector, []query_info.QueryInfo) {
	queryInfo := query_info.QueryInfo{
		Module:  "cosmovisor",
		Action:  "get_cosmovisor_version",
		Success: false,
	}

	cosmovisorVersion, err := v.Cosmovisor.GetCosmovisorVersion()
	if err != nil {
		v.Logger.Err(err).Msg("Could not get Cosmovisor version")
		return []prometheus.Collector{}, []query_info.QueryInfo{queryInfo}
	}

	queryInfo.Success = true

	cosmovisorVersionGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "cosmovisor_version",
			Help: "Cosmovisor version",
		},
		[]string{"version"},
	)

	cosmovisorVersionGauge.
		With(prometheus.Labels{"version": cosmovisorVersion}).
		Set(1)

	return []prometheus.Collector{cosmovisorVersionGauge}, []query_info.QueryInfo{queryInfo}
}
