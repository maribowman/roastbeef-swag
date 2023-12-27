package service

import (
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
	removeRegex      = regexp.MustCompile(`^(\*)?(?:\s*\d+)*\s*(\d+-\d+)?$`)
	leadingQuantity  = regexp.MustCompile(`^(\d+)\s.*`)
	trailingQuantity = regexp.MustCompile(`\s(\d+)$`)
)

type GroceryBot struct {
	botID                     string
	channelID                 string
	lastMessage               *discordgo.Message
	shoppingList              []model.ShoppingEntry
	previousShoppingListTable string
}

/*
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
		return
	}

	channelMessages, err := session.ChannelMessages(message.ChannelID, 100, message.ID, "", "")
	if err != nil {
		return
	}
	removableMessageIDs := []string{message.ID}
	for _, msg := range channelMessages {
		if msg.Author.ID == bot.botID {
			if bot.lastMessage == nil {
				bot.lastMessage = msg
				// TODO parse shopping list from msg.content
				continue
			} else if bot.lastMessage.Timestamp.After(msg.Timestamp) {
				bot.lastMessage = msg
				continue
			} else if bot.lastMessage.ID == msg.ID {
				continue
			}
		}
		removableMessageIDs = append(removableMessageIDs, msg.ID)
	}

	for _, line := range strings.Split(message.Content, ini.LineBreak) {
		line = strings.TrimSpace(line)

		if removeRegex.MatchString(line) {
			bot.remove(line)
		} else {
			bot.add(line)
		}
	}

	if err := session.ChannelMessagesBulkDelete(message.ChannelID, removableMessageIDs); err != nil {
		log.Error().Err(err).Msg("could not bulk delete channel messages")
	}

	bot.publish(session, message.ChannelID)
}

func (bot *GroceryBot) InteractionEvent(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	switch interaction.MessageComponentData().CustomID {
	case "edit":
		// TODO send modal to edit list
		break
	case "done":
		bot.shoppingList = []model.ShoppingEntry{}
	}
	bot.publish(session, interaction.ChannelID)
}

func (bot *GroceryBot) remove(line string) {
	var tempShoppingList []model.ShoppingEntry
	removeAllExcept := false

	// CAPTURE GROUP 0: entire string
	// CAPTURE GROUP 1: asterisk
	// CAPTURE GROUP 2: range
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
		entry.ID = len(tempShoppingList) + 1 // comment-out to run remove unit tests
		tempShoppingList = append(tempShoppingList, entry)
	}
	bot.shoppingList = tempShoppingList
}

func (bot *GroceryBot) add(line string) {
	leading := leadingQuantity.FindStringSubmatch(line)
	trailing := trailingQuantity.FindStringSubmatch(line)

	var quantity string
	if leading != nil {
		quantity = leading[1]
		line = strings.TrimPrefix(line, quantity)
	} else if trailing != nil {
		quantity = trailing[1]
		line = strings.TrimSuffix(line, quantity)
	}

	amount, err := strconv.Atoi(quantity)
	if err != nil {
		amount = 1
	}

	bot.shoppingList = append(bot.shoppingList, model.ShoppingEntry{
		ID:     len(bot.shoppingList) + 1,
		Item:   strings.TrimSpace(line),
		Amount: amount,
		Date:   time.Now().Truncate(time.Minute),
	})
}

func (bot *GroceryBot) publish(session *discordgo.Session, channelID string) {
	// edit message if exists
	if bot.lastMessage != nil {
		editedMessage := discordgo.NewMessageEdit(bot.lastMessage.ChannelID, bot.lastMessage.ID)
		editedMessage.SetContent(model.CreateShoppingListTable(bot.shoppingList))
		if _, err := session.ChannelMessageEditComplex(editedMessage); err != nil {
			log.Error().Err(err).Msgf("could not edit message %s", bot.lastMessage.ID)
		}
		return
	}

	// send new message
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
	}); err != nil {
		log.Error().Err(err).Msg("could not send complex message")
	}
}
