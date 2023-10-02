package config

type DiscordConfig struct {
	Token    string
	BotID    string
	Channels map[string]string
}
