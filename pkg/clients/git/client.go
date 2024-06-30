package git

import (
	"context"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/query_info"

	"go.opentelemetry.io/otel/trace"

	"github.com/rs/zerolog"
)

type Client interface {
	GetLatestRelease(ctx context.Context) (string, query_info.QueryInfo, error)
}

func GetClient(
	config configPkg.GitConfig,
	logger zerolog.Logger,
	tracer trace.Tracer,
) Client {
	if constants.GithubRegexp.Match([]byte(config.Repository)) {
		return NewGithub(config, logger, tracer)
	}

	if constants.GitopiaRegexp.Match([]byte(config.Repository)) {
		return NewGitopia(config, logger, tracer)
	}

	return nil
}
