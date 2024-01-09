package service

import (
	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/config"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/rs/zerolog/log"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	GroceriesChannel = "groceries"
	TkGoodsChannel   = "tkgoods"

	EditButton     = "edit-button"
	DoneButton     = "done-button"
	EditModal      = "edit-modal"
	EditModalInput = "edit-modal-input"
)

var (
	removeRegex      = regexp.MustCompile(`^(\*)?(?:\s*\d+)*\s*(\d+-\d+)?$`)
	leadingQuantity  = regexp.MustCompile(`^(\d+)\s.*`)
	trailingQuantity = regexp.MustCompile(`\s(\d+)$`)
)

func PreProcessMessageEvent(
	session *discordgo.Session,
	message *discordgo.MessageCreate,
) (
	items []model.PantryItem,
	lastBotMessageID string,
	content string,
	removableMessageIDs []string,
	err error,
) {
	channelMessages, err_ := session.ChannelMessages(message.ChannelID, 100, "", "", "")
	if err_ != nil {
		err = err_
		return
	}

	var lastBotMessage *discordgo.Message
	for _, msg := range channelMessages {
		if msg.Author.ID == config.Config.Discord.BotID {
			if lastBotMessage == nil {
				lastBotMessage = msg
				lastBotMessageID = msg.ID
				items = model.FromMarkdownTable(msg.Content)
				continue
			} else if lastBotMessage.Timestamp.After(msg.Timestamp) {
				removableMessageIDs = append(removableMessageIDs, lastBotMessageID) // remove previous bot msg
				lastBotMessage = msg
				lastBotMessageID = msg.ID
				items = model.FromMarkdownTable(msg.Content)
				continue
			}
		} else {
			content += "\n" + msg.Content
		}
		removableMessageIDs = append(removableMessageIDs, msg.ID)
	}
	return
}

func UpdateHandlerItems(items []model.PantryItem, content string) []model.PantryItem {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if removeRegex.MatchString(line) {
			items = remove(items, line)
		} else {
			items = add(items, line)
		}
	}
	return items
}

func remove(items []model.PantryItem, line string) []model.PantryItem {
	result := make([]model.PantryItem, 0)
	removeAllExcept := false

	// CAPTURE GROUP 0: entire string
	// CAPTURE GROUP 1: asterisk
	// CAPTURE GROUP 2: range
	captureGroups := removeRegex.FindStringSubmatch(line)

	// remove all (except)
	if captureGroups[1] == "*" {
		removeAllExcept = true
		if captureGroups[0] == captureGroups[1] {
			return result
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

	for _, entry := range items {
		if slices.Contains(ids, entry.ID) {
			if !removeAllExcept {
				continue
			}
		} else if removeAllExcept {
			continue
		}
		entry.ID = len(result) + 1
		result = append(result, entry)
	}
	return result
}

func add(items []model.PantryItem, line string) []model.PantryItem {
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

	return append(items, model.PantryItem{
		ID:     len(items) + 1,
		Item:   strings.TrimSpace(line),
		Amount: amount,
		Date:   time.Now().Truncate(time.Minute),
	})
}

func PublishItems(items []model.PantryItem, session *discordgo.Session, channelID, messageID, dateFormat string) {
	if messageID != "" {
		editedMessage := discordgo.NewMessageEdit(channelID, messageID)
		editedMessage.SetContent(model.ToMarkdownTable(items, dateFormat))
		if _, err := session.ChannelMessageEditComplex(editedMessage); err != nil {
			log.Error().Err(err).Msgf("Could not edit message %s", messageID)
		}
		return
	}

	if _, err := session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content:    model.ToMarkdownTable(items, dateFormat),
		Components: CreateMessageButtons(),
	}); err != nil {
		log.Error().Err(err).Msg("Could not send complex message")
	}
}

func CreateMessageButtons() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Emoji: discordgo.ComponentEmoji{
						Name: "📝",
					},
					Style:    discordgo.SecondaryButton,
					CustomID: EditButton,
				},
				discordgo.Button{
					Emoji: discordgo.ComponentEmoji{
						Name: "🏁",
					},
					Style:    discordgo.SecondaryButton,
					CustomID: DoneButton,
				},
			},
		},
	}
}
