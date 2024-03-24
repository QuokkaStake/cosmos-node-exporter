package metrics

import (
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/logger"

	"github.com/prometheus/client_golang/prometheus"
)

type Manager struct {
	Config           *configPkg.Config
	NodeCollectors   map[MetricName]*prometheus.GaugeVec
	GlobalCollectors map[MetricName]*prometheus.GaugeVec
}

func NewManager(config *configPkg.Config) *Manager {
	nodeCollectors := map[MetricName]*prometheus.GaugeVec{
		MetricNameCosmovisorVersion: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "cosmovisor_version",
				Help: "Cosmovisor version",
			},
			[]string{"node", "version"},
		),

		MetricNameCatchingUp: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "catching_up",
				Help: "Is node catching up?",
			},
			[]string{"node"},
		),

		MetricNameTimeSinceLatestBlock: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "time_since_latest_block",
				Help: "Time since latest block, in seconds",
			},
			[]string{"node"},
		),

		MetricNameVotingPower: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "voting_power",
				Help: "Node voting power",
			},
			[]string{"node"},
		),

		MetricNameNodeInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "node_info",
				Help: "Node info (moniker, network, etc.), always 1",
			},
			[]string{"node", "moniker", "chain"},
		),

		MetricNameTendermintVersion: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "tendermint_version",
				Help: "Tendermint/CometBFT version, always 1",
			},
			[]string{"node", "version"},
		),

		MetricNameRemoteVersion: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "remote_version",
				Help: "Latest version from Github",
			},
			[]string{"node", "version"},
		),

		MetricNameLocalVersion: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "local_version",
				Help: "Fullnode local version",
			},
			[]string{"node", "version"},
		),

		MetricNameIsLatest: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "is_latest",
				Help: "Is the fullnode using the same or latest version?",
			},
			[]string{"node", "local_version", "remote_version"},
		),

		MetricNameUpgradeComing: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "upgrade_coming",
				Help: "Is future upgrade planned?",
			},
			[]string{"node"},
		),

		MetricNameUpgradeInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "upgrade_info",
				Help: "Future upgrade info",
			},
			[]string{"node", "name", "info"},
		),

		MetricNameUpgradeHeight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "upgrade_height",
				Help: "Future upgrade height",
			},
			[]string{"node", "name", "info"},
		),

		MetricNameUpgradeEstimatedTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "upgrade_estimated_time",
				Help: "Estimated upgrade time, as Unix timestamp",
			},
			[]string{"node", "name", "info"},
		),

		MetricNameUpgradeBinaryPresent: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "upgrade_binary_present",
				Help: "Is upgrade binary present?",
			},
			[]string{"node", "name", "info"},
		),

		MetricNameQuerierEnabled: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "querier_enabled",
				Help: "Is querier enabled?",
			},
			[]string{"node", "querier"},
		),

		MetricNameMinimumGasPricesCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "minimum_gas_prices_count",
				Help: "Amount of minimum_gas_prices on a node",
			},
			[]string{"node"},
		),

		MetricNameMinimumGasPrice: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "minimum_gas_price",
				Help: "Minimum gas price for a specific denom",
			},
			[]string{"node", "denom"},
		),

		MetricNameCosmosSdkVersion: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "cosmos_sdk_version",
				Help: "cosmos-sdk version",
			},
			[]string{"node", "version"},
		),

		MetricNameRunningAppVersion: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "running_app_version",
				Help: "Version of the app that's running",
			},
			[]string{"node", "version", "name", "app_name", "git_commit"},
		),

		MetricNameGoVersion: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "go_version",
				Help: "Go version that was used to build the app",
			},
			[]string{"node", "version", "tags"},
		),
	}

	globalCollectors := map[MetricName]*prometheus.GaugeVec{
		MetricNameAppVersion: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "version",
				Help: "The app info and version.",
			},
			[]string{"version"},
		),

		MetricNameQuerySuccessful: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "query_successful",
				Help: "Was query successful?",
			},
			[]string{"node", "querier", "module", "action"},
		),

		MetricNameStartTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "start_time",
				Help: "Unix timestamp on when the app was started. Useful for annotations.",
			},
			[]string{},
		),
	}

	return &Manager{
		NodeCollectors:   nodeCollectors,
		GlobalCollectors: globalCollectors,
		Config:           config,
	}
}

func (m *Manager) CollectMetrics(
	nodeMetricsInfo map[string][]MetricInfo,
	globalMetrics []MetricInfo,
) *prometheus.Registry {
	registry := prometheus.NewRegistry()
	for _, collector := range m.NodeCollectors {
		collector.Reset()
		registry.MustRegister(collector)
	}

	for _, collector := range m.GlobalCollectors {
		collector.Reset()
		registry.MustRegister(collector)
	}

	for nodeName, nodeMetrics := range nodeMetricsInfo {
		for _, metricInfo := range nodeMetrics {
			collector, ok := m.NodeCollectors[metricInfo.MetricName]
			if !ok {
				logger.GetDefaultLogger().
					Fatal().
					Str("name", string(metricInfo.MetricName)).
					Msg("Could not find collector by name!")
			}

			metricInfo.Labels["node"] = nodeName

			collector.With(metricInfo.Labels).Set(metricInfo.Value)
		}
	}

	for _, metricInfo := range globalMetrics {
		collector, ok := m.GlobalCollectors[metricInfo.MetricName]
		if !ok {
			logger.GetDefaultLogger().
				Fatal().
				Str("name", string(metricInfo.MetricName)).
				Msg("Could not find collector by name!")
		}

		collector.With(metricInfo.Labels).Set(metricInfo.Value)
	}

	return registry
}
