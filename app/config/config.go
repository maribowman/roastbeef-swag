package config

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/joho/godotenv"
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

	loadAndReplaceFromDotEnv(&config)

	return config
}

func getConfigPath() string {
	return filepath.Join(getProjectRoot(), "configs")
}

func getProjectRoot() string {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not get current directory")
	}

	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return currentDir
		}
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			log.Fatal().Err(err).Msg("Could not find project root with go.mod file")
		}
		currentDir = parentDir
	}
}

func loadAndReplaceFromDotEnv(config *config) {
	err := godotenv.Load()
	if err != nil {
		return
	}

	log.Info().Msg("Found .env file -> overwriting configs")
	config.Discord.Token = os.Getenv("BOT_TOKEN")
	config.Discord.BotID = os.Getenv("BOT_ID")
}
