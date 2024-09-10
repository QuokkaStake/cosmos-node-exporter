package fetchers

import (
	"context"
	"main/pkg/constants"
	"main/pkg/query_info"
)

type Fetcher interface {
	Enabled() bool
	Get(ctx context.Context, data ...interface{}) (interface{}, []query_info.QueryInfo)
	Dependencies() []constants.FetcherName
	Name() constants.FetcherName
}

type Fetchers []Fetcher

func (f Fetchers) GetNames() []constants.FetcherName {
	names := make([]constants.FetcherName, len(f))

	for index, fetcher := range f {
		names[index] = fetcher.Name()
	}

	return names
}

func (f Fetchers) GetNamesAsString() []string {
	names := make([]string, len(f))

	for index, fetcher := range f {
		names[index] = string(fetcher.Name())
	}

	return names
}
