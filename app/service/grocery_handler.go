package service

import (
	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/rs/zerolog/log"
)

type GroceryHandler struct {
	channelID        string
	lineBreak        int
	shoppingList     []model.PantryItem
	lastShoppingList string
}

func NewGroceryHandler(channelID string, lineBreak int) model.BotHandler {
	log.Debug().Msg("Registering grocery handler")
	return &GroceryHandler{
		channelID: channelID,
		lineBreak: lineBreak,
	}
}

func (handler *GroceryHandler) ReadyEvent(session *discordgo.Session, ready *discordgo.Ready) {
	handler.MessageEvent(session, &discordgo.MessageCreate{Message: &discordgo.Message{Author: &discordgo.User{ID: "init"}}})
	log.Debug().Msg("Initialized grocery handler")
}

func (handler *GroceryHandler) MessageEvent(session *discordgo.Session, message *discordgo.MessageCreate) {
	items, lastBotMessageID, content, removableMessageIDs, err := PreProcessMessageEvent(session, message)
	if err != nil {
		return
	}

	handler.shoppingList = UpdateHandlerItems(items, content)

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
	case DoneButton:
		handler.shoppingList = []model.PantryItem{}
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
		handler.shoppingList = model.UpdateFromList(
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
