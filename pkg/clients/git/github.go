package git

import (
	"context"
	"encoding/json"
	"fmt"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/query_info"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"

	"github.com/rs/zerolog"
)

const API_BASE_URL = "https://api.github.com"

type Github struct {
	ApiBaseUrl     string
	Organization   string
	Repository     string
	Token          string
	Logger         zerolog.Logger
	LastResult     string
	LastResultTime time.Time
	Tracer         trace.Tracer
}

type GithubReleaseInfo struct {
	Name    string `json:"name"`
	TagName string `json:"tag_name"`
	Message string `json:"message"`
}

func NewGithub(config config.GitConfig, logger zerolog.Logger, tracer trace.Tracer) *Github {
	value := constants.GithubRegexp.FindStringSubmatch(config.Repository)

	return &Github{
		ApiBaseUrl:   API_BASE_URL,
		Organization: value[1],
		Repository:   value[2],
		Token:        config.Token,
		Logger:       logger.With().Str("component", "github").Logger(),
		LastResult:   "",
		Tracer:       tracer,
	}
}

func (g *Github) HasCachedResult() bool {
	if g.LastResult == "" {
		return false
	}

	return g.LastResultTime.Add(constants.UncachedGithubQueryTime).Sub(time.Now()) > 0
}

func (g *Github) GetLatestRelease(ctx context.Context) (string, query_info.QueryInfo, error) {
	if g.HasCachedResult() {
		g.Logger.Trace().
			Str("time-since-latest", time.Since(g.LastResultTime).String()).
			Msg("Use Github response from cache")
		return g.LastResult, query_info.QueryInfo{
			Module:  constants.ModuleGit,
			Action:  constants.ActionGitGetLatestRelease,
			Success: true,
		}, nil
	}

	childCtx, span := g.Tracer.Start(ctx, "HTTP request")
	defer span.End()

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	latestReleaseUrl := fmt.Sprintf(
		"%s/repos/%s/%s/releases/latest",
		g.ApiBaseUrl,
		g.Organization,
		g.Repository,
	)

	queryInfo := query_info.QueryInfo{
		Module:  constants.ModuleGit,
		Action:  constants.ActionGitGetLatestRelease,
		Success: false,
	}

	req, err := http.NewRequestWithContext(childCtx, http.MethodGet, latestReleaseUrl, nil)
	if err != nil {
		return "", queryInfo, err
	}

	g.Logger.Trace().
		Str("url", latestReleaseUrl).
		Msg("Querying GitHub")

	if g.Token != "" {
		g.Logger.Trace().Msg("Using personal token for Github requests")
		req.Header.Set("Authorization", "Bearer "+g.Token)
	}

	res, err := client.Do(req)
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}

	if err != nil {
		return "", queryInfo, err
	}

	// rate limiting
	rateLimitTimeHeader := res.Header.Get("x-ratelimit-reset") //nolint:canonicalheader
	rateLimitHeaderInt, err := strconv.ParseInt(rateLimitTimeHeader, 10, 64)
	if err != nil {
		return "", queryInfo, err
	}

	rateLimitTime := time.Unix(rateLimitHeaderInt, 0)

	g.Logger.Trace().
		Str("url", latestReleaseUrl).
		Str("ratelimit-limit", res.Header.Get("x-ratelimit-limit")).         //nolint:canonicalheader
		Str("ratelimit-remaining", res.Header.Get("x-ratelimit-remaining")). //nolint:canonicalheader
		Time("ratelimit-reset", rateLimitTime).
		Msg("GitHub query finished")

	releaseInfo := GithubReleaseInfo{}
	err = json.NewDecoder(res.Body).Decode(&releaseInfo)

	if err != nil {
		return "", queryInfo, err
	}

	// GitHub returned error, probably rate-limiting
	if releaseInfo.Message != "" {
		return "", queryInfo, fmt.Errorf("got error from Github: %s", releaseInfo.Message)
	}

	g.LastResultTime = time.Now()
	g.LastResult = releaseInfo.TagName

	queryInfo.Success = true

	return releaseInfo.TagName, queryInfo, err
}
