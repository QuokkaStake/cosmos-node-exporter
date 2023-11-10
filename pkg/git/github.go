package git

import (
	"encoding/json"
	"fmt"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/types"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type Github struct {
	Organization string
	Repository   string
	Token        string
	Logger       zerolog.Logger
	LastModified time.Time
	LastResult   string
}

func NewGithub(config *config.Config, logger *zerolog.Logger) *Github {
	value := constants.GithubRegexp.FindStringSubmatch(config.GitConfig.Repository)

	return &Github{
		Organization: value[1],
		Repository:   value[2],
		Token:        config.GitConfig.Token,
		Logger:       logger.With().Str("component", "git").Logger(),
		LastModified: time.Now(),
		LastResult:   "",
	}
}

func (g *Github) UseCache() bool {
	// If the last result is not present - do not use cache, for the first query.
	if g.LastResult == "" {
		return false
	}

	// We need to make uncached requests once in a while, to make sure everything is ok
	// (for example, if we messed up caching itself).
	diff := time.Since(g.LastModified)
	return diff < constants.UncachedGithubQueryTime
}

func (g *Github) GetLatestRelease() (string, error) {
	latestReleaseUrl := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/releases/latest",
		g.Organization,
		g.Repository,
	)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, latestReleaseUrl, nil)
	if err != nil {
		return "", err
	}

	useCache := g.UseCache()

	g.Logger.Trace().
		Str("url", latestReleaseUrl).
		Bool("cached", useCache).
		Str("time-since-latest", time.Since(g.LastModified).String()).
		Msg("Querying GitHub")

	if useCache {
		req.Header.Set("If-Modified-Since", g.LastModified.Format(http.TimeFormat))
	}

	if g.Token != "" {
		g.Logger.Trace().Msg("Using personal token for Github requests")
		req.Header.Set("Authorization", "Bearer "+g.Token)
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotModified && g.LastResult != "" {
		g.Logger.Trace().Msg("Github returned cached response")
		return g.LastResult, nil
	}

	releaseInfo := types.ReleaseInfo{}
	err = json.NewDecoder(res.Body).Decode(&releaseInfo)

	if err != nil {
		return "", err
	}

	// GitHub returned error, probably rate-limiting
	if releaseInfo.Message != "" {
		return "", fmt.Errorf("got error from Github: %s", releaseInfo.Message)
	}

	g.LastModified = time.Now()
	g.LastResult = releaseInfo.TagName

	return releaseInfo.TagName, err
}
