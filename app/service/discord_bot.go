package service

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/config"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/rs/zerolog/log"
)

type DiscordBot struct {
	session  *discordgo.Session
	handlers map[string]model.BotHandler
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
		log.Error().Msgf("Could not map channel `%s` to handler", channel.Name)
	}

	bot := DiscordBot{
		session:  session,
		handlers: handlers,
	}

	bot.session.AddHandler(bot.Ready)
	bot.session.AddHandler(bot.MessageDispatch)
	bot.session.AddHandler(bot.InteractionDispatch)

	if err = bot.session.Open(); err != nil {
		log.Fatal().Err(err).Msg("Could not open Discord session")
	}

	return &bot
}

func (bot *DiscordBot) Ready(session *discordgo.Session, ready *discordgo.Ready) {
	for _, handler := range bot.handlers {
		handler.ReadyEvent(session, ready)
	}
	log.Info().Msg("Bot is up!")
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
			// slash commands
		case discordgo.InteractionMessageComponent:
			handler.MessageComponentInteractionEvent(session, interaction)
		case discordgo.InteractionModalSubmit:
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
