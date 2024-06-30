package metrics

import (
	"main/pkg/constants"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

func TestMetricsManagerCollectNodeMetricNotFound(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	manager := NewManager()
	metric := MetricInfo{
		MetricName: MetricNameNotExisting,
		Labels:     map[string]string{},
		Value:      1,
	}
	metrics := map[string][]MetricInfo{"node": {metric}}
	manager.CollectMetrics(metrics, []MetricInfo{})
}

func TestMetricsManagerCollectNodeMetric(t *testing.T) {
	t.Parallel()

	manager := NewManager()
	metric := MetricInfo{
		MetricName: MetricNameCatchingUp,
		Labels:     map[string]string{},
		Value:      1,
	}
	metrics := map[string][]MetricInfo{"node": {metric}}
	registry := manager.CollectMetrics(metrics, []MetricInfo{})

	count, err := testutil.GatherAndCount(registry, constants.MetricsPrefix+"catching_up")
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestMetricsManagerCollectGlobalMetricNotFound(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	manager := NewManager()
	metric := MetricInfo{
		MetricName: MetricNameNotExisting,
		Labels:     map[string]string{},
		Value:      1,
	}
	manager.CollectMetrics(map[string][]MetricInfo{}, []MetricInfo{metric})
}

func TestMetricsManagerCollectGlobalMetric(t *testing.T) {
	t.Parallel()

	manager := NewManager()
	metric := MetricInfo{
		MetricName: MetricNameAppVersion,
		Labels:     map[string]string{"version": "1.2.3"},
		Value:      1,
	}
	registry := manager.CollectMetrics(map[string][]MetricInfo{}, []MetricInfo{metric})

	count, err := testutil.GatherAndCount(registry, constants.MetricsPrefix+"version")
	require.NoError(t, err)
	require.Equal(t, 1, count)
}
