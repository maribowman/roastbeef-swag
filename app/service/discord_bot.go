package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/config"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/rs/zerolog/log"
)

type DiscordBot struct {
	session  *discordgo.Session
	handlers map[string]model.BotHandler // Maps channel ID to corresponding handler
}

func NewDiscordBot(databaseClient model.DatabaseClient) model.DiscordBot {
	session, err := discordgo.New(fmt.Sprintf("Bot %s", config.Config.Discord.Token))
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating Discord session")
	}

	handlers := map[string]model.BotHandler{}
	for _, channel := range config.Config.Discord.Channels {
		switch channel.Name {
		case GroceriesChannel:
			handlers[channel.ID] = NewGroceryHandler(channel.ID, databaseClient, channel.LineBreak)
			continue
		case TkGoodsChannel:
			handlers[channel.ID] = NewTkHandler(channel.ID, databaseClient, channel.LineBreak)
			continue
		}
		log.Info().Msgf("Could not map channel `%s` to handler", channel.Name)
	}

	bot := DiscordBot{
		session:  session,
		handlers: handlers,
	}

	bot.session.AddHandler(bot.Ready)
	bot.session.AddHandler(bot.MessageDispatch)
	bot.session.AddHandler(bot.InteractionDispatch)

	if err = bot.session.Open(); err != nil {
		log.Fatal().Err(err).Msg("Could not open session with Discord server")
	}
	log.Info().Msg("Session to Discord server established, bot is up!")

	return &bot
}

func (bot *DiscordBot) Ready(session *discordgo.Session, _ *discordgo.Ready) {
	var waitGroup sync.WaitGroup
	errorChannel := make(chan error, len(bot.handlers))

	for channelID, handler := range bot.handlers {
		waitGroup.Add(1)

		go func(channelID string, handler model.BotHandler) {
			defer waitGroup.Done()

			timeout := time.After(5 * time.Second)
			done := make(chan struct{})

			go func() {

				if err := handler.ReadyEvent(session); err != nil {
					errorChannel <- fmt.Errorf("failed to initialize handler on channel %s: %w", channelID, err)
				}
				close(done)
			}()

			select {
			case <-timeout:
				errorChannel <- fmt.Errorf("handler initialization on channel  %s timed out", channelID)
			case <-done:
				log.Info().Msgf("Initialized handler on channel %s", channelID)
			}
		}(channelID, handler)
	}

	waitGroup.Wait()
	close(errorChannel)

	var hasErrors bool
	for err := range errorChannel {
		hasErrors = true
		log.Error().Err(err).Msg("Handler initialization failed")
	}

	if hasErrors {
		log.Fatal().Msg("Failed to initialize all handlers -> crashing bot!")
	}

	log.Info().Msg("All handlers initialized, bot is ready!")
}

func (bot *DiscordBot) MessageDispatch(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == config.Config.Discord.BotID {
		return
	}

	if handler, ok := bot.handlers[message.ChannelID]; ok {
		handler.MessageEvent(session, message)
	} else {
		log.Error().Msgf("Could not match handler for message event on channel `%s`", message.ChannelID)
	}
}

func (bot *DiscordBot) InteractionDispatch(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if handler, ok := bot.handlers[interaction.ChannelID]; ok {
		switch interaction.Type {
		case discordgo.InteractionApplicationCommand:
			// Slash commands
		case discordgo.InteractionMessageComponent:
			// Button interaction
			handler.MessageComponentInteractionEvent(session, interaction)
		case discordgo.InteractionModalSubmit:
			// Modal update
			handler.ModalSubmitInteractionEvent(session, interaction)
		default:
			log.Error().Msgf("Could not dispatch interaction event with type `%s`", interaction.Type)
		}
	} else {
		log.Error().Msgf("Could not match handler for interaction event on channel `%s`", interaction.ChannelID)
	}
}

func (bot *DiscordBot) CloseSession() {
	if err := bot.session.Close(); err != nil {
		log.Error().Err(err).Msg("Could not close Discord session")
	}
}
