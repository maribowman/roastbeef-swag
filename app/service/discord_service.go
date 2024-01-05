package service

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/config"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/rs/zerolog/log"
)

type DiscordService struct {
	session    *discordgo.Session
	groceryBot model.DiscordBot
}

func NewDiscordService() model.DiscordService {
	session, err := discordgo.New(fmt.Sprintf("Bot %s", config.Config.Discord.Token))
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating discord session")
	}

	service := DiscordService{
		session: session,
		// bots: []model.DiscordBot{}, // TODO accumulate bots as kv-pairs (name, discord-bot)
		groceryBot: NewGroceryBot(config.Config.Discord.BotID, config.Config.Discord.Channels[GroceriesChannelName]),
	}

	service.session.AddHandler(service.ReadyHandler)
	service.session.AddHandler(service.MessageDispatchHandler)
	service.session.AddHandler(service.InteractionDispatchHandler)

	if err = service.session.Open(); err != nil {
		log.Fatal().Err(err).Msg("Could not open discord session")
	}

	return &service
}

func (service *DiscordService) ReadyHandler(session *discordgo.Session, ready *discordgo.Ready) {
	service.groceryBot.ReadyEvent(session, ready)
	log.Info().Msg("Bot is up!")
}

func (service *DiscordService) MessageDispatchHandler(session *discordgo.Session, message *discordgo.MessageCreate) {
	switch message.ChannelID {
	case config.Config.Discord.Channels[GroceriesChannelName]:
		service.groceryBot.MessageEvent(session, message)
	default:
		log.Debug().Msg("Could not dispatch message event")
	}
}

func (service *DiscordService) InteractionDispatchHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var bot model.DiscordBot
	switch interaction.ChannelID {
	case config.Config.Discord.Channels[GroceriesChannelName]:
		bot = service.groceryBot
	default:
		log.Debug().Msgf("Could not match handler for interaction event on channel `%s`", interaction.ChannelID)
		return
	}

	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		// slash commands
	case discordgo.InteractionMessageComponent:
		bot.MessageComponentInteractionEvent(session, interaction)
	case discordgo.InteractionModalSubmit:
		bot.ModalSubmitInteractionEvent(session, interaction)
	default:
		log.Debug().Msgf("Could not dispatch interaction event with type `%s`", interaction.Type)
	}
}

func (service *DiscordService) CloseSession() {
	if err := service.session.Close(); err != nil {
		log.Error().Err(err).Msg("Could not close discord session")
	}
}
