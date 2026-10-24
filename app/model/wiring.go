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
	RemoveAllItems()
	GetItems() []PantryItem
	// CreateSnapshot overwrites the stored snapshot with the current items.
	CreateSnapshot()
	// RestoreSnapshot replaces all items with the stored snapshot and deletes it.
	// Returns false and changes nothing if no snapshot exists.
	RestoreSnapshot() bool
}
