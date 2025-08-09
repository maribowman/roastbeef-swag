package config

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var Config = loadConfig()

type config struct {
	Server   ServerConfig
	Logging  LoggingConfig
	Discord  DiscordConfig
	Database DatabaseConfig
}

func loadConfig() config {
	configFile := "local"
	if len(os.Args[1:]) > 0 {
		if slices.Contains([]string{"local", "int", "prod"}, os.Args[1]) {
			configFile = os.Args[1]
		} else if strings.Contains(os.Args[1], "test") { // TODO: Check input args for unit tests
			configFile = "test"
		}
	}

	// TODO: Replace Viper with plain yaml loader
	viper.SetConfigName(configFile)
	viper.AddConfigPath(getConfigPath())
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msgf("Viper error while trying to read config `%s`", configFile)
	}

	var config config
	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to decode config into struct")
	}

	return config
}

func getConfigPath() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not get current directory")
	}
	return filepath.Join(wd, "configs")
}
