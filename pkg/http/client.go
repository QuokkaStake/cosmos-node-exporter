package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"

	"github.com/rs/zerolog"
)

type Client struct {
	Logger zerolog.Logger
	Host   string
	Tracer trace.Tracer
}

func NewClient(logger zerolog.Logger, host string, tracer trace.Tracer) *Client {
	return &Client{
		Logger: logger.With().Str("component", "http_client").Logger(),
		Host:   host,
		Tracer: tracer,
	}
}

func (c Client) Query(ctx context.Context, relativeUrl string, output interface{}) error {
	childCtx, span := c.Tracer.Start(ctx, "HTTP request")
	defer span.End()

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	fullUrl := fmt.Sprintf("%s%s", c.Host, relativeUrl)
	req, err := http.NewRequestWithContext(childCtx, http.MethodGet, fullUrl, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "cosmos-node-exporter")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return json.NewDecoder(res.Body).Decode(&output)
}
