package fetchers

type AppFetcher struct {
	Version string
}

func NewAppFetcher(version string) *AppFetcher {
	return &AppFetcher{
		Version: version,
	}
}
