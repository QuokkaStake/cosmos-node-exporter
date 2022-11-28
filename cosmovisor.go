package main

import (
	"encoding/json"
	"errors"
	"fmt"
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
	config *Config,
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

func (c *Cosmovisor) GetVersion() (VersionInfo, error) {
	out, err := exec.
		Command(c.CosmovisorPath, "run", "version", "--long", "--output", "json").
		CombinedOutput()
	if err != nil {
		c.Logger.Error().Err(err).Str("output", string(out)).Msg("Could not get app version")
		return VersionInfo{}, err
	}

	jsonOutput := getJsonString(string(out))

	var versionInfo VersionInfo
	if err := json.Unmarshal([]byte(jsonOutput), &versionInfo); err != nil {
		c.Logger.Error().Err(err).Str("output", jsonOutput).Msg("Could not unmarshall app version")
		return versionInfo, err
	}

	return versionInfo, nil
}

func (c *Cosmovisor) GetUpgrades() ([]Upgrade, error) {
	upgradesFolder := fmt.Sprintf("%s/cosmovisor/upgrades", c.ChainFolder)
	upgradesFolderContent, err := os.ReadDir(upgradesFolder)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Could not fetch Cosmovisor upgrades folder content")
		return []Upgrade{}, err
	}

	upgrades := []Upgrade{}

	for _, upgradeFolder := range upgradesFolderContent {
		upgrade := Upgrade{
			Name:          upgradeFolder.Name(),
			BinaryPresent: false,
		}

		if !upgradeFolder.IsDir() {
			upgrades = append(upgrades, upgrade)
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
			upgrade.BinaryPresent = true
		}

		upgrades = append(upgrades, upgrade)
	}

	return upgrades, nil
}

func (c *Cosmovisor) GetUpgradePlan() (UpgradePlan, error) {
	out, err := exec.
		Command(c.CosmovisorPath, "run", "query", "upgrade", "plan", "--output", "json").
		CombinedOutput()
	if err != nil {
		// no upgrade planned is ok and shouldn't produce an error
		if strings.Contains(string(out), "no upgrade scheduled") {
			return UpgradePlan{}, nil
		}

		c.Logger.Error().Err(err).Str("output", string(out)).Msg("Could not get upgrade plan")
		return UpgradePlan{}, err
	}

	jsonOutput := getJsonString(string(out))

	var upgradePlan UpgradePlan
	if err := json.Unmarshal([]byte(jsonOutput), &upgradePlan); err != nil {
		c.Logger.Error().Err(err).Str("output", jsonOutput).Msg("Could not unmarshall upgrade plan")
		return upgradePlan, err
	}

	return upgradePlan, nil
}
