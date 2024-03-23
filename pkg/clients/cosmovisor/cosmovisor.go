package cosmovisor

import (
	"encoding/json"
	"errors"
	"fmt"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/query_info"
	"main/pkg/types"
	"main/pkg/utils"
	"os"
	"os/exec"
	"strings"

	"github.com/rs/zerolog"
)

type Cosmovisor struct {
	Logger zerolog.Logger
	Config config.CosmovisorConfig
}

func NewCosmovisor(
	config config.NodeConfig,
	logger zerolog.Logger,
) *Cosmovisor {
	return &Cosmovisor{
		Logger: logger.With().Str("component", "cosmovisor").Logger(),
		Config: config.CosmovisorConfig,
	}
}

// a helper to get the first string in a multiline string starting with { and ending with }
// it's a workaround for cosmovisor as it adds some extra output, causing
// it to not be valid JSON.
func getJsonString(input string) string {
	split := strings.Split(input, "\n")
	for _, line := range split {
		if strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}") {
			return line
		}
	}

	// return the whole line, there's no valid JSON there
	return input
}

func (c *Cosmovisor) GetVersion() (types.VersionInfo, query_info.QueryInfo, error) {
	queryInfo := query_info.QueryInfo{
		Module:  constants.ModuleCosmovisor,
		Action:  constants.ActionCosmovisorGetVersion,
		Success: false,
	}

	cmd := exec.Command(c.Config.CosmovisorPath, "run", "version", "--long", "--output", "json")
	cmd.Env = append(
		os.Environ(),
		"DAEMON_NAME="+c.Config.ChainBinaryName,
		"DAEMON_HOME="+c.Config.ChainFolder,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		c.Logger.Error().
			Err(err).
			Str("output", utils.DecolorifyString(string(out))).
			Msg("Could not get app version")
		return types.VersionInfo{}, queryInfo, err
	}

	jsonOutput := getJsonString(string(out))

	var versionInfo types.VersionInfo
	if err := json.Unmarshal([]byte(jsonOutput), &versionInfo); err != nil {
		c.Logger.Error().
			Err(err).
			Str("output", jsonOutput).
			Msg("Could not unmarshall app version")
		return versionInfo, queryInfo, err
	}

	queryInfo.Success = true
	return versionInfo, queryInfo, nil
}

func (c *Cosmovisor) GetCosmovisorVersion() (string, query_info.QueryInfo, error) {
	queryInfo := query_info.QueryInfo{
		Module:  constants.ModuleCosmovisor,
		Action:  constants.ActionCosmovisorGetCosmovisorVersion,
		Success: false,
	}

	cmd := exec.Command(c.Config.CosmovisorPath, "version")
	cmd.Env = append(
		os.Environ(),
		"DAEMON_NAME="+c.Config.ChainBinaryName,
		"DAEMON_HOME="+c.Config.ChainFolder,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		c.Logger.Error().
			Err(err).
			Str("output", utils.DecolorifyString(string(out))).
			Msg("Could not get Cosmovisor version")
		return "", queryInfo, err
	}

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

func (c *Cosmovisor) GetUpgrades() (types.UpgradesPresent, query_info.QueryInfo, error) {
	cosmovisorGetUpgradesQueryInfo := query_info.QueryInfo{
		Action:  constants.ActionCosmovisorGetUpgrades,
		Module:  constants.ModuleCosmovisor,
		Success: false,
	}

	upgradesFolder := c.Config.ChainFolder + "/cosmovisor/upgrades"
	upgradesFolderContent, err := os.ReadDir(upgradesFolder)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Could not fetch Cosmovisor upgrades folder content")
		return map[string]bool{}, cosmovisorGetUpgradesQueryInfo, err
	}

	upgrades := map[string]bool{}

	for _, upgradeFolder := range upgradesFolderContent {
		if !upgradeFolder.IsDir() {
			upgrades[upgradeFolder.Name()] = false
			continue
		}

		upgradeBinaryPath := fmt.Sprintf(
			"%s/%s/bin/%s",
			upgradesFolder,
			upgradeFolder.Name(),
			c.Config.ChainBinaryName,
		)

		if _, err := os.Stat(upgradeBinaryPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				c.Logger.Warn().Err(err).Msg("Error fetching Cosmovisor upgrade")
			}
		} else {
			upgrades[upgradeFolder.Name()] = true
		}
	}

	cosmovisorGetUpgradesQueryInfo.Success = true

	return upgrades, cosmovisorGetUpgradesQueryInfo, nil
}
