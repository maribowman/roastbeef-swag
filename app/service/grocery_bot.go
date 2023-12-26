package service

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/rs/zerolog/log"
	"gopkg.in/ini.v1"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	GroceriesChannelName = "groceries"
)

var (
	removeRegex             = regexp.MustCompile(`(\*)?(?:\s*\d+)*\s*(\d+-\d+)?$`)
	addRegex                = regexp.MustCompile(`^(?:(\d*)\s)?([a-zA-Z0-9\s-_]*)?$`)
	addWithLeadingQuantity  = "^(\\d*)?.*"
	addWithTrailingQuantity = "(\\d*)?$"
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
 - only update existing shopping list message
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
		_ = session.ChannelMessagesBulkDelete(bot.channelID, messageIDs)
		return
	}

	// TODO getCurrentShoppingListTable()
	for _, line := range strings.Split(message.Content, ini.LineBreak) {
		line = strings.TrimSpace(line)

		if addRegex.MatchString(line) {
			bot.add(line)
		} else if removeRegex.MatchString(line) {
			bot.remove(line)
		} else {
			_, err := session.ChannelMessageSendReply(message.ChannelID, fmt.Sprintf("Cannot process input:\n> %s", message.Content), message.MessageReference)
			if err != nil {
				log.Error().Err(err).Msg("could not answer message")
			}
			return
		}
	}

	// TODO edit instead of new message
	//if _, err := session.ChannelMessageSend(message.ChannelID, resultTable); err != nil {
	//	log.Error().Err(err).Msg("could not send message")
	//}

	// clean up chat history
	if err := session.ChannelMessageDelete(message.ChannelID, message.ID); err != nil {
		log.Error().Err(err).Msg("could not delete previous message")
	}

	bot.publishShoppingList(session, message.ChannelID)
}

func (bot *GroceryBot) InteractionEvent(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	switch interaction.MessageComponentData().CustomID {
	//case "edit":
	//case "done":
	default:
		bot.publishShoppingList(session, interaction.ChannelID)
	}
}

func (bot *GroceryBot) remove(line string) {
	var tempShoppingList []model.ShoppingEntry
	removeAllExcept := false

	// CAP GROUP 0: entire string
	// CAP GROUP 1: asterisk
	// CAP GROUP 2: range
	captureGroups := removeRegex.FindStringSubmatch(line)

	// remove all (except)
	if captureGroups[1] == "*" {
		removeAllExcept = true
		if captureGroups[0] == captureGroups[1] {
			bot.shoppingList = tempShoppingList
			return
		}
	}

	// add single removable IDs
	var ids []int
	if captureGroups[0] != captureGroups[2] {
		for _, value := range strings.Split(captureGroups[0], " ") {
			if id, err := strconv.Atoi(value); err == nil {
				ids = append(ids, id)
			}
		}
	}

	// add range to removable IDs
	if captureGroups[2] != "" {
		range_ := strings.Split(captureGroups[2], "-")
		rangeStart, _ := strconv.Atoi(range_[0])
		rangeEnd, _ := strconv.Atoi(range_[1])

		for i := rangeStart; i <= rangeEnd; i++ {
			ids = append(ids, i)
		}
	}

	for _, entry := range bot.shoppingList {
		if slices.Contains(ids, entry.ID) {
			if !removeAllExcept {
				continue
			}
		} else if removeAllExcept {
			continue
		}
		// entry.ID = len(tempShoppingList) + 1 // TODO think about new indexes
		tempShoppingList = append(tempShoppingList, entry)
	}
	bot.shoppingList = tempShoppingList
}

func (bot *GroceryBot) add(line string) {
	captureGroups := addRegex.FindStringSubmatch(line)

	amount := 1
	//if quantity, err := strconv.Atoi(captureGroups[1]); err == nil {
	//	amount = quantity
	//}
	//if quantity, err := strconv.Atoi(captureGroups[3]); err == nil {
	//	amount = quantity
	//}

	bot.shoppingList = append(bot.shoppingList, model.ShoppingEntry{
		ID:     len(bot.shoppingList) + 1,
		Item:   captureGroups[2],
		Amount: amount,
		Date:   time.Now().Truncate(time.Minute),
	})
}

func (bot *GroceryBot) publishShoppingList(session *discordgo.Session, channelID string) {
	if _, err := session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: model.CreateShoppingListTable(bot.shoppingList),
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Emoji: discordgo.ComponentEmoji{
							Name: "📝",
						},
						Style:    discordgo.SecondaryButton,
						CustomID: "edit",
					},
					discordgo.Button{
						Emoji: discordgo.ComponentEmoji{
							Name: "🏁",
						},
						Style:    discordgo.SecondaryButton,
						CustomID: "done",
					},
				},
			},
		},
		Files:           nil,
		AllowedMentions: nil,
		Reference:       nil,
		File:            nil,
		Embed:           nil,
	}); err != nil {
		log.Error().Err(err).Msg("could not send complex message")
	}
}
