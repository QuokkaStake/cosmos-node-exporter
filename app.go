package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

type App struct {
	Logger   zerolog.Logger
	Queriers []Querier
}

func NewApp(
	logger *zerolog.Logger,
	config *Config,
) *App {
	return &App{
		Logger:   logger.With().Str("component", "app").Logger(),
		Queriers: []Querier{},
	}
}

func (a *App) HandleRequest(w http.ResponseWriter, r *http.Request) {
	requestStart := time.Now()

	sublogger := a.Logger.With().
		Str("request-id", uuid.New().String()).
		Logger()

	var wg sync.WaitGroup
	var mu sync.Mutex
	allResults := map[string][]prometheus.Collector{}

	for _, querier := range a.Queriers {
		if !querier.Enabled() {
			continue
		}

		wg.Add(1)
		go func() {
			querierResults := querier.Get()
			mu.Lock()
			allResults[querier.Name()] = querierResults
		}()
	}

	registry := prometheus.NewRegistry()
	for _, querierResults := range allResults {
		for _, result := range querierResults {
			registry.MustRegister(result)
		}
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)

	sublogger.Info().
		Str("method", http.MethodGet).
		Str("endpoint", "/metrics").
		Float64("request-time", time.Since(requestStart).Seconds()).
		Msg("Request processed")
}
