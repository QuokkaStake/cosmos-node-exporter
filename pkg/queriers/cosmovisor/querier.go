package cosmovisor

import (
	configPkg "main/pkg/config"
	"main/pkg/constants"
	cosmovisorPkg "main/pkg/cosmovisor"
	"main/pkg/query_info"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type Querier struct {
	Config     configPkg.NodeConfig
	Logger     zerolog.Logger
	Cosmovisor *cosmovisorPkg.Cosmovisor
}

func NewQuerier(
	config configPkg.NodeConfig,
	logger *zerolog.Logger,
	cosmovisor *cosmovisorPkg.Cosmovisor,
) *Querier {
	return &Querier{
		Config:     config,
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
		[]string{"node", "version"},
	)

	cosmovisorVersionGauge.
		With(prometheus.Labels{"node": v.Config.Name, "version": cosmovisorVersion}).
		Set(1)

	return []prometheus.Collector{cosmovisorVersionGauge}, []query_info.QueryInfo{queryInfo}
}
