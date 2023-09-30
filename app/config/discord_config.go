package config

type DiscordConfig struct {
	Token    string
	Channels []Channel
}

type Channel struct {
	Name string
	ID   string
}
