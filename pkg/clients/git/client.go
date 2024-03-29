package git

import (
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/query_info"

	"github.com/rs/zerolog"
)

type Client interface {
	GetLatestRelease() (string, query_info.QueryInfo, error)
}

func GetClient(config configPkg.NodeConfig, logger zerolog.Logger) Client {
	if constants.GithubRegexp.Match([]byte(config.GitConfig.Repository)) {
		return NewGithub(config, logger)
	}

	if constants.GitopiaRegexp.Match([]byte(config.GitConfig.Repository)) {
		return NewGitopia(config, logger)
	}

	return nil
}
