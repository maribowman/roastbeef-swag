package service

import (
	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/maribowman/roastbeef-swag/app/repository"
	"github.com/rs/zerolog/log"
)

type FreezerHandler struct {
	channelID    string
	pantryClient model.PantryClient
	lineBreak    int
	dateFormat   string
}

func NewFreezerHandler(channelID string, databaseClient model.DatabaseClient, lineBreak int) model.BotHandler {
	log.Debug().Msg("Registering TK handler")
	return &FreezerHandler{
		channelID:    channelID,
		pantryClient: repository.NewSqlitePantryClient(databaseClient, "freezer"),
		lineBreak:    lineBreak,
		dateFormat:   "02.01.06",
	}
}

func (handler *FreezerHandler) ReadyEvent(session *discordgo.Session) (err error) {
	handler.MessageEvent(session, &discordgo.MessageCreate{Message: &discordgo.Message{Author: &discordgo.User{ID: "init"}}})
	_, userInput, _, err := PreProcessMessageEvent(session, handler.channelID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to pre-process message events for TK handler")
		return
	}

	UpdateItems(handler.pantryClient, userInput)
	return
}

func (handler *FreezerHandler) MessageEvent(session *discordgo.Session, message *discordgo.MessageCreate) {
	lastBotMessageID, userInput, removableMessageIDs, err := PreProcessMessageEvent(session, handler.channelID)
	if err != nil {
		log.Error().Err(err).Msg("Error while processing message event")
		return
	}

	if err := session.ChannelMessagesBulkDelete(message.ChannelID, removableMessageIDs); err != nil {
		log.Error().Err(err).Msg("Could not bulk delete channel messages")
	}

	UpdateItems(handler.pantryClient, userInput)
	PublishItems(handler.pantryClient.GetItems(), session, handler.channelID, lastBotMessageID, handler.lineBreak, handler.dateFormat)
}

func (handler *FreezerHandler) MessageComponentInteractionEvent(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var response *discordgo.InteractionResponse

	switch interaction.MessageComponentData().CustomID {
	case EditButton:
		response = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: EditModal,
				Title:    "Edit TK list",
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
	case UndoButton:
		handler.UseLatestSnapshot() // TODO: Implement snapshots
		response = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    model.ToMarkdownTable(handler.pantryClient.GetItems(), handler.lineBreak, handler.dateFormat),
				Components: CreateMessageButtons(),
			},
		}
	default:
		log.Error().Msgf("Could not map message component interaction event `%s`", interaction.MessageComponentData().CustomID)
	}

	_ = session.InteractionRespond(interaction.Interaction, response)
}

func (handler *FreezerHandler) ModalSubmitInteractionEvent(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var response *discordgo.InteractionResponse

	switch interaction.ModalSubmitData().CustomID {
	case EditModal:
		UpdateItemsFromModal(
			handler.pantryClient,
			interaction.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
		)
		response = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    model.ToMarkdownTable(handler.pantryClient.GetItems(), handler.lineBreak, handler.dateFormat),
				Components: CreateMessageButtons(),
			},
		}
	default:
		log.Error().Msgf("Could not map modal-submit interaction event `%s`", interaction.ModalSubmitData().CustomID)
	}

	_ = session.InteractionRespond(interaction.Interaction, response)
}
