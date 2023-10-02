package model

import "github.com/bwmarrin/discordgo"

type DiscordService interface {
	DispatchHandler(*discordgo.Session, *discordgo.MessageCreate)
	CloseSession()
}

type DiscordBot interface {
	MessageEvent(*discordgo.Session, *discordgo.MessageCreate)
}
