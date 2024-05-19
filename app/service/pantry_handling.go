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
	TkGoodsChannel   = "tkGoods"

	EditButton     = "edit-button"
	UndoButton     = "undo-button"
	EditModal      = "edit-modal"
	EditModalInput = "edit-modal-input"
)

var (
	NumberPrefixRegex = regexp.MustCompile(`^\[(\d+)]\s`)
	removeRegex       = regexp.MustCompile(`^(\*)?(?:\s*\d+)*\s*(\d+-\d+)?$`)
	leadingQuantity   = regexp.MustCompile(`^(\d+)\s.*`)
	trailingQuantity  = regexp.MustCompile(`\s(\d+)$`)
)

func PreProcessMessageEvent(session *discordgo.Session, channelID, dateFormat string) (
	items []model.PantryItem,
	lastBotMessageID string,
	content string,
	removableMessageIDs []string,
	err error,
) {
	channelMessages, err_ := session.ChannelMessages(channelID, 100, "", "", "")
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
				items = model.FromMarkdownTable(msg.Content, dateFormat)
				continue
			} else if lastBotMessage.Timestamp.After(msg.Timestamp) {
				removableMessageIDs = append(removableMessageIDs, lastBotMessageID) // remove previous bot msg
				lastBotMessage = msg
				lastBotMessageID = msg.ID
				items = model.FromMarkdownTable(msg.Content, dateFormat)
				// TODO read from file if present
				continue
			}
		} else {
			content += "\n" + msg.Content
		}
		removableMessageIDs = append(removableMessageIDs, msg.ID)
	}
	return
}

func UpdateItemsFromList(items []model.PantryItem, updatedList string) []model.PantryItem {
	var updatedItems []model.PantryItem
	var newItems []string

	updates := strings.Split(updatedList, "\n")
	for _, update := range updates {
		if strings.TrimSpace(update) == "" {
			continue
		}

		rawNumber := NumberPrefixRegex.FindStringSubmatch(update)
		item := strings.TrimSpace(NumberPrefixRegex.ReplaceAllString(update, ""))
		var number int
		if len(rawNumber) == 2 { // regex matches full string + capture group
			number, _ = strconv.Atoi(rawNumber[1])
		} else { // add new item at the end if there's no number
			newItems = append(newItems, item)
			continue
		}
		var getOldItemDate = func(oldItems []model.PantryItem) time.Time {
			for _, oldItem := range oldItems {
				if oldItem.Number == number {
					return oldItem.Date
				}
			}
			return time.Now().Truncate(time.Minute)
		}

		updatedItems = add(updatedItems, item, getOldItemDate(items))
	}

	for _, newItem := range newItems {
		updatedItems = add(updatedItems, newItem, time.Now().Truncate(time.Minute))
	}
	return updatedItems
}

func UpdateItems(items []model.PantryItem, content string) []model.PantryItem {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if removeRegex.MatchString(line) {
			items = remove(items, line)
		} else {
			items = add(items, line, time.Now().Truncate(time.Minute))
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

	// add single removable numbers
	var numbers []int
	if captureGroups[0] != captureGroups[2] {
		for _, value := range strings.Split(captureGroups[0], " ") {
			if number, err := strconv.Atoi(value); err == nil {
				numbers = append(numbers, number)
			}
		}
	}

	// add range to removable numbers
	if captureGroups[2] != "" {
		range_ := strings.Split(captureGroups[2], "-")
		rangeStart, _ := strconv.Atoi(range_[0])
		rangeEnd, _ := strconv.Atoi(range_[1])

		for i := rangeStart; i <= rangeEnd; i++ {
			numbers = append(numbers, i)
		}
	}

	for _, entry := range items {
		if slices.Contains(numbers, entry.Number) {
			if !removeAllExcept {
				continue
			}
		} else if removeAllExcept {
			continue
		}
		entry.Number = len(result) + 1
		result = append(result, entry)
	}
	return result
}

func add(items []model.PantryItem, line string, date time.Time) []model.PantryItem {
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
		Number: len(items) + 1,
		Item:   strings.TrimSpace(line),
		Amount: amount,
		Date:   date,
	})
}

// PublishItems sends the latest []PantryItem state to the active channel (via new or edited message).
// Because of a character limit of 2000, the function automatically splits the Markdown table line by line
// and sends multiple messages to the channel. The last message always contains the buttons to interact with the bot.
func PublishItems(items []model.PantryItem, session *discordgo.Session, channelID, messageID string, lineBreak int, dateFormat string) {
	markdownTable := model.ToMarkdownTable(items, lineBreak, dateFormat)

	if len(markdownTable) <= 2000 { // 2000 is the message length limit
		if messageID != "" { // update existing message
			editedMessage := discordgo.NewMessageEdit(channelID, messageID)
			editedMessage.SetContent(markdownTable)
			if _, err := session.ChannelMessageEditComplex(editedMessage); err != nil {
				log.Error().Err(err).Msgf("Could not edit message %s", messageID)
			}
		} else {
			if _, err := session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
				Content:    model.ToMarkdownTable(items, lineBreak, dateFormat),
				Components: CreateMessageButtons(),
			}); err != nil {
				log.Error().Err(err).Msg("Could not send complex message")
			}
		}
	} else { // split table line by line
		markdownTableSplit := strings.Split(markdownTable, "\n")
		tempTable := ""

		for _, line := range markdownTableSplit {
			if len(tempTable)+len(line) <= 1980 {
				tempTable += line + "\n"
				continue
			}
			tempTable += "...```"

			if messageID != "" {
				editedMessage := discordgo.NewMessageEdit(channelID, messageID)
				editedMessage.SetContent(tempTable)
				if _, err := session.ChannelMessageEditComplex(editedMessage); err != nil {
					log.Error().Err(err).Msgf("Could not edit message %s", messageID)
				}
			} else {
				if _, err := session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
					Content:    tempTable,
					Components: CreateMessageButtons(),
				}); err != nil {
					log.Error().Err(err).Msg("Could not send complex message")
				}
			}
			return
		}
	}
}

func CreateMessageButtons() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Emoji: &discordgo.ComponentEmoji{
						Name: "📝",
					},
					Style:    discordgo.SecondaryButton,
					CustomID: EditButton,
				},
				discordgo.Button{
					Emoji: &discordgo.ComponentEmoji{
						Name: "🔙",
					},
					Style:    discordgo.SecondaryButton,
					CustomID: UndoButton,
				},
			},
		},
	}
}
