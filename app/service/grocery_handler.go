package service

import (
	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/rs/zerolog/log"
	"strings"
)

type GroceryHandler struct {
	botID            string
	channelID        string
	shoppingList     []model.PantryItem
	lastShoppingList string
}

func NewGroceryHandler(botID string, channelID string) model.BotHandler {
	log.Debug().Msg("Registering grocery handler")
	return &GroceryHandler{
		botID:     botID,
		channelID: channelID,
	}
}

func (handler *GroceryHandler) ReadyEvent(session *discordgo.Session, ready *discordgo.Ready) {
	handler.MessageEvent(session, &discordgo.MessageCreate{Message: &discordgo.Message{Author: &discordgo.User{ID: "init"}}})
	log.Debug().Msg("Initialized grocery handler")
}

func (handler *GroceryHandler) MessageEvent(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == handler.botID {
		return
	}

	channelMessages, err := session.ChannelMessages(handler.channelID, 100, "", "", "")
	if err != nil {
		return
	}

	var lastBotMessage *discordgo.Message
	var content string
	var removableMessageIDs []string

	for _, msg := range channelMessages {
		if msg.Author.ID == handler.botID {
			if lastBotMessage == nil {
				lastBotMessage = msg
				handler.shoppingList = model.FromMarkdownTable(msg.Content)
				continue
			} else if lastBotMessage.Timestamp.After(msg.Timestamp) {
				removableMessageIDs = append(removableMessageIDs, lastBotMessage.ID)
				lastBotMessage = msg
				handler.shoppingList = model.FromMarkdownTable(msg.Content)
				continue
			}
		} else {
			content += "\n" + msg.Content
		}
		removableMessageIDs = append(removableMessageIDs, msg.ID)
	}

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if removeRegex.MatchString(line) {
			handler.shoppingList = Remove(handler.shoppingList, line)
		} else {
			handler.shoppingList = Add(handler.shoppingList, line)
		}
	}

	if err := session.ChannelMessagesBulkDelete(handler.channelID, removableMessageIDs); err != nil {
		log.Error().Err(err).Msg("Could not bulk delete channel messages")
	}

	PublishItems(handler.shoppingList, session, lastBotMessage.ChannelID, lastBotMessage.ID)
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
				Content:    model.ToMarkdownTable(handler.shoppingList, ""),
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
				Content:    model.ToMarkdownTable(handler.shoppingList, ""),
				Components: CreateMessageButtons(),
			},
		}
	default:
		log.Error().Msgf("Could not map modal-submit interaction event `%s`", interaction.ModalSubmitData().CustomID)
	}

	_ = session.InteractionRespond(interaction.Interaction, response)
}
