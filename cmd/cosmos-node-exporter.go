package main

import (
	"main/pkg"
	"main/pkg/config"
	"main/pkg/fs"
	"main/pkg/logger"

	"github.com/spf13/cobra"
)

var (
	version = "unknown"
)

func ExecuteMain(configPath string) {
	filesystem := &fs.OsFS{}

	app := pkg.NewApp(filesystem, configPath, version)
	app.Start()
}

func ExecuteValidateConfig(configPath string) {
	filesystem := &fs.OsFS{}

	appConfig, err := config.GetConfig(filesystem, configPath)
	if err != nil {
		logger.GetDefaultLogger().Panic().Err(err).Msg("Could not load config!")
	}

	if err = appConfig.Validate(); err != nil {
		logger.GetDefaultLogger().Panic().Err(err).Msg("Provided config is invalid!")
	}

	logger.GetDefaultLogger().Info().Msg("Provided config is valid.")
}

func main() {
	var ConfigPath string

	rootCmd := &cobra.Command{
		Use:     "cosmos-node-exporter --config [config path]",
		Long:    "A Prometheus scraper to return data about fullnode sync status and present upgrades.",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			ExecuteMain(ConfigPath)
		},
	}

	validateConfigCmd := &cobra.Command{
		Use:     "validate-config --config [config path]",
		Long:    "Validate config.",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			ExecuteValidateConfig(ConfigPath)
		},
	}

	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")
	_ = rootCmd.MarkPersistentFlagRequired("config")

	validateConfigCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")
	_ = validateConfigCmd.MarkPersistentFlagRequired("config")

	rootCmd.AddCommand(validateConfigCmd)

	if err := rootCmd.Execute(); err != nil {
		logger.GetDefaultLogger().Panic().Err(err).Msg("Could not start application")
	}
}
