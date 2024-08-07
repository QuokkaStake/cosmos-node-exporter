package pkg

import (
	"context"
	configPkg "main/pkg/config"
	"main/pkg/fs"
	"main/pkg/logger"
	"main/pkg/metrics"
	"main/pkg/queriers/app"
	"main/pkg/queriers/uptime"
	"main/pkg/query_info"
	"main/pkg/tracing"
	"main/pkg/types"
	"net/http"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/google/uuid"

	"go.opentelemetry.io/otel/attribute"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type App struct {
	Logger         zerolog.Logger
	Config         *configPkg.Config
	NodeHandlers   []*NodeHandler
	MetricsManager *metrics.Manager
	GlobalQueriers []types.Querier
	Tracer         trace.Tracer
	Server         *http.Server
}

func NewApp(
	filesystem fs.FS,
	configPath string,
	version string,
) *App {
	appConfig, err := configPkg.GetConfig(filesystem, configPath)
	if err != nil {
		logger.GetDefaultLogger().Panic().Err(err).Msg("Could not load config")
	}

	if err = appConfig.Validate(); err != nil {
		logger.GetDefaultLogger().Panic().Err(err).Msg("Provided config is invalid!")
	}

	log := logger.GetLogger(appConfig.LogConfig)
	tracer := tracing.InitTracer(appConfig.TracingConfig, version)

	nodeHandlers := make([]*NodeHandler, len(appConfig.NodeConfigs))

	for index, nodeConfig := range appConfig.NodeConfigs {
		nodeHandlers[index] = NewNodeHandler(log, nodeConfig, tracer)
	}

	globalQueriers := []types.Querier{
		app.NewQuerier(version),
		uptime.NewQuerier(),
	}

	server := &http.Server{Addr: appConfig.ListenAddress, Handler: nil}

	return &App{
		Logger:         log.With().Str("component", "app").Logger(),
		Config:         appConfig,
		NodeHandlers:   nodeHandlers,
		MetricsManager: metrics.NewManager(),
		GlobalQueriers: globalQueriers,
		Tracer:         tracer,
		Server:         server,
	}
}

func (a *App) Start() {
	otelHandler := otelhttp.NewHandler(http.HandlerFunc(a.HandleRequest), "prometheus")
	handler := http.NewServeMux()
	handler.Handle("/metrics", otelHandler)
	handler.HandleFunc("/healthcheck", a.Healthcheck)
	a.Server.Handler = handler

	a.Logger.Info().Str("addr", a.Config.ListenAddress).Msg("Listening")

	err := a.Server.ListenAndServe()
	if err != nil {
		a.Logger.Panic().Err(err).Msg("Could not start application")
	}
}

func (a *App) Stop() {
	a.Logger.Info().Str("addr", a.Config.ListenAddress).Msg("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = a.Server.Shutdown(ctx)
}

func (a *App) HandleRequest(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()

	span := trace.SpanFromContext(r.Context())
	span.SetAttributes(attribute.String("request-id", requestID))
	rootSpanCtx := r.Context()

	sublogger := a.Logger.With().
		Str("request-id", requestID).
		Logger()

	defer span.End()

	requestStart := time.Now()

	allResults := make(map[string][]metrics.MetricInfo)
	allQueries := make(map[string]map[string][]query_info.QueryInfo)

	globalResults := make([]metrics.MetricInfo, 0)

	var wg sync.WaitGroup
	var mutex sync.Mutex

	// Global handlers
	for _, globalQuerier := range a.GlobalQueriers {
		wg.Add(1)
		go func(querier types.Querier) {
			_, fetcherSpan := a.Tracer.Start(
				rootSpanCtx,
				"Querier "+querier.Name(),
				trace.WithAttributes(attribute.String("querier", querier.Name())),
			)
			defer fetcherSpan.End()

			defer wg.Done()

			results, _ := querier.Get(rootSpanCtx)
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

			results, queries := nodeHandler.Process(rootSpanCtx)
			mutex.Lock()
			allResults[nodeHandler.Config.Name] = results
			allQueries[nodeHandler.Config.Name] = queries
			mutex.Unlock()
		}(nodeHandler)
	}

	wg.Wait()

	globalResults = append(globalResults, query_info.GetQueryInfoMetrics(allQueries)...)

	registry := a.MetricsManager.CollectMetrics(allResults, globalResults)

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)

	sublogger.Info().
		Str("method", http.MethodGet).
		Str("endpoint", "/metrics").
		Float64("request-time", time.Since(requestStart).Seconds()).
		Msg("Request processed")
}

func (a *App) Healthcheck(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("ok"))
}
