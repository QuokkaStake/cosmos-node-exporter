package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

type Cosmovisor struct {
	Logger          zerolog.Logger
	ChainFolder     string
	ChainBinaryName string
}

func NewCosmovisor(
	config *Config,
	logger *zerolog.Logger,
) *Cosmovisor {
	return &Cosmovisor{
		Logger:          logger.With().Str("component", "cosmovisor").Logger(),
		ChainFolder:     config.CosmovisorConfig.ChainFolder,
		ChainBinaryName: config.CosmovisorConfig.ChainBinaryName,
	}
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

	}

	return upgrades, nil
}
