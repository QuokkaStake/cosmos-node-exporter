package pkg

import (
	"context"
	"main/assets"
	configPkg "main/pkg/config"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

//nolint:paralleltest // disabled due to httpmock usage
func TestNodeHandler(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/status",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("status.json")),
	)

	config := configPkg.NodeConfig{
		TendermintConfig: configPkg.TendermintConfig{
			Enabled:       null.BoolFrom(true),
			Address:       "https://example.com",
			QueryUpgrades: null.BoolFrom(false),
		},
		CosmovisorConfig: configPkg.CosmovisorConfig{
			Enabled:         null.BoolFrom(true),
			ChainFolder:     ".",
			ChainBinaryName: "appd",
			CosmovisorPath:  "./not-found",
		},
		GrpcConfig: configPkg.GrpcConfig{
			Enabled: null.BoolFrom(true),
			Address: "invalid",
		},
	}

	logger := loggerPkg.GetDefaultLogger()
	tracer := tracing.InitNoopTracer()
	handler := NewNodeHandler(logger, config, tracer)

	metrics, queryInfos := handler.Process(context.Background())
	assert.NotEmpty(t, queryInfos)
	assert.NotEmpty(t, metrics)
}
