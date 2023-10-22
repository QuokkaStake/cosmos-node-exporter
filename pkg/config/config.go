package config

import (
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

type GithubConfig struct {
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
		return fmt.Errorf("chain-binary-name is not specified")
	}

	if c.ChainFolder == "" {
		return fmt.Errorf("chain-folder is not specified")
	}

	if c.CosmovisorPath == "" {
		return fmt.Errorf("cosmovisor-path is not specified")
	}

	return nil
}

type Config struct {
	LogConfig        LogConfig        `toml:"log"`
	TendermintConfig TendermintConfig `toml:"tendermint"`
	CosmovisorConfig CosmovisorConfig `toml:"cosmovisor"`
	GithubConfig     GithubConfig     `toml:"github"`
	ListenAddress    string           `default:":9500"   toml:"listen-address"`
}

func (c *GithubConfig) Validate() error {
	if c.Repository == "" {
		return nil
	}

	if !constants.GithubRegexp.Match([]byte(c.Repository)) {
		return fmt.Errorf("repository is not valid")
	}

	return nil
}

func (c *Config) Validate() error {
	if err := c.GithubConfig.Validate(); err != nil {
		return fmt.Errorf("GitHub config is invalid: %s", err)
	}

	if err := c.CosmovisorConfig.Validate(); err != nil {
		return fmt.Errorf("Cosmovisor config is invalid: %s", err)
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
