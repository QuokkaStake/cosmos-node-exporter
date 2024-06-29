package config

import (
	"errors"

	"gopkg.in/guregu/null.v4"
)

type CosmovisorConfig struct {
	Enabled         null.Bool `default:"true"           toml:"enabled"`
	ChainBinaryName string    `toml:"chain-binary-name"`
	ChainFolder     string    `toml:"chain-folder"`
	CosmovisorPath  string    `toml:"cosmovisor-path"`
}

func (c *CosmovisorConfig) Validate() error {
	if !c.Enabled.Bool {
		return nil
	}

	if c.ChainBinaryName == "" {
		return errors.New("chain-binary-name is not specified")
	}

	if c.ChainFolder == "" {
		return errors.New("chain-folder is not specified")
	}

	if c.CosmovisorPath == "" {
		return errors.New("cosmovisor-path is not specified")
	}

	return nil
}
