package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	metricsPkg "main/pkg/metrics"

	"github.com/cosmos/cosmos-sdk/client/grpc/node"
	cosmosTypes "github.com/cosmos/cosmos-sdk/types"
)

type NodeConfigGenerator struct{}

func NewNodeConfigGenerator() *NodeConfigGenerator {
	return &NodeConfigGenerator{}
}

func (g *NodeConfigGenerator) Get(state fetchers.State) []metricsPkg.MetricInfo {
	nodeConfig, nodeConfigFound := fetchers.StateGet[*node.ConfigResponse](state, constants.FetcherNameNodeConfig)
	if !nodeConfigFound {
		return []metricsPkg.MetricInfo{}
	}

	coinsParsed, err := cosmosTypes.ParseDecCoins(nodeConfig.MinimumGasPrice)
	if err != nil {
		panic(err)
	}

	metrics := []metricsPkg.MetricInfo{}
	metrics = append(metrics, metricsPkg.MetricInfo{
		MetricName: metricsPkg.MetricNameMinimumGasPricesCount,
		Labels:     map[string]string{},
		Value:      float64(len(coinsParsed)),
	})

	for _, amount := range coinsParsed {
		metrics = append(metrics, metricsPkg.MetricInfo{
			MetricName: metricsPkg.MetricNameMinimumGasPrice,
			Labels:     map[string]string{"denom": amount.Denom},
			Value:      amount.Amount.MustFloat64(),
		})
	}

	if nodeConfig.HaltHeight > 0 {
		metrics = append(metrics, metricsPkg.MetricInfo{
			MetricName: metricsPkg.MetricNameHaltHeight,
			Labels:     map[string]string{},
			Value:      float64(nodeConfig.HaltHeight),
		})
	}

	return metrics
}
