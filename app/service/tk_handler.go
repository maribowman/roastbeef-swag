package service

import (
	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/rs/zerolog/log"
)

type TkHandler struct {
	channelID     string
	inventory     []model.PantryItem
	lastInventory string
}

func NewTkHandler(channelID string) model.BotHandler {
	log.Debug().Msg("Registering tk handler")
	return &TkHandler{
		channelID: channelID,
	}
}

func (handler *TkHandler) ReadyEvent(session *discordgo.Session, ready *discordgo.Ready) {
	handler.MessageEvent(session, &discordgo.MessageCreate{Message: &discordgo.Message{Author: &discordgo.User{ID: "init"}}})
	log.Debug().Msg("Initialized tk handler")
}

func (handler *TkHandler) MessageEvent(session *discordgo.Session, message *discordgo.MessageCreate) {
	items, lastBotMessageID, content, removableMessageIDs, err := PreProcessMessageEvent(session, message)
	if err != nil {
		return
	}

	handler.inventory = UpdateHandlerItems(items, content)

	if err := session.ChannelMessagesBulkDelete(message.ChannelID, removableMessageIDs); err != nil {
		log.Error().Err(err).Msg("Could not bulk delete channel messages")
	}

	PublishItems(handler.inventory, session, handler.channelID, lastBotMessageID)
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
	case DoneButton:
		handler.inventory = []model.PantryItem{}
		response = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    model.ToMarkdownTable(handler.inventory, "02.01.06"),
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
		handler.inventory = model.UpdateFromList(
			handler.inventory,
			interaction.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
		)
		response = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    model.ToMarkdownTable(handler.inventory, ""),
				Components: CreateMessageButtons(),
			},
		}
	default:
		log.Error().Msgf("Could not map modal-submit interaction event `%s`", interaction.ModalSubmitData().CustomID)
	}

	_ = session.InteractionRespond(interaction.Interaction, response)
}
