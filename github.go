package main

import (
	"encoding/json"
	"fmt"
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
	LastResult   *ReleaseInfo
}

func NewGithub(config *Config, logger *zerolog.Logger) *Github {
	value := GithubRegexp.FindStringSubmatch(config.GithubConfig.Repository)

	return &Github{
		Organization: value[1],
		Repository:   value[2],
		Token:        config.GithubConfig.Token,
		Logger:       logger.With().Str("component", "github").Logger(),
		LastModified: time.Now(),
	}
}

func (g *Github) GetLatestRelease() (ReleaseInfo, error) {
	latestReleaseUrl := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/releases/latest",
		g.Organization,
		g.Repository,
	)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", latestReleaseUrl, nil)
	if err != nil {
		return ReleaseInfo{}, err
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
		return ReleaseInfo{}, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotModified && g.LastResult != nil {
		g.Logger.Trace().Msg("Github returned cached response")
		g.LastModified = time.Now()
		return *g.LastResult, nil
	}

	releaseInfo := ReleaseInfo{}
	err = json.NewDecoder(res.Body).Decode(&releaseInfo)

	g.LastModified = time.Now()
	g.LastResult = &releaseInfo

	return releaseInfo, err
}
