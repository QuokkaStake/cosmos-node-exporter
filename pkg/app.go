package pkg

import (
	configPkg "main/pkg/config"
	"main/pkg/queriers/app"
	"main/pkg/query_info"
	"main/pkg/types"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

type App struct {
	Logger         zerolog.Logger
	Config         *configPkg.Config
	NodeHandlers   []*NodeHandler
	GlobalQueriers []types.Querier
}

func NewApp(
	logger *zerolog.Logger,
	config *configPkg.Config,
	version string,
) *App {
	nodeHandlers := make([]*NodeHandler, len(config.NodeConfigs))

	for index, nodeConfig := range config.NodeConfigs {
		nodeHandlers[index] = NewNodeHandler(logger, nodeConfig)
	}

	globalQueriers := []types.Querier{
		app.NewQuerier(version),
	}

	return &App{
		Logger:         logger.With().Str("component", "app").Logger(),
		Config:         config,
		NodeHandlers:   nodeHandlers,
		GlobalQueriers: globalQueriers,
	}
}

func (a *App) HandleRequest(w http.ResponseWriter, r *http.Request) {
	requestStart := time.Now()

	registry := prometheus.NewRegistry()

	allResults := make(map[string]map[string][]prometheus.Collector)
	allQueries := make(map[string]map[string][]query_info.QueryInfo)

	globalResults := make([]prometheus.Collector, 0)

	var wg sync.WaitGroup
	var mutex sync.Mutex

	// Global handlers
	for _, globalQuerier := range a.GlobalQueriers {
		wg.Add(1)
		go func(querier types.Querier) {
			defer wg.Done()

			results, _ := querier.Get()
			mutex.Lock()
			globalResults = append(globalResults, results...)
			mutex.Unlock()
		}(globalQuerier)
	}

	// Per-node handlers
	for _, nodeHandler := range a.NodeHandlers {
		wg.Add(1)

		go func(nodeHandler *NodeHandler) {
			defer wg.Done()

			results, queries := nodeHandler.Process()
			mutex.Lock()
			allResults[nodeHandler.Config.Name] = results
			allQueries[nodeHandler.Config.Name] = queries
			mutex.Unlock()
		}(nodeHandler)
	}

	wg.Wait()

	globalResults = append(globalResults, query_info.GetQueryInfoMetrics(allQueries)...)

	for _, perNodeResults := range allResults {
		for _, nodeQuerierMetrics := range perNodeResults {
			registry.MustRegister(nodeQuerierMetrics...)
		}
	}

	registry.MustRegister(globalResults...)

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)

	a.Logger.Info().
		Str("method", http.MethodGet).
		Str("endpoint", "/metrics").
		Float64("request-time", time.Since(requestStart).Seconds()).
		Msg("Request processed")
}
