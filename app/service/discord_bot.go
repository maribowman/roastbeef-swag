package service

import (
	"encoding/base64"
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
	service := DiscordService{}
	tokenBytes, err := base64.StdEncoding.DecodeString(config.Config.Discord.Token)
	if err != nil {
		log.Fatal().Err(err).Msg("could not decode token")
	}
	session, err := discordgo.New("Bot " + string(tokenBytes))
	if err != nil {
		log.Fatal().Err(err).Msg("error creating discord session")
	}

	session.AddHandler(service.DispatchHandler)
	session.Identify.Intents = discordgo.IntentsGuildMessages

	if err = session.Open(); err != nil {
		log.Fatal().Err(err).Msg("could not open discord session")
	}

	service.session = session
	service.groceryBot = NewGroceryBot(config.Config.Discord.BotID)

	return &service
}

func (service *DiscordService) DispatchHandler(session *discordgo.Session, message *discordgo.MessageCreate) {
	switch message.ChannelID {
	case "1084632136180572230":
		service.groceryBot.MessageEvent(session, message)
	}
}

func (service *DiscordService) CloseSession() {
	if err := service.session.Close(); err != nil {
		log.Error().Err(err).Msg("could not close discord session")
	}
}
