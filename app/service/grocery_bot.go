package service

import (
	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/config"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/rs/zerolog/log"
	"regexp"
	"strconv"
	"time"
)

const groceries_channel_name = "groceries"

type GroceryBot struct {
	botID            string
	groceryChannelID string
	shoppingList     []model.ShoppingEntry
}

/*
 - add dynamic quantity to item
 - parse and add multiline items
 - only edit and not send/delete message
 - parse previous table -> re-instantiate from channel
 -- only 1 message from bot possible if implemented correctly
 - clear all items
*/

func NewGroceryBot(botID string) model.DiscordBot {
	log.Debug().Msg("registering grocery client handler")
	client := GroceryBot{}
	client.botID = botID
	for _, channel := range config.Config.Discord.Channels {
		if channel.Name == groceries_channel_name {
			client.groceryChannelID = channel.ID
			break
		}
	}
	return &client
}

func (bot *GroceryBot) MessageEvent(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.ChannelID != bot.groceryChannelID {
		return
	}
	if message.Author.ID == bot.botID {
		messages, err := session.ChannelMessages(bot.groceryChannelID, 100, message.Message.ID, "", "")
		if err != nil {
			return
		}
		var messageIDs []string
		for _, msg := range messages {
			if msg.Author.ID == bot.botID {
				messageIDs = append(messageIDs, msg.ID)
			}
		}
		session.ChannelMessagesBulkDelete(bot.groceryChannelID, messageIDs)
		return
	}

	if regexp.MustCompile(`\d`).MatchString(message.Content) {
		bot.removeItemFromShoppingList(message.Content)
	} else {
		bot.shoppingList = append(bot.shoppingList, model.ShoppingEntry{
			ID:     len(bot.shoppingList),
			Item:   message.Content,
			Amount: 1,
			Date:   time.Now(),
		})
	}

	shoppingListTable := model.CreateShoppingListTable(bot.shoppingList)
	if _, err := session.ChannelMessageSend(message.ChannelID, shoppingListTable); err != nil {
		log.Error().Err(err).Msg("could not send message")
	}

	// clean up chat history
	if err := session.ChannelMessageDelete(message.ChannelID, message.ID); err != nil {
		log.Error().Err(err).Msg("could not delete previous message")
	}
}

func (bot *GroceryBot) removeItemFromShoppingList(content string) {
	id, _ := strconv.Atoi(content)

	var tempShoppingList []model.ShoppingEntry
	for _, entry := range bot.shoppingList {
		if entry.ID == id {
			continue
		}
		entry.ID = len(tempShoppingList)
		tempShoppingList = append(tempShoppingList, entry)
	}

	bot.shoppingList = tempShoppingList
}
