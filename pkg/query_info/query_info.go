package query_info

import (
	"main/pkg/constants"
	"main/pkg/utils"

	"github.com/prometheus/client_golang/prometheus"
)

type QueryInfo struct {
	Module  string
	Action  string
	Success bool
}

func GetQueryInfoMetrics(allQueries map[string]map[string][]QueryInfo) []prometheus.Collector {
	querySuccessfulGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "query_successful",
			Help: "Was query successful?",
		},
		[]string{"node", "querier", "module", "action"},
	)

	for node, nodeQueryInfos := range allQueries {
		for name, queryInfos := range nodeQueryInfos {
			for _, queryInfo := range queryInfos {
				querySuccessfulGauge.
					With(prometheus.Labels{
						"node":    node,
						"querier": name,
						"module":  queryInfo.Module,
						"action":  queryInfo.Action,
					}).
					Set(utils.BoolToFloat64(queryInfo.Success))
			}
		}
	}

	return []prometheus.Collector{querySuccessfulGauge}
}
