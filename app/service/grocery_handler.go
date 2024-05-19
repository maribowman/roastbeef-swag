package service

import (
	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/rs/zerolog/log"
)

type GroceryHandler struct {
	channelID            string
	databaseClient       model.DatabaseClient
	lineBreak            int
	shoppingList         []model.PantryItem
	previousShoppingList []model.PantryItem
}

func NewGroceryHandler(channelID string, databaseClient model.DatabaseClient, lineBreak int) model.BotHandler {
	log.Debug().Msg("Registering grocery handler")
	return &GroceryHandler{
		channelID:      channelID,
		databaseClient: databaseClient,
		lineBreak:      lineBreak,
	}
}

func (handler *GroceryHandler) ReadyEvent(session *discordgo.Session, ready *discordgo.Ready) {
	handler.MessageEvent(session, &discordgo.MessageCreate{Message: &discordgo.Message{Author: &discordgo.User{ID: "init"}}})
	items, _, content, _, err := PreProcessMessageEvent(session, handler.channelID, "02.01.")
	if err != nil {
		log.Error().Err(err).Msg("Error while processing message event")
		return
	}
	handler.shoppingList = UpdateItems(items, content)
	log.Debug().Msg("Initialized grocery handler")
}

func (handler *GroceryHandler) MessageEvent(session *discordgo.Session, message *discordgo.MessageCreate) {
	items, lastBotMessageID, content, removableMessageIDs, err := PreProcessMessageEvent(session, handler.channelID, "02.01.")
	if err != nil {
		log.Error().Err(err).Msg("Error while processing message event")
		return
	}

	handler.previousShoppingList = handler.shoppingList
	handler.shoppingList = UpdateItems(items, content)

	if err := session.ChannelMessagesBulkDelete(handler.channelID, removableMessageIDs); err != nil {
		log.Error().Err(err).Msg("Could not bulk delete channel messages")
	}

	PublishItems(handler.shoppingList, session, handler.channelID, lastBotMessageID, handler.lineBreak, "02.01.")
}

func (handler *GroceryHandler) MessageComponentInteractionEvent(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var response *discordgo.InteractionResponse

	switch interaction.MessageComponentData().CustomID {
	case EditButton:
		response = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: EditModal,
				Title:    "Edit grocery list",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID: EditModalInput,
								Style:    discordgo.TextInputParagraph,
								Value:    model.ToList(handler.shoppingList),
							},
						},
					},
				},
			},
		}
	case UndoButton:
		handler.shoppingList = handler.previousShoppingList
		response = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    model.ToMarkdownTable(handler.shoppingList, handler.lineBreak, "02.01."),
				Components: CreateMessageButtons(),
			},
		}
	default:
		log.Error().Msgf("Could not map message component interaction event `%s`", interaction.MessageComponentData().CustomID)
	}

	_ = session.InteractionRespond(interaction.Interaction, response)
}

func (handler *GroceryHandler) ModalSubmitInteractionEvent(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var response *discordgo.InteractionResponse

	switch interaction.ModalSubmitData().CustomID {
	case EditModal:
		handler.previousShoppingList = handler.shoppingList
		handler.shoppingList = UpdateItemsFromList(
			handler.shoppingList,
			interaction.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
		)
		response = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    model.ToMarkdownTable(handler.shoppingList, handler.lineBreak, "02.01."),
				Components: CreateMessageButtons(),
			},
		}
	default:
		log.Error().Msgf("Could not map modal-submit interaction event `%s`", interaction.ModalSubmitData().CustomID)
	}

	_ = session.InteractionRespond(interaction.Interaction, response)
}
