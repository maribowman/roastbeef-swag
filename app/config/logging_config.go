package config

type LoggingConfig struct {
	LogLevel     string `yaml:"logLevel"`
	OutputFormat string `yaml:"outputFormat"`
}
