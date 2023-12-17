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
	addRegex    = regexp.MustCompile("^(.*?)( [0-9]+)?$")
	updateRegex = regexp.MustCompile("^([0-9]+)( [a-zA-Z]*)( .*)*$")
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
 	-> only 1 message from bot possible if implemented correctly
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

	// TODO getCurrentShoppingListTable()
	var resultTable string

	switch bot.ParseContent(message.Content) {
	case add:
		// TODO iterate content for each line and parse quantity at end
		bot.shoppingList = append(bot.shoppingList, model.ShoppingEntry{
			ID:     len(bot.shoppingList),
			Item:   message.Content,
			Amount: 1,
			Date:   time.Now(),
		})
		resultTable = model.CreateShoppingListTable(bot.shoppingList)
	case update:
		resultTable = model.CreateShoppingListTable(bot.shoppingList)
	case remove:
		bot.removeItemFromShoppingList(message.Content)
		resultTable = model.CreateShoppingListTable(bot.shoppingList)
	case undo:
		resultTable = bot.previousShoppingListTable
	}

	if _, err := s.ApplicationCommandCreate(*AppID, *GuildID, &discordgo.ApplicationCommand{
		Name:        "buttons",
		Description: "Test the buttons if you got courage",
	}); err != nil {
		log.Error().Err(err).Msg("Cannot create button command")
	}

	// TODO edit instead of new message
	if _, err := session.ChannelMessageSend(message.ChannelID, resultTable); err != nil {
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
	// TODO refactor this
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
