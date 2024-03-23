package config

import (
	"errors"
	"fmt"
	"main/pkg/constants"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/creasty/defaults"
	"gopkg.in/guregu/null.v4"
)

type LogConfig struct {
	LogLevel   string    `default:"info"  toml:"level"`
	JSONOutput null.Bool `default:"false" toml:"json"`
}

type TendermintConfig struct {
	Enabled       null.Bool `default:"true"                   toml:"enabled"`
	Address       string    `default:"http://localhost:26657" toml:"address"`
	QueryUpgrades null.Bool `default:"true"                   toml:"query-upgrades"`
}

type GitConfig struct {
	Repository string `default:""   toml:"repository"`
	Token      string `toml:"token"`
}

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

type NodeConfig struct {
	Name             string           `toml:"name"`
	TendermintConfig TendermintConfig `toml:"tendermint"`
	CosmovisorConfig CosmovisorConfig `toml:"cosmovisor"`
	GitConfig        GitConfig        `toml:"git"`
}

type Config struct {
	LogConfig     LogConfig    `toml:"log"`
	NodeConfigs   []NodeConfig `toml:"node"`
	ListenAddress string       `default:":9500" toml:"listen-address"`
}

func (c *GitConfig) Validate() error {
	if c.Repository == "" {
		return nil
	}

	if !constants.GithubRegexp.Match([]byte(c.Repository)) && !constants.GitopiaRegexp.Match([]byte(c.Repository)) {
		return errors.New("repository is not valid")
	}

	return nil
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

func (c *Config) Validate() error {
	if len(c.NodeConfigs) == 0 {
		return errors.New("0 nodes provided")
	}

	for index, nodeConfig := range c.NodeConfigs {
		if err := nodeConfig.Validate(); err != nil {
			return fmt.Errorf("invalid config for node %d: %s", index, err)
		}
	}

	return nil
}

func GetConfig(path string) (*Config, error) {
	configBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	configString := string(configBytes)

	configStruct := Config{}
	if _, err = toml.Decode(configString, &configStruct); err != nil {
		return nil, err
	}

	if err := defaults.Set(&configStruct); err != nil {
		return nil, err
	}
	return &configStruct, nil
}
