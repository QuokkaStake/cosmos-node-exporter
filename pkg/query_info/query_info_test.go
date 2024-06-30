package query_info

import (
	"main/pkg/constants"
	"testing"

	"github.com/stretchr/testify/assert"

	metricsPkg "main/pkg/metrics"
)

func TestQueryInfo(t *testing.T) {
	t.Parallel()

	input := map[string]map[string][]QueryInfo{
		"node": {
			"querier": []QueryInfo{{
				Module:  constants.ModuleTendermint,
				Action:  constants.ActionTendermintGetNodeStatus,
				Success: true,
			}},
		},
	}
	metrics := GetQueryInfoMetrics(input)
	assert.Len(t, metrics, 1)

	metric := metrics[0]
	assert.Equal(t, metricsPkg.MetricNameQuerySuccessful, metric.MetricName)
	assert.Equal(t, map[string]string{
		"node":    "node",
		"querier": "querier",
		"module":  "tendermint",
		"action":  "get_node_status",
	}, metric.Labels)
	assert.InDelta(t, 1, metric.Value, 0.01)
}
