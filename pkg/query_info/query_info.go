package query_info

import (
	"main/pkg/constants"
	"main/pkg/metrics"
	"main/pkg/utils"
)

type QueryInfo struct {
	Module  constants.Module
	Action  string
	Success bool
}

func GetQueryInfoMetrics(allQueries map[string]map[string][]QueryInfo) []metrics.MetricInfo {
	metricsInfos := []metrics.MetricInfo{}

	for node, nodeQueryInfos := range allQueries {
		for name, queryInfos := range nodeQueryInfos {
			for _, queryInfo := range queryInfos {
				metricsInfos = append(metricsInfos, metrics.MetricInfo{
					MetricName: metrics.MetricNameQuerySuccessful,
					Labels: map[string]string{
						"node":    node,
						"querier": name,
						"module":  string(queryInfo.Module),
						"action":  queryInfo.Action,
					},
					Value: utils.BoolToFloat64(queryInfo.Success),
				})
			}
		}
	}

	return metricsInfos
}
