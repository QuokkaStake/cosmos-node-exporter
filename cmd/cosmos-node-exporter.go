package main

import (
	"main/pkg"
	"main/pkg/config"
	"main/pkg/logger"
	"net/http"

	"github.com/spf13/cobra"
)

var (
	version = "unknown"
)

func Execute(configPath string) {
	appConfig, err := config.GetConfig(configPath)
	if err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not load config")
	}

	if err = appConfig.Validate(); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Provided config is invalid!")
	}

	log := logger.GetLogger(appConfig.LogConfig)
	app := pkg.NewApp(log, appConfig)

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		app.HandleRequest(w, r)
	})

	log.Info().Str("addr", appConfig.ListenAddress).Msg("Listening")
	err = http.ListenAndServe(appConfig.ListenAddress, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not start application")
	}
}

func main() {
	var ConfigPath string

	rootCmd := &cobra.Command{
		Use:     "cosmos-node-exporter",
		Long:    "A Prometheus scraper to return data about fullnode sync status and present upgrades.",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			Execute(ConfigPath)
		},
	}

	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")
	if err := rootCmd.MarkPersistentFlagRequired("config"); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not set flag as required")
	}

	if err := rootCmd.Execute(); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not start application")
	}
}
