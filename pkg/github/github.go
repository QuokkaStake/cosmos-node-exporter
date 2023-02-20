package github

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
	LastResult   *types.ReleaseInfo
}

func NewGithub(config *config.Config, logger *zerolog.Logger) *Github {
	value := constants.GithubRegexp.FindStringSubmatch(config.GithubConfig.Repository)

	return &Github{
		Organization: value[1],
		Repository:   value[2],
		Token:        config.GithubConfig.Token,
		Logger:       logger.With().Str("component", "github").Logger(),
		LastModified: time.Now(),
	}
}

func (g *Github) GetLatestRelease() (types.ReleaseInfo, error) {
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
		return types.ReleaseInfo{}, err
	}

	if g.LastResult != nil {
		req.Header.Set("If-Modified-Since", g.LastModified.Format(http.TimeFormat))
	}

	if g.Token != "" {
		g.Logger.Trace().Msg("Using personal token for Github requests")
		req.Header.Set("Authorization", "Bearer "+g.Token)
	}

	res, err := client.Do(req)
	if err != nil {
		return types.ReleaseInfo{}, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotModified && g.LastResult != nil {
		g.Logger.Trace().Msg("Github returned cached response")
		g.LastModified = time.Now()
		return *g.LastResult, nil
	}

	releaseInfo := types.ReleaseInfo{}
	err = json.NewDecoder(res.Body).Decode(&releaseInfo)

	if err != nil {
		return releaseInfo, err
	}

	// Github returned error, probably rate-limiting
	if releaseInfo.Message != "" {
		return releaseInfo, fmt.Errorf("got error from Github: %s", releaseInfo.Message)
	}

	g.LastModified = time.Now()
	g.LastResult = &releaseInfo

	return releaseInfo, err
}
