package model

import "github.com/bwmarrin/discordgo"

type DiscordService interface {
	ReadyHandler(*discordgo.Session, *discordgo.Ready)
	MessageDispatchHandler(*discordgo.Session, *discordgo.MessageCreate)
	InteractionDispatchHandler(*discordgo.Session, *discordgo.InteractionCreate)
	CloseSession()
}

type DiscordBot interface {
	ReadyEvent(*discordgo.Session, *discordgo.Ready)
	MessageEvent(*discordgo.Session, *discordgo.MessageCreate)
	MessageComponentInteractionEvent(*discordgo.Session, *discordgo.InteractionCreate)
	ModalSubmitInteractionEvent(*discordgo.Session, *discordgo.InteractionCreate)
}
