package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var Config = loadConfig()

type config struct {
	Server   ServerConfig   `yaml:"server"`
	Logging  LoggingConfig  `yaml:"logging"`
	Discord  DiscordConfig  `yaml:"discord"`
	Database DatabaseConfig `yaml:"database"`
}

func loadConfig() config {
	configFile := "local"
	if len(os.Args[1:]) > 0 {
		if slices.Contains([]string{"local", "prod"}, os.Args[1]) {
			configFile = os.Args[1]
		} else if strings.Contains(os.Args[1], "test") {
			configFile = "test"
		}
	}

	var config config
	configFilePath := filepath.Join(getConfigPath(), fmt.Sprintf("%s.%s", configFile, "yaml"))
	yamlConfig, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Fatal().Err(err).Msgf("Unable to read config from %s", configFilePath)
	}
	err = yaml.Unmarshal(yamlConfig, &config)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to unmarshal config yaml into Config struct")
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
			log.Info().Msg("Could not find project root with go.mod file, returning `/`")
			return parentDir
		}
		currentDir = parentDir
	}
}

func loadAndReplaceFromDotEnv(config *config) {
	if err := godotenv.Load(); err != nil {
		log.Info().Msg("No .env file found, skipping overwrite")
		return
	}

	log.Info().Msg("Found .env file -> overwriting configs")
	config.Discord.Token = os.Getenv("BOT_TOKEN")
	config.Discord.BotID = os.Getenv("BOT_ID")
}
