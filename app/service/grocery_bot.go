package service

import (
	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/rs/zerolog/log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	GroceriesChannelName = "groceries"

	add    = "add"
	update = "update"
	remove = "remove"
	undo   = "undo"
)

var (
	updateRegex = regexp.MustCompile("^[0-9]+( [a-zA-Z]*)( .*)*$")
	removeRegex = regexp.MustCompile("^(\\*( [0-9]+)*|[0-9]+( [0-9]+)*)$")
	//removeRegex = regexp.MustCompile("^\\*[ [0-9]+]*|[0-9]+[ [0-9]+]*$")
	undoRegex = regexp.MustCompile("^&$")
)

type GroceryBot struct {
	botID                     string
	channelID                 string
	shoppingList              []model.ShoppingEntry
	previousShoppingListTable string
}

/*
 - add dynamic quantity to item
 - parse and add multiline items
 - only edit and not send/delete message
 - parse previous table -> re-instantiate from channel
 -- only 1 message from bot possible if implemented correctly
 - clear all items
*/

func NewGroceryBot(botID string, channelID string) model.DiscordBot {
	log.Debug().Msg("registering grocery client handler")
	return &GroceryBot{
		botID:     botID,
		channelID: channelID,
	}
}

func (bot *GroceryBot) MessageEvent(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == bot.botID {
		messages, err := session.ChannelMessages(bot.channelID, 100, message.Message.ID, "", "")
		if err != nil {
			return
		}
		var messageIDs []string
		for _, msg := range messages {
			if msg.Author.ID == bot.botID {
				messageIDs = append(messageIDs, msg.ID)
			}
		}
		session.ChannelMessagesBulkDelete(bot.channelID, messageIDs)
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

func (bot *GroceryBot) ParseContent(content string) string {
	strings.TrimSpace(content)
	if updateRegex.MatchString(content) {
		return update
	}
	if removeRegex.MatchString(content) {
		return remove
	}
	if undoRegex.MatchString(content) {
		return undo
	}
	return add
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
