package cosmovisor

import (
	configPkg "main/pkg/config"
	cosmovisorPkg "main/pkg/cosmovisor"
	"main/pkg/metrics"
	"main/pkg/query_info"

	"github.com/rs/zerolog"
)

type Querier struct {
	Config     configPkg.NodeConfig
	Logger     zerolog.Logger
	Cosmovisor *cosmovisorPkg.Cosmovisor
}

func NewQuerier(
	logger zerolog.Logger,
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

func (v *Querier) Get() ([]metrics.MetricInfo, []query_info.QueryInfo) {
	queryInfo := query_info.QueryInfo{
		Module:  "cosmovisor",
		Action:  "get_cosmovisor_version",
		Success: false,
	}

	cosmovisorVersion, err := v.Cosmovisor.GetCosmovisorVersion()
	if err != nil {
		v.Logger.Err(err).Msg("Could not get Cosmovisor version")
		return []metrics.MetricInfo{}, []query_info.QueryInfo{queryInfo}
	}

	queryInfo.Success = true

	return []metrics.MetricInfo{
		{
			MetricName: metrics.MetricNameCosmovisorVersion,
			Labels:     map[string]string{"version": cosmovisorVersion},
			Value:      1,
		},
	}, []query_info.QueryInfo{queryInfo}
}
