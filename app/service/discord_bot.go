package service

import (
	"encoding/base64"
	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/config"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/maribowman/roastbeef-swag/app/repository"
	"github.com/rs/zerolog/log"
)

type DiscordBot struct {
	session       *discordgo.Session
	groceryClient model.GroceryClient
}

func NewDiscordBot() model.DiscordBot {
	tokenBytes, err := base64.StdEncoding.DecodeString(config.Config.Discord.Token)
	if err != nil {
		log.Fatal().Err(err).Msg("could not decode token")
	}
	session, err := discordgo.New("Bot " + string(tokenBytes))
	if err != nil {
		log.Fatal().Err(err).Msg("error creating discord session")
	}

	groceryClient := repository.NewGroceryClient(session, config.Config.Discord.BotID)

	session.Identify.Intents = discordgo.IntentsGuildMessages

	if err = session.Open(); err != nil {
		log.Fatal().Err(err).Msg("could not open discord session")
	}

	return &DiscordBot{
		session:       session,
		groceryClient: groceryClient,
	}
}

func (bot *DiscordBot) CloseSession() {
	if err := bot.session.Close(); err != nil {
		log.Error().Err(err).Msg("could not close discord session")
	}
}
