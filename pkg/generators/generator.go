package generators

import (
	"main/pkg/fetchers"
	"main/pkg/metrics"
)

type Generator interface {
	Get(state fetchers.State) []metrics.MetricInfo
}
