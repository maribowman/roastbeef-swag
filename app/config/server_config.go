package config

type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}
