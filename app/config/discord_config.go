package config

type DiscordConfig struct {
	Token    string    `yaml:"token"`
	BotID    string    `yaml:"botID"`
	Channels []Channel `yaml:"channels"`
}

type Channel struct {
	Name      string `yaml:"name"`
	ID        string `yaml:"id"`
	LineBreak int    `yaml:"lineBreak"`
}
