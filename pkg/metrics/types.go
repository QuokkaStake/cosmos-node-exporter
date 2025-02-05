package metrics

type MetricName string

const (
	MetricNameCosmovisorVersion     MetricName = "cosmovisor_version"
	MetricNameCatchingUp            MetricName = "catching_up"
	MetricNameLatestBlockHeight     MetricName = "latest_block_height"
	MetricNameLatestBlockTime       MetricName = "latest_block_time"
	MetricNameNodeInfo              MetricName = "node_info"
	MetricNameTendermintVersion     MetricName = "tendermint_version"
	MetricNameVotingPower           MetricName = "voting_power"
	MetricNameRemoteVersion         MetricName = "remote_version"
	MetricNameLocalVersion          MetricName = "local_version"
	MetricNameIsLatest              MetricName = "is_latest"
	MetricNameUpgradeComing         MetricName = "upgrade_coming"
	MetricNameUpgradeInfo           MetricName = "upgrade_info"
	MetricNameUpgradeHeight         MetricName = "upgrade_height"
	MetricNameUpgradeEstimatedTime  MetricName = "upgrade_estimated_time"
	MetricNameUpgradeBinaryPresent  MetricName = "upgrade_binary_present"
	MetricNameAppVersion            MetricName = "version"
	MetricNameQuerySuccessful       MetricName = "query_successful"
	MetricNameQuerierEnabled        MetricName = "querier_enabled"
	MetricNameStartTime             MetricName = "start_time"
	MetricNameMinimumGasPricesCount MetricName = "minimum_gas_prices_count"
	MetricNameMinimumGasPrice       MetricName = "minimum_gas_price"
	MetricNameCosmosSdkVersion      MetricName = "cosmos_sdk_version"
	MetricNameRunningAppVersion     MetricName = "running_app_version"
	MetricNameGoVersion             MetricName = "go_version"
	MetricNameHaltHeight            MetricName = "halt_height"
	MetricNameNotExisting           MetricName = "not_existing" // for tests only
)

type MetricInfo struct {
	MetricName MetricName
	Labels     map[string]string
	Value      float64
}
