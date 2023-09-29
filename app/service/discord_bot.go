package service

import (
	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/config"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/maribowman/roastbeef-swag/app/repository"
	"log"
)

type DiscordBot struct {
	groceryClient model.GroceryClient
}

func NewDiscordBot() model.DiscordBot {
	session, err := discordgo.New("Bot " + config.Config.Discord.Token)
	if err != nil {
		log.Fatal("error creating discord session", err)
	}

	groceryClient := repository.NewGroceryClient(session)

	session.Identify.Intents = discordgo.IntentsGuildMessages

	if err = session.Open(); err != nil {
		log.Fatal("could not open session", err)
	}

	return &DiscordBot{
		groceryClient: groceryClient,
	}
}
