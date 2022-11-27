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
}

func NewGithub(config *Config, logger *zerolog.Logger) *Github {
	value := GithubRegexp.FindStringSubmatch(config.GithubConfig.Repository)

	return &Github{
		Organization: value[1],
		Repository:   value[2],
		Token:        config.GithubConfig.Token,
		Logger:       logger.With().Str("component", "github").Logger(),
	}
}

func (g *Github) GetLatestRelease() (ReleaseInfo, error) {
	latestReleaseUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", g.Organization, g.Repository)

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", latestReleaseUrl, nil)
	if err != nil {
		return ReleaseInfo{}, err
	}

	if g.Token != "" {
		req.Header.Set("Authorization", "Bearer "+g.Token)
	}

	res, err := client.Do(req)
	if err != nil {
		return ReleaseInfo{}, err
	}
	defer res.Body.Close()

	releaseInfo := ReleaseInfo{}
	err = json.NewDecoder(res.Body).Decode(&releaseInfo)

	return releaseInfo, err
}
