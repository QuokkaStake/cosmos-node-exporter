package git

import (
	"context"
	"encoding/json"
	"fmt"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/query_info"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"

	"github.com/rs/zerolog"
)

const API_BASE_URL = "https://api.github.com"

type Github struct {
	ApiBaseUrl   string
	Organization string
	Repository   string
	Token        string
	Logger       zerolog.Logger
	LastModified time.Time
	LastResult   string
	Tracer       trace.Tracer
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
		LastModified: time.Now(),
		LastResult:   "",
		Tracer:       tracer,
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

func (g *Github) GetLatestRelease(ctx context.Context) (string, query_info.QueryInfo, error) {
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
		return "", queryInfo, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotModified && g.LastResult != "" {
		queryInfo.Success = true
		g.Logger.Trace().Msg("Github returned cached response")
		return g.LastResult, queryInfo, nil
	}

	releaseInfo := GithubReleaseInfo{}
	err = json.NewDecoder(res.Body).Decode(&releaseInfo)

	if err != nil {
		return "", queryInfo, err
	}

	// GitHub returned error, probably rate-limiting
	if releaseInfo.Message != "" {
		return "", queryInfo, fmt.Errorf("got error from Github: %s", releaseInfo.Message)
	}

	g.LastModified = time.Now()
	g.LastResult = releaseInfo.TagName

	queryInfo.Success = true

	return releaseInfo.TagName, queryInfo, err
}
