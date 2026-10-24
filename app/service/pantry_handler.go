package service

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/maribowman/roastbeef-swag/app/repository"
	"github.com/rs/zerolog/log"
)

type PantryHandler struct {
	// mu serializes syncs on this channel. discordgo runs every event handler in its
	// own goroutine, so without it two syncs could read the same user messages and
	// apply them twice.
	mu           sync.Mutex
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

func (handler *PantryHandler) ReadyEvent(session *discordgo.Session) error {
	return handler.syncChannel(session, nil)
}

func (handler *PantryHandler) MessageEvent(session *discordgo.Session, _ *discordgo.MessageCreate) {
	if err := handler.syncChannel(session, nil); err != nil {
		log.Error().Err(err).Msgf("Could not sync channel `%s` on message event", handler.channelID)
	}
}

// syncChannel is the single source of truth for cleaning up the channel and
// re-publishing the latest pantry state. Under the per-channel lock it reads the
// channel, deletes every message except the one bot message to be reused, runs the
// optional action, applies any pending user input, and republishes — splitting
// across multiple messages when needed. If reading or deleting fails, it returns
// before mutating anything.
func (handler *PantryHandler) syncChannel(session *discordgo.Session, action func()) error {
	handler.mu.Lock()
	defer handler.mu.Unlock()

	lastBotMessageID, userInput, removableMessageIDs, err := PreProcessMessageEvent(session, handler.channelID)
	if err != nil {
		return fmt.Errorf("could not read channel messages: %w", err)
	}

	if err := deleteMessages(session, handler.channelID, removableMessageIDs); err != nil {
		if strings.TrimSpace(userInput) != "" {
			log.Warn().Msgf("Discarding user input on channel `%s` after failed delete: %q", handler.channelID, userInput)
		}
		return fmt.Errorf("could not delete channel messages: %w", err)
	}

	if action != nil {
		action()
	}

	if strings.TrimSpace(userInput) != "" {
		ApplyUserInput(handler.pantryClient, userInput)
	}

	PublishItems(handler.pantryClient.GetItems(), session, handler.channelID, lastBotMessageID, handler.lineBreak, handler.dateFormat)
	return nil
}

// deleteMessages removes the given messages from the channel. It tries a bulk delete
// first and, if that fails (e.g. rate limit, or messages older than 14 days, which
// bulk delete rejects), falls back to deleting each message on its own.
//
// If any individual delete still fails, an error is returned and the caller aborts
// the sync without applying the input. Trade-off: input from messages that were
// already deleted in the fallback is lost (and logged by the caller) rather than
// being re-applied on every later sync because the messages stay in the channel.
func deleteMessages(session *discordgo.Session, channelID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	bulkErr := session.ChannelMessagesBulkDelete(channelID, ids)
	if bulkErr == nil {
		return nil
	}
	log.Error().Err(bulkErr).Msgf("Could not bulk delete %d messages on channel `%s`; deleting one by one", len(ids), channelID)

	var errs []error
	for _, id := range ids {
		if err := session.ChannelMessageDelete(channelID, id); err != nil {
			log.Error().Err(err).Msgf("Could not delete message `%s` on channel `%s`", id, channelID)
			errs = append(errs, fmt.Errorf("message %s: %w", id, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to delete %d of %d messages: %w", len(errs), len(ids), errors.Join(errs...))
	}
	return nil
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
		// Undo never takes a snapshot, so a second click is a no-op instead of a redo.
		if err := handler.syncChannel(session, func() {
			if !handler.pantryClient.RestoreSnapshot() {
				log.Debug().Msg("Nothing to undo")
			}
		}); err != nil {
			log.Error().Err(err).Msgf("Could not sync channel `%s` on undo", handler.channelID)
		}
	default:
		log.Error().Msgf("Could not map message component interaction event `%s`", interaction.MessageComponentData().CustomID)
	}
}

func (handler *PantryHandler) ModalSubmitInteractionEvent(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	switch interaction.ModalSubmitData().CustomID {
	case EditModal:
		modalInput := interaction.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		// Acknowledge first: Discord requires a response within 3 seconds, and the sync
		// below may wait on the channel lock.
		if err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		}); err != nil {
			log.Error().Err(err).Msg("Failed to acknowledge interaction")
		}
		// Apply the modal under the channel lock, then republish through the shared flow
		// so multi-message tables stay consistent and never exceed the 2000-character limit.
		if err := handler.syncChannel(session, func() {
			ApplyModalInput(handler.pantryClient, modalInput)
		}); err != nil {
			log.Error().Err(err).Msgf("Could not sync channel `%s` on modal submit", handler.channelID)
		}
	default:
		log.Error().Msgf("Could not map modal-submit interaction event `%s`", interaction.ModalSubmitData().CustomID)
	}
}
