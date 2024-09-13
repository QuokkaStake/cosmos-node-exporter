package constants

import (
	"regexp"
	"time"
)

type Module string
type Action string

type FetcherName string

const (
	MetricsPrefix                  = "cosmos_node_exporter_"
	UncachedGithubQueryTime        = 120 * time.Second
	BlocksBehindToCheck            = 1000
	ModuleCosmovisor        Module = "cosmovisor"
	ModuleTendermint        Module = "tendermint"
	ModuleGit               Module = "git"
	ModuleGrpc              Module = "grpc"

	ActionCosmovisorGetVersion           Action = "get_version"
	ActionCosmovisorGetCosmovisorVersion Action = "get_cosmovisor_version"
	ActionCosmovisorGetUpgrades          Action = "get_upgrades"
	ActionGitGetLatestRelease            Action = "get_latest_release"
	ActionTendermintGetNodeStatus        Action = "get_node_status"
	ActionTendermintGetUpgradePlan       Action = "get_upgrade_plan"
	ActionTendermintGetBlockTime         Action = "get_block_time"
	ActionGrpcGetNodeConfig              Action = "get_node_config"
	ActionGrpcGetNodeInfo                Action = "get_node_info"

	FetcherNameNodeStatus         FetcherName = "node_status"
	FetcherNameCosmovisorVersion  FetcherName = "cosmovisor_version"
	FetcherNameNodeConfig         FetcherName = "node_config"
	FetcherNameNodeInfo           FetcherName = "node_info"
	FetcherNameUptime             FetcherName = "uptime"
	FetcherNameAppVersion         FetcherName = "app_version"
	FetcherNameRemoteVersion      FetcherName = "remote_version"
	FetcherNameLocalVersion       FetcherName = "local_version"
	FetcherNameUpgrades           FetcherName = "upgrades"
	FetcherNameBlockTime          FetcherName = "block_time"
	FetcherNameCosmovisorUpgrades FetcherName = "cosmovisor_upgrades"
)

var (
	GithubRegexp  = regexp.MustCompile("https://github.com/(?P<Org>[a-zA-Z0-9-].*)/(?P<Repo>[a-zA-Z0-9-].*)")
	GitopiaRegexp = regexp.MustCompile("gitopia://(?P<Org>[a-zA-Z0-9-].*)/(?P<Repo>[a-zA-Z0-9-].*)")
	ColorsRegexp  = regexp.MustCompile("[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))")
)
