package service

import (
	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/maribowman/roastbeef-swag/app/repository"
	"github.com/rs/zerolog/log"
)

type PantryHandler struct {
	channelID    string
	pantryClient model.PantryClient
	lineBreak    int
	dateFormat   string
	modalTitle   string
}

func NewPantryHandler(channelID string, databaseClient model.DatabaseClient, lineBreak int, tableName, dateFormat, modalTitle string) model.BotHandler {
	log.Debug().Msgf("Registering pantry handler for table `%s`", tableName)
	return &PantryHandler{
		channelID:    channelID,
		pantryClient: repository.NewSqlitePantryClient(databaseClient, tableName),
		lineBreak:    lineBreak,
		dateFormat:   dateFormat,
		modalTitle:   modalTitle,
	}
}

func (handler *PantryHandler) ReadyEvent(session *discordgo.Session) (err error) {
	handler.MessageEvent(session, &discordgo.MessageCreate{Message: &discordgo.Message{Author: &discordgo.User{ID: "init"}}})
	return
}

func (handler *PantryHandler) MessageEvent(session *discordgo.Session, _ *discordgo.MessageCreate) {
	handler.syncChannel(session)
}

// syncChannel is the single source of truth for cleaning up the channel and
// re-publishing the latest pantry state. It reads the channel, deletes every message
// except the one bot message to be reused, applies any pending user input, and
// republishes — splitting across multiple messages when needed.
func (handler *PantryHandler) syncChannel(session *discordgo.Session) {
	lastBotMessageID, userInput, removableMessageIDs, err := PreProcessMessageEvent(session, handler.channelID)
	if err != nil {
		log.Error().Err(err).Msg("Error while processing message event")
		return
	}

	if err := session.ChannelMessagesBulkDelete(handler.channelID, removableMessageIDs); err != nil {
		log.Error().Err(err).Msg("Could not bulk delete channel messages")
	}

	UpdateItems(handler.pantryClient, userInput)
	PublishItems(handler.pantryClient.GetItems(), session, handler.channelID, lastBotMessageID, handler.lineBreak, handler.dateFormat)
}

func (handler *PantryHandler) MessageComponentInteractionEvent(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	switch interaction.MessageComponentData().CustomID {
	case EditButton:
		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: EditModal,
				Title:    handler.modalTitle,
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID: EditModalInput,
								Style:    discordgo.TextInputParagraph,
								Label:    "Edit",
								Value:    model.ToList(handler.pantryClient.GetItems()),
							},
						},
					},
				},
			},
		}
		if err := session.InteractionRespond(interaction.Interaction, response); err != nil {
			log.Error().Err(err).Msg("Failed to return interaction response")
		}
	case UndoButton:
		// Acknowledge without touching the source message, then republish through the
		// shared flow so multi-message tables stay consistent.
		if err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		}); err != nil {
			log.Error().Err(err).Msg("Failed to acknowledge interaction")
		}
		handler.syncChannel(session)
	default:
		log.Error().Msgf("Could not map message component interaction event `%s`", interaction.MessageComponentData().CustomID)
	}
}

func (handler *PantryHandler) ModalSubmitInteractionEvent(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	switch interaction.ModalSubmitData().CustomID {
	case EditModal:
		UpdateItemsFromModal(
			handler.pantryClient,
			interaction.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
		)
		// Acknowledge the modal, then republish through the shared flow so multi-message
		// tables stay consistent and never exceed the 2000-character message limit.
		if err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		}); err != nil {
			log.Error().Err(err).Msg("Failed to acknowledge interaction")
		}
		handler.syncChannel(session)
	default:
		log.Error().Msgf("Could not map modal-submit interaction event `%s`", interaction.ModalSubmitData().CustomID)
	}
}
