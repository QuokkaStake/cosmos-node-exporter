package git

import (
	configPkg "main/pkg/config"

	"github.com/rs/zerolog"
)

type Client interface {
	GetLatestRelease() (string, error)
}

func GetClient(config *configPkg.Config, logger *zerolog.Logger) Client {
	if config.GitConfig.Repository != "" {
		return NewGithub(config, logger)
	}

	return nil
}
