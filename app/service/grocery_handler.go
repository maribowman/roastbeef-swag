package service

import (
	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/rs/zerolog/log"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	GroceriesChannelName = "groceries"

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

type GroceryHandler struct {
	botID            string
	channelID        string
	shoppingList     []model.GroceryItem
	lastShoppingList string
}

func NewGroceryHandler(botID string, channelID string) model.BotHandler {
	log.Debug().Msg("Registering grocery handler")
	return &GroceryHandler{
		botID:     botID,
		channelID: channelID,
	}
}

func (handler *GroceryHandler) ReadyEvent(session *discordgo.Session, ready *discordgo.Ready) {
	handler.MessageEvent(session, &discordgo.MessageCreate{Message: &discordgo.Message{Author: &discordgo.User{ID: "init"}}})
}

func (handler *GroceryHandler) MessageEvent(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == handler.botID {
		return
	}

	channelMessages, err := session.ChannelMessages(handler.channelID, 100, "", "", "")
	if err != nil {
		return
	}

	var lastBotMessage *discordgo.Message
	var content string
	var removableMessageIDs []string

	for _, msg := range channelMessages {
		if msg.Author.ID == handler.botID {
			if lastBotMessage == nil {
				lastBotMessage = msg
				handler.shoppingList = model.FromMarkdownTable(msg.Content)
				continue
			} else if lastBotMessage.Timestamp.After(msg.Timestamp) {
				removableMessageIDs = append(removableMessageIDs, lastBotMessage.ID)
				lastBotMessage = msg
				handler.shoppingList = model.FromMarkdownTable(msg.Content)
				continue
			}
		} else {
			content += "\n" + msg.Content
		}
		removableMessageIDs = append(removableMessageIDs, msg.ID)
	}

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if removeRegex.MatchString(line) {
			handler.remove(line)
		} else {
			handler.add(line)
		}
	}

	if err := session.ChannelMessagesBulkDelete(handler.channelID, removableMessageIDs); err != nil {
		log.Error().Err(err).Msg("could not bulk delete channel messages")
	}

	handler.publish(session, lastBotMessage.ChannelID, lastBotMessage.ID)
}

func (handler *GroceryHandler) MessageComponentInteractionEvent(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var response *discordgo.InteractionResponse

	switch interaction.MessageComponentData().CustomID {
	case EditButton:
		response = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: EditModal,
				Title:    "Edit grocery list",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID: EditModalInput,
								Style:    discordgo.TextInputParagraph,
								Value:    model.ToList(handler.shoppingList),
							},
						},
					},
				},
			},
		}
	case DoneButton:
		handler.shoppingList = []model.GroceryItem{}
		response = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    model.ToMarkdownTable(handler.shoppingList, ""),
				Components: createMessageButtons(),
			},
		}
	default:
		log.Error().Msgf("Could not map message component interaction event `%s`", interaction.MessageComponentData().CustomID)
	}

	_ = session.InteractionRespond(interaction.Interaction, response)
}

func (handler *GroceryHandler) ModalSubmitInteractionEvent(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var response *discordgo.InteractionResponse

	switch interaction.ModalSubmitData().CustomID {
	case EditModal:
		handler.shoppingList = model.UpdateFromList(
			handler.shoppingList,
			interaction.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
		)
		response = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    model.ToMarkdownTable(handler.shoppingList, ""),
				Components: createMessageButtons(),
			},
		}
	default:
		log.Error().Msgf("Could not map modal-submit interaction event `%s`", interaction.ModalSubmitData().CustomID)
	}

	_ = session.InteractionRespond(interaction.Interaction, response)
}

func (handler *GroceryHandler) remove(line string) {
	var tempShoppingList []model.GroceryItem
	removeAllExcept := false

	// CAPTURE GROUP 0: entire string
	// CAPTURE GROUP 1: asterisk
	// CAPTURE GROUP 2: range
	captureGroups := removeRegex.FindStringSubmatch(line)

	// remove all (except)
	if captureGroups[1] == "*" {
		removeAllExcept = true
		if captureGroups[0] == captureGroups[1] {
			handler.shoppingList = tempShoppingList
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

	for _, entry := range handler.shoppingList {
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
	handler.shoppingList = tempShoppingList
}

func (handler *GroceryHandler) add(line string) {
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

	handler.shoppingList = append(handler.shoppingList, model.GroceryItem{
		ID:     len(handler.shoppingList) + 1,
		Item:   strings.TrimSpace(line),
		Amount: amount,
		Date:   time.Now().Truncate(time.Minute),
	})
}

func (handler *GroceryHandler) publish(session *discordgo.Session, channelID, messageID string) {
	var items []model.GroceryItem
	//switch channelID {
	//case handler.groceryChannelID:
	//	items = handler.shoppingList
	//case handler.tkChannelID:
	//	items = handler.tkInventory
	//}

	if messageID != "" {
		editedMessage := discordgo.NewMessageEdit(channelID, messageID)
		editedMessage.SetContent(model.ToMarkdownTable(items, ""))
		if _, err := session.ChannelMessageEditComplex(editedMessage); err != nil {
			log.Error().Err(err).Msgf("Could not edit message %s", messageID)
		}
		return
	}

	if _, err := session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content:    model.ToMarkdownTable(items, ""),
		Components: createMessageButtons(),
	}); err != nil {
		log.Error().Err(err).Msg("Could not send complex message")
	}
}

func createMessageButtons() []discordgo.MessageComponent {
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
