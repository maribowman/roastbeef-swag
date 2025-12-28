package service

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/config"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/rs/zerolog/log"
)

const (
	GroceriesChannel = "groceries"
	FreezerChannel   = "freezer"

	EditButton     = "edit-button"
	UndoButton     = "undo-button"
	EditModal      = "edit-modal"
	EditModalInput = "edit-modal-input"
)

var (
	// Prefix for modal items
	modalIndexPrefixRegex = regexp.MustCompile(`^\[(\d+)]\s`)
	// Remove items with *, a list (e.g. 4 5 6) or a range (e.g. 4-6) of numbers
	removeRegex = regexp.MustCompile(`^(\*)?(?:\s*\d+)*\s*(\d+-\d+)?$`)
	// Allow to specify quantities at the beginning or end of a message
	leadingQuantity  = regexp.MustCompile(`^(\d+)\s.*`)
	trailingQuantity = regexp.MustCompile(`\s(\d+)$`)
)

// PreProcessMessageEvent consumes and preprocesses all channel messages.
// It returns the message ID for the original bot Markdown table, the accumulated user input,
// a list of message IDs (which can be dropped) and a potential error.
func PreProcessMessageEvent(session *discordgo.Session, channelID string) (string, string, []string, error) {
	var lastBotMessageID string
	var userInput string
	var removableMessageIDs []string

	channelMessages, err := session.ChannelMessages(channelID, 100, "", "", "")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get messages from Discord channel")
		return "", "", nil, err
	}

	var lastBotMessage *discordgo.Message
	for _, msg := range channelMessages {
		if msg.Author.ID == config.Config.Discord.BotID {
			if lastBotMessage == nil {
				lastBotMessage = msg
				lastBotMessageID = msg.ID
				continue
			} else if lastBotMessage.Timestamp.After(msg.Timestamp) {
				removableMessageIDs = append(removableMessageIDs, lastBotMessageID) // Add previous bot message to remove list
				lastBotMessage = msg
				lastBotMessageID = msg.ID
				continue
			}
		} else {
			userInput += "\n" + msg.Content
		}
		removableMessageIDs = append(removableMessageIDs, msg.ID)
	}
	return lastBotMessageID, userInput, removableMessageIDs, err
}

func UpdateItemsFromModal(pantryClient model.PantryClient, modalInput string) {
	items := pantryClient.GetItems()
	updatedItems := make([]int, len(items))

	for line := range strings.Lines(modalInput) {
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		matches := modalIndexPrefixRegex.FindStringSubmatch(line)
		line = strings.TrimSpace(modalIndexPrefixRegex.ReplaceAllString(line, ""))
		modalItem := generateNewPantryItem(line)

		if len(matches) == 2 { // Regex matches full string + capture group
			// Update pantry item with name and amount from modal input
			index, _ := strconv.Atoi(matches[1])
			item := items[index-1]
			updatedItem := model.PantryItem{
				ID:     item.ID,
				Name:   modalItem.Name,
				Amount: modalItem.Amount,
				Date:   item.Date,
			}

			pantryClient.UpdateItem(updatedItem)
			updatedItems = append(updatedItems, index)
		} else {
			// Add new item at the end if there's no number
			pantryClient.AddItem(modalItem)
		}
	}

	// Remove missing (non-updated) IDs
	for index, item := range items {
		if !slices.Contains(updatedItems, index+1) {
			pantryClient.RemoveItem(item.ID)
		}
	}
}

func UpdateItems(pantryClient model.PantryClient, userInput string) {
	for line := range strings.Lines(userInput) {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		if removeRegex.MatchString(line) {
			removableIndices := determineRemovableIndices(line)
			for index, item := range pantryClient.GetItems() {
				if slices.Contains(removableIndices, index+1) {
					pantryClient.RemoveItem(item.ID)
				}
			}

		} else {
			pantryClient.AddItem(generateNewPantryItem(line))
		}
	}
}

// determineRemovableIndices returns a list of indices which can be removed
func determineRemovableIndices(line string) []int {
	var result []int
	// removeAllExcept := false

	// CAPTURE GROUP 0: entire string
	// CAPTURE GROUP 1: asterisk
	// CAPTURE GROUP 2: range
	captureGroups := removeRegex.FindStringSubmatch(line)

	// Remove all (except)
	if captureGroups[1] == "*" {
		// removeAllExcept = true
		if captureGroups[0] == captureGroups[1] {
			return result
		}
	}

	// Add single removable numbers
	var indices []int
	if captureGroups[0] != captureGroups[2] {
		for value := range strings.SplitSeq(captureGroups[0], " ") {
			if index, err := strconv.Atoi(value); err == nil {
				indices = append(indices, index)
			}
		}
	}

	// Add range to removable numbers
	if captureGroups[2] != "" {
		range_ := strings.Split(captureGroups[2], "-")
		rangeStart, _ := strconv.Atoi(range_[0])
		rangeEnd, _ := strconv.Atoi(range_[1])

		for i := rangeStart; i <= rangeEnd; i++ {
			indices = append(indices, i)
		}
	}

	//	for _, entry := range items {
	//		if slices.Contains(numbers, entry.Number) {
	//			if !removeAllExcept {
	//				continue
	//			}
	//		} else if removeAllExcept {
	//			continue
	//		}
	//		entry.Number = len(result) + 1
	//		result = append(result, entry)
	//	}
	return result
}

// generateNewPantryItem creates a new PantryItem
func generateNewPantryItem(line string) model.PantryItem {
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

	return model.PantryItem{
		Name:   strings.TrimSpace(line),
		Amount: amount,
		Date:   time.Now().Truncate(time.Minute),
	}
}

// PublishItems sends the latest []PantryItem state to the active channel (via new or edited message).
// Because of a character limit of 2000, the function automatically splits the Markdown table line by line
// and sends multiple messages to the channel. The last message always contains the buttons to interact with the bot.
func PublishItems(items []model.PantryItem, session *discordgo.Session, channelID, messageID string, lineBreak int, dateFormat string) {
	markdownTable := model.ToMarkdownTable(items, lineBreak, dateFormat)

	if len(markdownTable) <= 2000 { // 2000 is the message length limit
		if messageID != "" { // Update existing message
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
	} else { // Split table line by line
		tempTable := ""

		for line := range strings.Lines(markdownTable) {
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

// CreateMessageButtons generates the necessary buttons for the bots Markdown table message
func CreateMessageButtons() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Emoji: &discordgo.ComponentEmoji{
						Name: "📝",
					},
					CustomID: EditButton,
					Style:    discordgo.SecondaryButton,
					Label:    "Edit",
				},
				discordgo.Button{
					Emoji: &discordgo.ComponentEmoji{
						Name: "🔙",
					},
					CustomID: UndoButton,
					Style:    discordgo.SecondaryButton,
					Label:    "Undo",
				},
			},
		},
	}
}
