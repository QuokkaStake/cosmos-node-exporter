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
	LogLevel   string    `toml:"level" default:"info"`
	JSONOutput null.Bool `toml:"json" default:"false"`
}

type TendermintConfig struct {
	Enabled null.Bool `toml:"enabled" default:"true"`
	Address string    `toml:"address" default:"http://localhost:26657"`
}

type GrpcConfig struct {
	Enabled null.Bool `toml:"enabled" default:"true"`
	Address string    `toml:"address" default:"localhost:9090"`
}

type GithubConfig struct {
	Repository string `toml:"repository" default:""`
	Token      string `toml:"token"`
}

type CosmovisorConfig struct {
	Enabled         null.Bool `toml:"enabled" default:"true"`
	ChainBinaryName string    `toml:"chain-binary-name"`
	ChainFolder     string    `toml:"chain-folder"`
	CosmovisorPath  string    `toml:"cosmovisor-path"`
}

type Config struct {
	LogConfig        LogConfig        `toml:"log"`
	TendermintConfig TendermintConfig `toml:"tendermint"`
	GrpcConfig       GrpcConfig       `toml:"grpc"`
	CosmovisorConfig CosmovisorConfig `toml:"cosmovisor"`
	GithubConfig     GithubConfig     `toml:"github"`
	ListenAddress    string           `toml:"listen-address" default:":9500"`
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
