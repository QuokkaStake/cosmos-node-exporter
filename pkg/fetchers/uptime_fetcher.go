package fetchers

import (
	"context"
	"main/pkg/constants"
	"main/pkg/query_info"
	"time"
)

type UptimeFetcher struct {
	StartTime time.Time
}

func NewUptimeFetcher() *UptimeFetcher {
	return &UptimeFetcher{
		StartTime: time.Now(),
	}
}

func (u *UptimeFetcher) Enabled() bool {
	return true
}

func (u *UptimeFetcher) Name() constants.FetcherName {
	return constants.FetcherNameUptime
}

func (u *UptimeFetcher) Dependencies() []constants.FetcherName {
	return []constants.FetcherName{}
}

func (u *UptimeFetcher) Get(ctx context.Context) (interface{}, []query_info.QueryInfo) {
	return u.StartTime, []query_info.QueryInfo{}
}
