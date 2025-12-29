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
	// Routing Regexes
	editRegex   = regexp.MustCompile(`^(\d+)(\+\+|--)(\d+)?$`)
	removeRegex = regexp.MustCompile(`^(\*|(\*\s)?[\d\s\-]+)$`)
	// Detects digits and ranges for removal
	indexIdentifierRegex = regexp.MustCompile(`(\d+)(?:-(\d+))?`)
	// Allow quantity specification at the beginning or end
	quantityIdentifierRegex = regexp.MustCompile(`^(?:(\d+)\s+)?(.*?)(?:\s+(\d+))?$`)
	leadingQuantity         = regexp.MustCompile(`^(\d+)\s.*`)
	trailingQuantity        = regexp.MustCompile(`\s(\d+)$`)
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
				ID:       item.ID,
				Name:     modalItem.Name,
				Quantity: modalItem.Quantity,
				Date:     item.Date,
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

		if editRegex.MatchString(line) {
			// Edit quantity of pantry item(s)
			index, quantityDelta := determineQuantityDelta(line)
			for idx, item := range pantryClient.GetItems() {
				if idx+1 == index {
					updatedQuantity := item.Quantity + quantityDelta
					if updatedQuantity <= 0 {
						pantryClient.RemoveItem(item.ID)
					}
					pantryClient.UpdateItem(model.PantryItem{
						ID:       item.ID,
						Name:     item.Name,
						Quantity: updatedQuantity,
						Date:     item.Date,
					})
				}
			}

		} else if removeRegex.MatchString(line) {
			// Remove pantry item(s)
			isRemoveAll := false
			if strings.HasPrefix(line, "*") {
				isRemoveAll = true
				line = strings.TrimSpace(strings.TrimPrefix(line, "*"))
			}

			determinedIndices := determineIndices(line)
			if isRemoveAll && len(determinedIndices) == 0 {
				pantryClient.RemoveAllItems()
			}

			for index, item := range pantryClient.GetItems() {
				if isRemoveAll {
					if slices.Contains(determinedIndices, index+1) {
						continue
					}
					pantryClient.RemoveItem(item.ID)
				} else if slices.Contains(determinedIndices, index+1) {
					pantryClient.RemoveItem(item.ID)
				}
			}

		} else {
			// Add new pantry item(s)
			pantryClient.AddItem(generateNewPantryItem(line))
		}
	}
}

func determineQuantityDelta(input string) (int, int) {
	// match[0] entire string (e.g. `1--` or `1--4`)
	// match[1] pantry item index (e.g. `1`)
	// match[2] operator (e.g. `++` or `--`)
	// match[3] delta (e.g. `4`)
	matches := editRegex.FindStringSubmatch(input)

	index, _ := strconv.Atoi(matches[1])
	var operator int
	switch matches[2] {
	case "++":
		operator = 1
	case "--":
		operator = -1
	}

	delta := 1
	if matches[3] != "" {
		delta, _ = strconv.Atoi(matches[3])
	}

	return index, delta * operator
}

// determineIndices resolves all specified indices
func determineIndices(input string) []int {
	var result []int

	// match[0] entire string (e.g. `3` or `3-6`)
	// match[1] range start or single digit (e.g. `3`)
	// match[2] range end (e.g. `6`)
	matches := indexIdentifierRegex.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		start, _ := strconv.Atoi(match[1])

		if match[2] != "" {
			// Range detected
			end, _ := strconv.Atoi(match[2])
			if start <= end {
				for i := start; i <= end; i++ {
					result = append(result, i)
				}
			}
		} else {
			// Single digit
			result = append(result, start)
		}
	}

	return result
}

// generateNewPantryItem creates a new PantryItem
func generateNewPantryItem(input string) model.PantryItem {
	leading := leadingQuantity.FindStringSubmatch(input)
	trailing := trailingQuantity.FindStringSubmatch(input)

	var quantity string
	if leading != nil {
		quantity = leading[1]
		input = strings.TrimPrefix(input, quantity)
	} else if trailing != nil {
		quantity = trailing[1]
		input = strings.TrimSuffix(input, quantity)
	}

	amount, err := strconv.Atoi(quantity)
	if err != nil {
		amount = 1
	}

	return model.PantryItem{
		Name:     strings.TrimSpace(input),
		Quantity: amount,
		Date:     time.Now().Truncate(time.Minute),
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
