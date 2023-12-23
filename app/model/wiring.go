package model

import "github.com/bwmarrin/discordgo"

type DiscordService interface {
	MessageDispatchHandler(*discordgo.Session, *discordgo.MessageCreate)
	InteractionDispatchHandler(*discordgo.Session, *discordgo.InteractionCreate)
	CloseSession()
}

type DiscordBot interface {
	MessageEvent(*discordgo.Session, *discordgo.MessageCreate)
	InteractionEvent(*discordgo.Session, *discordgo.InteractionCreate)
}
