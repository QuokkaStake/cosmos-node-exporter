package git

import (
	configPkg "main/pkg/config"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetClientGithub(t *testing.T) {
	t.Parallel()

	config := configPkg.GitConfig{Repository: "https://github.com/QuokkaStake/cosmos-node-exporter"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := GetClient(config, *logger, tracer)

	require.NotNil(t, client)
	require.IsType(t, &Github{}, client)
}

func TestGetClientGitopia(t *testing.T) {
	t.Parallel()

	config := configPkg.GitConfig{Repository: "gitopia://gitopia/gitopia"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := GetClient(config, *logger, tracer)

	require.NotNil(t, client)
	require.IsType(t, &Gitopia{}, client)
}

func TestGetClientEmpty(t *testing.T) {
	t.Parallel()

	config := configPkg.GitConfig{}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := GetClient(config, *logger, tracer)

	require.Nil(t, client)
}
