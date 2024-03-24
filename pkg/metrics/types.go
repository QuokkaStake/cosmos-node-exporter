package metrics

type MetricName string

const (
	MetricNameCosmovisorVersion     = "cosmovisor_version"
	MetricNameCatchingUp            = "catching_up"
	MetricNameTimeSinceLatestBlock  = "time_since_latest_block"
	MetricNameNodeInfo              = "node_info"
	MetricNameTendermintVersion     = "tendermint_version"
	MetricNameVotingPower           = "voting_power"
	MetricNameRemoteVersion         = "remote_version"
	MetricNameLocalVersion          = "local_version"
	MetricNameIsLatest              = "is_latest"
	MetricNameUpgradeComing         = "upgrade_coming"
	MetricNameUpgradeInfo           = "upgrade_info"
	MetricNameUpgradeHeight         = "upgrade_height"
	MetricNameUpgradeEstimatedTime  = "upgrade_estimated_time"
	MetricNameUpgradeBinaryPresent  = "upgrade_binary_present"
	MetricNameAppVersion            = "version"
	MetricNameQuerySuccessful       = "query_successful"
	MetricNameQuerierEnabled        = "querier_enabled"
	MetricNameStartTime             = "start_time"
	MetricNameMinimumGasPricesCount = "minimum_gas_prices_count"
	MetricNameMinimumGasPrice       = "minimum_gas_price"
)

type MetricInfo struct {
	MetricName MetricName
	Labels     map[string]string
	Value      float64
}
