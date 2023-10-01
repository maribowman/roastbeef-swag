package repository

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

type GroceryClient struct {
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

func NewGroceryClient(session *discordgo.Session, botID string) model.GroceryClient {
	log.Debug().Msg("registering grocery client handler")
	client := GroceryClient{}
	client.botID = botID
	for _, channel := range config.Config.Discord.Channels {
		if channel.Name == groceries_channel_name {
			client.groceryChannelID = channel.ID
			break
		}
	}
	session.AddHandler(client.groceryAction)
	return &client
}

func (client *GroceryClient) groceryAction(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.ChannelID != client.groceryChannelID {
		return
	}
	if message.Author.ID == client.botID {
		messages, err := session.ChannelMessages(client.groceryChannelID, 100, message.Message.ID, "", "")
		if err != nil {
			return
		}
		var messageIDs []string
		for _, msg := range messages {
			if msg.Author.ID == client.botID {
				messageIDs = append(messageIDs, msg.ID)
			}
		}
		session.ChannelMessagesBulkDelete(client.groceryChannelID, messageIDs)
		return
	}

	if regexp.MustCompile(`\d`).MatchString(message.Content) {
		client.removeItemFromShoppingList(message.Content)
	} else {
		client.shoppingList = append(client.shoppingList, model.ShoppingEntry{
			ID:     len(client.shoppingList),
			Item:   message.Content,
			Amount: 1,
			Date:   time.Now(),
		})
	}

	shoppingListTable := model.CreateShoppingListTable(client.shoppingList)
	if _, err := session.ChannelMessageSend(message.ChannelID, shoppingListTable); err != nil {
		log.Error().Err(err).Msg("could not send message")
	}

	// clean up chat history
	if err := session.ChannelMessageDelete(message.ChannelID, message.ID); err != nil {
		log.Error().Err(err).Msg("could not delete previous message")
	}
}

func (client *GroceryClient) removeItemFromShoppingList(content string) {
	id, _ := strconv.Atoi(content)

	var tempShoppingList []model.ShoppingEntry
	for _, entry := range client.shoppingList {
		if entry.ID == id {
			continue
		}
		entry.ID = len(tempShoppingList)
		tempShoppingList = append(tempShoppingList, entry)
	}

	client.shoppingList = tempShoppingList
}
