package config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var Config = loadConfig()

type config struct {
	Server  ServerConfig
	Logging LoggingConfig
}

func loadConfig() config {
	configFile := "local"
	configPath := "./configs"
	if len(os.Args[1:]) > 0 {
		if contains([]string{"local", "int", "prod"}, os.Args[1]) {
			configFile = os.Args[1]
		} else if strings.Contains(os.Args[1], "test") {
			configFile = "test"
			pwd, _ := os.Getwd()
			configPath = subStringAfterBefore(pwd, "app") + "/configs"
		}
	}
	viper.SetConfigName(configFile)
	viper.AddConfigPath(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msgf("viper error while trying to read '%s' file, %s", configFile)
	}
	var config config
	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to decode config into struct")
	}
	log.Info().Msgf("starting service in %s mode", configFile)
	return config
}

func contains(values []string, key string) bool {
	for _, iter := range values {
		if iter == key {
			return true
		}
	}
	return false
}

func subStringAfterBefore(input, delimiter string) string {
	pos := strings.Index(input, delimiter)
	if pos == -1 {
		return ""
	}
	return input[0:pos]
}
