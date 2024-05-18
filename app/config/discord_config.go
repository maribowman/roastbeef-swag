package config

type DiscordConfig struct {
	Token    string
	BotID    string
	Channels []Channel
}

type Channel struct {
	Name      string
	ID        string
	LineBreak int `default:"100"`
}
