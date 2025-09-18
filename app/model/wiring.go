package model

import (
	"database/sql"

	"github.com/bwmarrin/discordgo"
)

type DiscordBot interface {
	Ready(*discordgo.Session, *discordgo.Ready)
	MessageDispatch(*discordgo.Session, *discordgo.MessageCreate)
	InteractionDispatch(*discordgo.Session, *discordgo.InteractionCreate)
	CloseSession()
}

type BotHandler interface {
	ReadyEvent(*discordgo.Session) error
	MessageEvent(*discordgo.Session, *discordgo.MessageCreate)
	MessageComponentInteractionEvent(*discordgo.Session, *discordgo.InteractionCreate)
	ModalSubmitInteractionEvent(*discordgo.Session, *discordgo.InteractionCreate)
}

type DatabaseClient interface {
	GetDatabaseConnection() *sql.DB
	CloseDatabaseConnection()
}

type PantryClient interface {
	AddItem(PantryItem) (int, error)
	UpdateItem(PantryItem)
	RemoveItem(int)
	GetItems() []PantryItem
}
