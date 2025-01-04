package cosmovisor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/exec"
	"main/pkg/fs"
	"main/pkg/query_info"
	"main/pkg/types"
	"main/pkg/utils"
	"os"
	"strings"

	"go.opentelemetry.io/otel/trace"

	"github.com/rs/zerolog"
)

type Cosmovisor struct {
	Logger               zerolog.Logger
	Config               config.CosmovisorConfig
	Tracer               trace.Tracer
	CommandExecutor      exec.CommandExecutor
	Filesystem           fs.FS
	UpgradeSubfolderPath string
}

func NewCosmovisor(
	config config.CosmovisorConfig,
	logger zerolog.Logger,
	tracer trace.Tracer,
) *Cosmovisor {
	return &Cosmovisor{
		Logger:               logger.With().Str("component", "cosmovisor").Logger(),
		Config:               config,
		Tracer:               tracer,
		CommandExecutor:      &exec.NativeCommandExecutor{},
		Filesystem:           &fs.OsFS{},
		UpgradeSubfolderPath: "/cosmovisor/upgrades",
	}
}

// a helper to get the first string in a multiline string starting with { and ending with }
// it's a workaround for cosmovisor as it adds some extra output, causing
// it to not be valid JSON.
func getJsonString(input string) string {
	split := strings.Split(input, "\n")
	for _, line := range split {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
			return trimmed
		}
	}

	// return the whole line, there's no valid JSON there
	return input
}

func (c *Cosmovisor) GetVersion(ctx context.Context) (types.VersionInfo, query_info.QueryInfo, error) {
	_, span := c.Tracer.Start(
		ctx,
		"Fetching cosmovisor app version",
	)
	defer span.End()

	queryInfo := query_info.QueryInfo{
		Module:  constants.ModuleCosmovisor,
		Action:  constants.ActionCosmovisorGetVersion,
		Success: false,
	}

	env := append(
		os.Environ(),
		"DAEMON_NAME="+c.Config.ChainBinaryName,
		"DAEMON_HOME="+c.Config.ChainFolder,
	)

	out, err := c.CommandExecutor.RunWithEnv(
		c.Config.CosmovisorPath,
		[]string{"run", "version", "--long", "--output", "json"},
		env,
	)
	if err != nil {
		c.Logger.Error().
			Err(err).
			Str("output", utils.DecolorifyString(string(out))).
			Msg("Could not get app version")
		span.RecordError(err)
		return types.VersionInfo{}, queryInfo, err
	}

	jsonOutput := getJsonString(string(out))

	var versionInfo types.VersionInfo
	if err := json.Unmarshal([]byte(jsonOutput), &versionInfo); err != nil {
		c.Logger.Error().
			Err(err).
			Str("output", jsonOutput).
			Msg("Could not unmarshall app version")
		span.RecordError(err)
		return versionInfo, queryInfo, err
	}

	queryInfo.Success = true
	return versionInfo, queryInfo, nil
}

func (c *Cosmovisor) GetCosmovisorVersion(ctx context.Context) (string, query_info.QueryInfo, error) {
	_, span := c.Tracer.Start(
		ctx,
		"Fetching cosmovisor version",
	)
	defer span.End()

	queryInfo := query_info.QueryInfo{
		Module:  constants.ModuleCosmovisor,
		Action:  constants.ActionCosmovisorGetCosmovisorVersion,
		Success: false,
	}

	env := append(
		os.Environ(),
		"DAEMON_NAME="+c.Config.ChainBinaryName,
		"DAEMON_HOME="+c.Config.ChainFolder,
	)

	out, _ := c.CommandExecutor.RunWithEnv(c.Config.CosmovisorPath, []string{"version"}, env)
	outSplit := strings.Split(string(out), "\n")

	cosmovisorVersionPrefix := "cosmovisor version: "

	for _, outString := range outSplit {
		if strings.HasPrefix(outString, cosmovisorVersionPrefix) {
			queryInfo.Success = true
			return outString[len(cosmovisorVersionPrefix):], queryInfo, nil
		}
	}

	return "", queryInfo, errors.New("could not find version in Cosmovisor response")
}

func (c *Cosmovisor) GetUpgrades(ctx context.Context) (*types.UpgradesPresent, query_info.QueryInfo, error) {
	_, span := c.Tracer.Start(
		ctx,
		"Fetching cosmovisor upgrades",
	)
	defer span.End()

	cosmovisorGetUpgradesQueryInfo := query_info.QueryInfo{
		Action:  constants.ActionCosmovisorGetUpgrades,
		Module:  constants.ModuleCosmovisor,
		Success: false,
	}

	upgradesFolder := c.Config.ChainFolder + c.UpgradeSubfolderPath
	upgradesFolderContent, err := c.Filesystem.ReadDir(upgradesFolder)
	if err != nil {
		span.RecordError(err)
		c.Logger.Error().Err(err).Msg("Could not fetch Cosmovisor upgrades folder content")
		return nil, cosmovisorGetUpgradesQueryInfo, err
	}

	upgrades := types.UpgradesPresent{}

	for _, upgradeFolder := range upgradesFolderContent {
		if !upgradeFolder.IsDir() {
			continue
		}

		upgradeBinaryPath := fmt.Sprintf(
			"%s/%s/bin/%s",
			upgradesFolder,
			upgradeFolder.Name(),
			c.Config.ChainBinaryName,
		)

		if _, err := c.Filesystem.Stat(upgradeBinaryPath); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				c.Logger.Warn().Err(err).Msg("Error fetching Cosmovisor upgrade")
				span.RecordError(err)
				return &upgrades, cosmovisorGetUpgradesQueryInfo, err
			}

			upgrades[upgradeFolder.Name()] = false
		} else {
			upgrades[upgradeFolder.Name()] = true
		}
	}

	cosmovisorGetUpgradesQueryInfo.Success = true

	return &upgrades, cosmovisorGetUpgradesQueryInfo, nil
}
