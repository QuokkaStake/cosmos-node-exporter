package fetchers

import (
	"context"
	"main/pkg/constants"
	"main/pkg/query_info"
)

type AppVersionFetcher struct {
	Version string
}

func NewAppVersionFetcher(version string) *AppVersionFetcher {
	return &AppVersionFetcher{
		Version: version,
	}
}

func (u *AppVersionFetcher) Enabled() bool {
	return true
}

func (u *AppVersionFetcher) Name() constants.FetcherName {
	return constants.FetcherNameAppVersion
}

func (u *AppVersionFetcher) Dependencies() []constants.FetcherName {
	return []constants.FetcherName{}
}

func (u *AppVersionFetcher) Get(ctx context.Context, data ...interface{}) (interface{}, []query_info.QueryInfo) {
	return u.Version, []query_info.QueryInfo{}
}
