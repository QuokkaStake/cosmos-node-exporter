package logger_test

import (
	"main/pkg/config"
	loggerPkg "main/pkg/logger"
	"testing"

	"gopkg.in/guregu/null.v4"

	"github.com/stretchr/testify/require"
)

func TestGetDefaultLogger(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetDefaultLogger()
	require.NotNil(t, logger)
}

func TestGetLoggerInvalidLogLevel(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	loggerPkg.GetLogger(config.LogConfig{LogLevel: "invalid"})
}

func TestGetLoggerValidPlain(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetLogger(config.LogConfig{LogLevel: "info"})
	require.NotNil(t, logger)
}

func TestGetLoggerValidJSON(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetLogger(config.LogConfig{LogLevel: "info", JSONOutput: null.BoolFrom(true)})
	require.NotNil(t, logger)
}

func TestGetLoggerNop(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	require.NotNil(t, logger)
}
