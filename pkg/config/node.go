package config

import (
	"errors"
	"fmt"
)

type NodeConfig struct {
	Name             string           `toml:"name"`
	TendermintConfig TendermintConfig `toml:"tendermint"`
	CosmovisorConfig CosmovisorConfig `toml:"cosmovisor"`
	GrpcConfig       GrpcConfig       `toml:"grpc"`
	GitConfig        GitConfig        `toml:"git"`
}

func (c *NodeConfig) Validate() error {
	if c.Name == "" {
		return errors.New("node name is empty")
	}

	if err := c.GitConfig.Validate(); err != nil {
		return fmt.Errorf("GitHub config is invalid: %s", err)
	}

	if err := c.CosmovisorConfig.Validate(); err != nil {
		return fmt.Errorf("Cosmovisor config is invalid: %s", err)
	}

	return nil
}
