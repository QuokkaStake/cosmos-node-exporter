package http

import (
	"context"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHttpClientInitFail(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewClient(*logger, "://", tracer)

	err := client.Query(context.Background(), "/", nil)
	require.Error(t, err)
}
