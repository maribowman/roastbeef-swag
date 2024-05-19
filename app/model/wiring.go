package model

import "github.com/bwmarrin/discordgo"

type DiscordBot interface {
	Ready(*discordgo.Session, *discordgo.Ready)
	MessageDispatch(*discordgo.Session, *discordgo.MessageCreate)
	InteractionDispatch(*discordgo.Session, *discordgo.InteractionCreate)
	CloseSession()
}

type BotHandler interface {
	ReadyEvent(*discordgo.Session, *discordgo.Ready)
	MessageEvent(*discordgo.Session, *discordgo.MessageCreate)
	MessageComponentInteractionEvent(*discordgo.Session, *discordgo.InteractionCreate)
	ModalSubmitInteractionEvent(*discordgo.Session, *discordgo.InteractionCreate)
}

type DatabaseClient interface {
	CloseDatabaseConnections()
}
