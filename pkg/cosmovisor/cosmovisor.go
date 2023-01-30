package cosmovisor

import (
	"encoding/json"
	"errors"
	"fmt"
	"main/pkg/config"
	"main/pkg/types"
	"os"
	"os/exec"
	"strings"

	"github.com/rs/zerolog"
)

type Cosmovisor struct {
	Logger          zerolog.Logger
	ChainFolder     string
	ChainBinaryName string
	CosmovisorPath  string
}

func NewCosmovisor(
	config *config.Config,
	logger *zerolog.Logger,
) *Cosmovisor {
	return &Cosmovisor{
		Logger:          logger.With().Str("component", "cosmovisor").Logger(),
		ChainFolder:     config.CosmovisorConfig.ChainFolder,
		ChainBinaryName: config.CosmovisorConfig.ChainBinaryName,
		CosmovisorPath:  config.CosmovisorConfig.CosmovisorPath,
	}
}

// a helper to get the first string in a multiline string starting with { and ending with }
// it's a workaround for cosmovisor as it adds some extra output, causing
// it to not be valid JSON
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

func (c *Cosmovisor) GetVersion() (types.VersionInfo, error) {
	out, err := exec.
		Command(c.CosmovisorPath, "run", "version", "--long", "--output", "json").
		CombinedOutput()
	if err != nil {
		c.Logger.Error().Err(err).Str("output", string(out)).Msg("Could not get app version")
		return types.VersionInfo{}, err
	}

	jsonOutput := getJsonString(string(out))

	var versionInfo types.VersionInfo
	if err := json.Unmarshal([]byte(jsonOutput), &versionInfo); err != nil {
		c.Logger.Error().Err(err).Str("output", jsonOutput).Msg("Could not unmarshall app version")
		return versionInfo, err
	}

	return versionInfo, nil
}

func (c *Cosmovisor) GetUpgrades() (types.UpgradesPresent, error) {
	upgradesFolder := fmt.Sprintf("%s/cosmovisor/upgrades", c.ChainFolder)
	upgradesFolderContent, err := os.ReadDir(upgradesFolder)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Could not fetch Cosmovisor upgrades folder content")
		return map[string]bool{}, err
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
			c.ChainBinaryName,
		)

		if _, err := os.Stat(upgradeBinaryPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				c.Logger.Warn().Err(err).Msg("Error fetching Cosmovisor upgrade")
			}
		} else {
			upgrades[upgradeFolder.Name()] = true
		}

	}

	return upgrades, nil
}
