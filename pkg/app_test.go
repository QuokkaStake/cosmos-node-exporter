package pkg

import (
	"io"
	"main/assets"
	"main/pkg/fs"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppConfigFailToLoad(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	filesystem := &fs.TestFS{}
	NewApp(filesystem, "config-not-found.toml", "1.2.3")
}

func TestAppConfigInvalid(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	filesystem := &fs.TestFS{}
	NewApp(filesystem, "config-invalid.toml", "1.2.3")
}

func TestAppConfigValid(t *testing.T) {
	t.Parallel()

	filesystem := &fs.TestFS{}
	app := NewApp(filesystem, "config-valid.toml", "1.2.3")
	assert.NotNil(t, app)
}

//nolint:paralleltest // disabled
func TestAppFailToStart(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	filesystem := &fs.TestFS{}

	app := NewApp(filesystem, "config-invalid-address.toml", "1.2.3")
	app.Start()
}

//nolint:paralleltest // disabled
func TestAppStopOperation(_ *testing.T) {
	filesystem := &fs.TestFS{}

	app := NewApp(filesystem, "config-valid.toml", "1.2.3")
	app.Stop()
}

//nolint:paralleltest // disabled
func TestAppLoadConfigOk(t *testing.T) {
	filesystem := &fs.TestFS{}

	app := NewApp(filesystem, "config-valid.toml", "1.2.3")
	go app.Start()

	for {
		request, err := http.Get("http://localhost:9500/healthcheck")
		_ = request.Body.Close()
		if err == nil {
			break
		}

		time.Sleep(time.Millisecond * 100)
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/status",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("status.json")),
	)

	httpmock.RegisterResponder("GET", "http://localhost:9500/healthcheck", httpmock.InitialTransport.RoundTrip)
	httpmock.RegisterResponder("GET", "http://localhost:9500/metrics", httpmock.InitialTransport.RoundTrip)

	response, err := http.Get("http://localhost:9500/metrics")
	require.NoError(t, err)
	require.NotEmpty(t, response)

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)

	err = response.Body.Close()
	require.NoError(t, err)
}
