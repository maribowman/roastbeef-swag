package service

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/config"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/rs/zerolog/log"
)

type DiscordBotService struct {
	session        *discordgo.Session
	groceryHandler model.BotHandler
	tkHandler      model.BotHandler
}

func NewDiscordService() model.DiscordBot {
	session, err := discordgo.New(fmt.Sprintf("Bot %s", config.Config.Discord.Token))
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating discord session")
	}

	service := DiscordBotService{
		session: session,
		// bots: []model.DiscordBot{}, // TODO accumulate bots as kv-pairs (name, discord-bot)
		groceryHandler: NewGroceryHandler(config.Config.Discord.BotID, config.Config.Discord.Channels[GroceriesChannelName]),
		tkHandler:      NewTkHandler(config.Config.Discord.BotID, config.Config.Discord.Channels[TkGoodsChannelName]),
	}

	service.session.AddHandler(service.ReadyHandler)
	service.session.AddHandler(service.MessageDispatchHandler)
	service.session.AddHandler(service.InteractionDispatchHandler)

	if err = service.session.Open(); err != nil {
		log.Fatal().Err(err).Msg("Could not open discord session")
	}

	return &service
}

func (service *DiscordBotService) ReadyHandler(session *discordgo.Session, ready *discordgo.Ready) {
	service.groceryHandler.ReadyEvent(session, ready)
	log.Info().Msg("Bot is up!")
}

func (service *DiscordBotService) MessageDispatchHandler(session *discordgo.Session, message *discordgo.MessageCreate) {
	switch message.ChannelID {
	case config.Config.Discord.Channels[GroceriesChannelName]:
		service.groceryHandler.MessageEvent(session, message)
	default:
		log.Debug().Msg("Could not dispatch message event")
	}
}

func (service *DiscordBotService) InteractionDispatchHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var handler model.BotHandler
	switch interaction.ChannelID {
	case config.Config.Discord.Channels[GroceriesChannelName]:
		handler = service.groceryHandler
	default:
		log.Debug().Msgf("Could not match handler for interaction event on channel `%s`", interaction.ChannelID)
		return
	}

	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		// slash commands
	case discordgo.InteractionMessageComponent:
		handler.MessageComponentInteractionEvent(session, interaction)
	case discordgo.InteractionModalSubmit:
		handler.ModalSubmitInteractionEvent(session, interaction)
	default:
		log.Debug().Msgf("Could not dispatch interaction event with type `%s`", interaction.Type)
	}
}

func (service *DiscordBotService) CloseSession() {
	if err := service.session.Close(); err != nil {
		log.Error().Err(err).Msg("Could not close discord session")
	}
}
