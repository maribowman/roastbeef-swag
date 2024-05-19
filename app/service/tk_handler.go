package service

import (
	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/rs/zerolog/log"
)

type TkHandler struct {
	channelID         string
	databaseClient    model.DatabaseClient
	lineBreak         int
	inventory         []model.PantryItem
	previousInventory []model.PantryItem // use to undo actions
}

func NewTkHandler(channelID string, databaseClient model.DatabaseClient, lineBreak int) model.BotHandler {
	log.Debug().Msg("Registering tk handler")
	return &TkHandler{
		channelID:      channelID,
		databaseClient: databaseClient,
		lineBreak:      lineBreak,
	}
}

func (handler *TkHandler) ReadyEvent(session *discordgo.Session, ready *discordgo.Ready) {
	handler.MessageEvent(session, &discordgo.MessageCreate{Message: &discordgo.Message{Author: &discordgo.User{ID: "init"}}})
	items, _, content, _, err := PreProcessMessageEvent(session, handler.channelID, "02.01.06")
	if err != nil {
		log.Error().Err(err).Msg("Error while processing message event")
		return
	}
	handler.inventory = UpdateItems(items, content)
	log.Debug().Msg("Initialized tk handler")
}

func (handler *TkHandler) MessageEvent(session *discordgo.Session, message *discordgo.MessageCreate) {
	items, lastBotMessageID, content, removableMessageIDs, err := PreProcessMessageEvent(session, handler.channelID, "02.01.06")
	if err != nil {
		log.Error().Err(err).Msg("Error while processing message event")
		return
	}

	handler.previousInventory = handler.inventory
	handler.inventory = UpdateItems(items, content)

	if err := session.ChannelMessagesBulkDelete(message.ChannelID, removableMessageIDs); err != nil {
		log.Error().Err(err).Msg("Could not bulk delete channel messages")
	}

	PublishItems(handler.inventory, session, handler.channelID, lastBotMessageID, handler.lineBreak, "02.01.06")
}

func (handler *TkHandler) MessageComponentInteractionEvent(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var response *discordgo.InteractionResponse

	switch interaction.MessageComponentData().CustomID {
	case EditButton:
		response = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: EditModal,
				Title:    "Edit inventory list",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID: EditModalInput,
								Style:    discordgo.TextInputParagraph,
								Value:    model.ToList(handler.inventory),
							},
						},
					},
				},
			},
		}
	case UndoButton:
		handler.inventory = handler.previousInventory
		response = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    model.ToMarkdownTable(handler.inventory, handler.lineBreak, "02.01.06"),
				Components: CreateMessageButtons(),
			},
		}
	default:
		log.Error().Msgf("Could not map message component interaction event `%s`", interaction.MessageComponentData().CustomID)
	}

	_ = session.InteractionRespond(interaction.Interaction, response)
}

func (handler *TkHandler) ModalSubmitInteractionEvent(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var response *discordgo.InteractionResponse

	switch interaction.ModalSubmitData().CustomID {
	case EditModal:
		handler.previousInventory = handler.inventory
		handler.inventory = UpdateItemsFromList(
			handler.inventory,
			interaction.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
		)
		response = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    model.ToMarkdownTable(handler.inventory, handler.lineBreak, "02.01.06"),
				Components: CreateMessageButtons(),
			},
		}
	default:
		log.Error().Msgf("Could not map modal-submit interaction event `%s`", interaction.ModalSubmitData().CustomID)
	}

	_ = session.InteractionRespond(interaction.Interaction, response)
}
