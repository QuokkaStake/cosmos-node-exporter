package constants

import (
	"regexp"
	"time"
)

type Module string

const (
	MetricsPrefix                  = "cosmos_node_exporter_"
	UncachedGithubQueryTime        = 30 * time.Second
	ModuleCosmovisor        Module = "cosmovisor"
	ModuleTendermint        Module = "tendermint"
	ModuleGit               Module = "git"
)

var (
	GithubRegexp  = regexp.MustCompile("https://github.com/(?P<Org>[a-zA-Z0-9-].*)/(?P<Repo>[a-zA-Z0-9-].*)")
	GitopiaRegexp = regexp.MustCompile("gitopia://(?P<Org>[a-zA-Z0-9-].*)/(?P<Repo>[a-zA-Z0-9-].*)")
)
