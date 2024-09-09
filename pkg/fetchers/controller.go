package fetchers

import (
	"context"
	"main/pkg/constants"
	"main/pkg/query_info"
	"sync"

	"github.com/rs/zerolog"
)

type FetcherProcessStatus string

const (
	FetcherProcessStatusProcessing FetcherProcessStatus = "processing"
	FetcherProcessStatusDone       FetcherProcessStatus = "done"
)

type FetchersStatuses map[constants.FetcherName]FetcherProcessStatus

func (s FetchersStatuses) IsAllDone(fetcherNames []constants.FetcherName) bool {
	for _, fetcherName := range fetcherNames {
		if value, ok := s[fetcherName]; !ok || value != FetcherProcessStatusDone {
			return false
		}
	}

	return true
}

type State map[constants.FetcherName]interface{}

type Controller struct {
	Fetchers Fetchers
	Logger   zerolog.Logger
}

func NewController(
	fetchers Fetchers,
	logger zerolog.Logger,
	chainName string,
) *Controller {
	return &Controller{
		Logger: logger.With().
			Str("component", "controller").
			Str("node", chainName).
			Logger(),
		Fetchers: fetchers,
	}
}

func (c *Controller) Fetch(ctx context.Context) (
	State,
	map[constants.FetcherName][]query_info.QueryInfo,
) {
	data := map[constants.FetcherName]interface{}{}
	queries := map[constants.FetcherName][]query_info.QueryInfo{}
	fetchersStatus := FetchersStatuses{}

	var mutex sync.Mutex
	var wg sync.WaitGroup

	var startAllPendingFetchers func()
	var processFetcher func(fetcher Fetcher)

	startAllPendingFetchers = func() {
		c.Logger.Trace().Msg("Processing all pending fetchers...")

		if fetchersStatus.IsAllDone(c.Fetchers.GetNames()) {
			c.Logger.Trace().Msg("All fetchers are fetched.")
			return
		}

		for _, fetcher := range c.Fetchers {
			mutex.Lock()
			if _, ok := fetchersStatus[fetcher.Name()]; ok {
				c.Logger.Trace().
					Str("name", string(fetcher.Name())).
					Msg("Fetcher is already being processed or is processed, skipping.")
				mutex.Unlock()
				continue
			}

			if !fetchersStatus.IsAllDone(fetcher.Dependencies()) {
				c.Logger.Trace().
					Str("name", string(fetcher.Name())).
					Msg("Fetcher's dependencies are not yet processed, skipping for now.")
				mutex.Unlock()
				continue
			}

			mutex.Unlock()

			wg.Add(1)
			go processFetcher(fetcher)
		}
	}

	processFetcher = func(fetcher Fetcher) {
		if !fetcher.Enabled() {
			c.Logger.Trace().Str("name", string(fetcher.Name())).Msg("Fetcher is disabled, skipping.")

			mutex.Lock()
			fetchersStatus[fetcher.Name()] = FetcherProcessStatusDone
			mutex.Unlock()

			startAllPendingFetchers()
			wg.Done()
			return
		}

		c.Logger.Trace().Str("name", string(fetcher.Name())).Msg("Processing fetcher...")

		mutex.Lock()
		fetchersStatus[fetcher.Name()] = FetcherProcessStatusProcessing
		mutex.Unlock()

		fetcherData, fetcherQueries := fetcher.Get(ctx)

		mutex.Lock()
		data[fetcher.Name()] = fetcherData
		queries[fetcher.Name()] = fetcherQueries
		fetchersStatus[fetcher.Name()] = FetcherProcessStatusDone
		mutex.Unlock()

		c.Logger.Trace().
			Str("name", string(fetcher.Name())).
			Msg("Processed fetcher")

		startAllPendingFetchers()
		wg.Done()
	}

	startAllPendingFetchers()
	wg.Wait()
	return data, queries
}
