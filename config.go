package main

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/mcuadros/go-defaults"
)

type LogConfig struct {
	LogLevel   string `toml:"level" default:"info"`
	JSONOutput bool   `toml:"json" default:"false"`
}

type TendermintConfig struct {
	Address string `toml:"address" default:"http://localhost:26657"`
}

type Config struct {
	LogConfig        LogConfig        `toml:"log"`
	TendermintConfig TendermintConfig `toml:"tendermint"`
	ListenAddress    string           `toml:"listen-address" default:":9500"`
}

func (c *Config) Validate() error {
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

	defaults.SetDefaults(&configStruct)
	return &configStruct, nil
}
