package git

import (
	"context"
	"encoding/json"
	"errors"
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

type Gitopia struct {
	Organization string
	Repository   string
	Logger       zerolog.Logger
	Tracer       trace.Tracer
}

type GitopiaResponse struct {
	Message string          `json:"message"`
	Release *GitopiaRelease `json:"Release"`
}

type GitopiaRelease struct {
	TagName string `json:"tagName"`
}

func NewGitopia(config config.NodeConfig, logger zerolog.Logger, tracer trace.Tracer) *Gitopia {
	value := constants.GitopiaRegexp.FindStringSubmatch(config.GitConfig.Repository)

	return &Gitopia{
		Organization: value[1],
		Repository:   value[2],
		Logger:       logger.With().Str("component", "gitopia").Logger(),
		Tracer:       tracer,
	}
}

func (g *Gitopia) GetLatestRelease(ctx context.Context) (string, query_info.QueryInfo, error) {
	childCtx, span := g.Tracer.Start(ctx, "HTTP request")
	defer span.End()

	latestReleaseUrl := fmt.Sprintf(
		"https://api.gitopia.com/gitopia/gitopia/gitopia/%s/repository/%s/releases/latest",
		g.Organization,
		g.Repository,
	)

	queryInfo := query_info.QueryInfo{
		Module:  constants.ModuleGit,
		Action:  constants.ActionGitGetLatestRelease,
		Success: false,
	}

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	req, err := http.NewRequestWithContext(childCtx, http.MethodGet, latestReleaseUrl, nil)
	if err != nil {
		return "", queryInfo, err
	}

	g.Logger.Trace().
		Str("url", latestReleaseUrl).
		Msg("Querying Gitopia")

	res, err := client.Do(req)
	if err != nil {
		return "", queryInfo, err
	}
	defer res.Body.Close()

	response := GitopiaResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)

	if err != nil {
		return "", queryInfo, err
	}

	if response.Message != "" {
		return "", queryInfo, errors.New(response.Message)
	}

	if response.Release == nil {
		return "", queryInfo, errors.New("malformed response from Gitopia")
	}

	queryInfo.Success = true

	return response.Release.TagName, queryInfo, err
}
