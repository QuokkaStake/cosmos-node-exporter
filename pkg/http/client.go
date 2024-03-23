package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type Client struct {
	Logger zerolog.Logger
	Host   string
}

func NewClient(logger zerolog.Logger, host string) *Client {
	return &Client{
		Logger: logger.With().Str("component", "http_client").Logger(),
		Host:   host,
	}
}

func (c Client) Query(relativeUrl string, output interface{}) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	fullUrl := fmt.Sprintf("%s%s", c.Host, relativeUrl)
	req, err := http.NewRequest(http.MethodGet, fullUrl, nil)

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
