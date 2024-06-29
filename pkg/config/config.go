package config

import (
	"errors"
	"fmt"
	"main/pkg/fs"

	"gopkg.in/guregu/null.v4"

	"github.com/BurntSushi/toml"
	"github.com/creasty/defaults"
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

type GrpcConfig struct {
	Enabled null.Bool `default:"true"           toml:"enabled"`
	Address string    `default:"localhost:9090" toml:"address"`
}

type Config struct {
	LogConfig     LogConfig     `toml:"log"`
	TracingConfig TracingConfig `toml:"tracing"`
	NodeConfigs   []NodeConfig  `toml:"node"`
	ListenAddress string        `default:":9500" toml:"listen-address"`
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

func GetConfig(filesystem fs.FS, path string) (*Config, error) {
	configBytes, err := filesystem.ReadFile(path)
	if err != nil {
		return nil, err
	}

	configString := string(configBytes)

	configStruct := Config{}
	if _, err = toml.Decode(configString, &configStruct); err != nil {
		return nil, err
	}

	defaults.MustSet(&configStruct)
	return &configStruct, nil
}
