package config

import (
	"errors"
	"fmt"
	"main/pkg/constants"
	"os"

	"gopkg.in/guregu/null.v4"

	"github.com/BurntSushi/toml"
	"github.com/creasty/defaults"
)

type TracingConfig struct {
	Enabled                   null.Bool `default:"false"                     toml:"enabled"`
	OpenTelemetryHTTPHost     string    `toml:"open-telemetry-http-host"`
	OpenTelemetryHTTPInsecure null.Bool `default:"true"                      toml:"open-telemetry-http-insecure"`
	OpenTelemetryHTTPUser     string    `toml:"open-telemetry-http-user"`
	OpenTelemetryHTTPPassword string    `toml:"open-telemetry-http-password"`
}

func (c *TracingConfig) Validate() error {
	if c.Enabled.Bool && c.OpenTelemetryHTTPHost == "" {
		return errors.New("tracing is enabled, but open-telemetry-http-host is not provided")
	}

	return nil
}

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

type GrpcConfig struct {
	Enabled null.Bool `default:"true"           toml:"enabled"`
	Address string    `default:"localhost:9090" toml:"address"`
}

type NodeConfig struct {
	Name             string           `toml:"name"`
	TendermintConfig TendermintConfig `toml:"tendermint"`
	CosmovisorConfig CosmovisorConfig `toml:"cosmovisor"`
	GrpcConfig       GrpcConfig       `toml:"grpc"`
	GitConfig        GitConfig        `toml:"git"`
}

type Config struct {
	LogConfig     LogConfig     `toml:"log"`
	TracingConfig TracingConfig `toml:"tracing"`
	NodeConfigs   []NodeConfig  `toml:"node"`
	ListenAddress string        `default:":9500" toml:"listen-address"`
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

	if err := c.TracingConfig.Validate(); err != nil {
		return fmt.Errorf("tracing config is invalid: %s", err)
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
