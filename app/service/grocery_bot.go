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

type GroceryBot struct {
	botID                     string
	channelID                 string
	shoppingList              []model.ShoppingListItem
	previousShoppingListTable string
}

func NewGroceryBot(botID string, channelID string) model.DiscordBot {
	log.Debug().Msg("Registering grocery client handler")
	return &GroceryBot{
		botID:     botID,
		channelID: channelID,
	}
}

func (bot *GroceryBot) ReadyEvent(session *discordgo.Session, ready *discordgo.Ready) {
	bot.MessageEvent(session, &discordgo.MessageCreate{Message: &discordgo.Message{Author: &discordgo.User{ID: "init"}}})
}

func (bot *GroceryBot) MessageEvent(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == bot.botID {
		return
	}

	channelMessages, err := session.ChannelMessages(bot.channelID, 100, "", "", "")
	if err != nil {
		return
	}

	var lastMessage *discordgo.Message
	var content string
	var removableMessageIDs []string

	for _, msg := range channelMessages {
		if msg.Author.ID == bot.botID {
			if lastMessage == nil {
				lastMessage = msg
				bot.shoppingList = model.FromShoppingListTable(msg.Content)
				continue
			} else if lastMessage.Timestamp.After(msg.Timestamp) {
				removableMessageIDs = append(removableMessageIDs, lastMessage.ID)
				lastMessage = msg
				bot.shoppingList = model.FromShoppingListTable(msg.Content)
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
			bot.remove(line)
		} else {
			bot.add(line)
		}
	}

	if err := session.ChannelMessagesBulkDelete(bot.channelID, removableMessageIDs); err != nil {
		log.Error().Err(err).Msg("could not bulk delete channel messages")
	}

	bot.publish(session, lastMessage)
}

func (bot *GroceryBot) MessageComponentInteractionEvent(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
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
								Value:    model.ToShoppingList(bot.shoppingList),
							},
						},
					},
				},
			},
		}
	case DoneButton:
		bot.shoppingList = []model.ShoppingListItem{}
		response = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    model.ToShoppingListTable(bot.shoppingList),
				Components: createMessageButtons(),
			},
		}
	default:
		log.Error().Msgf("Could not map message component interaction event `%s`", interaction.MessageComponentData().CustomID)
	}

	_ = session.InteractionRespond(interaction.Interaction, response)
}

func (bot *GroceryBot) ModalSubmitInteractionEvent(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	switch interaction.ModalSubmitData().CustomID {
	case EditModal:
		// TODO use modal input
		//interaction.ModalSubmitData().Components[0]
	default:
		log.Error().Msgf("Could not map modal-submit interaction event `%s`", interaction.ModalSubmitData().CustomID)
	}
}

func (bot *GroceryBot) remove(line string) {
	var tempShoppingList []model.ShoppingListItem
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

	bot.shoppingList = append(bot.shoppingList, model.ShoppingListItem{
		ID:     len(bot.shoppingList) + 1,
		Item:   strings.TrimSpace(line),
		Amount: amount,
		Date:   time.Now().Truncate(time.Minute),
	})
}

func (bot *GroceryBot) publish(session *discordgo.Session,
	lastMessage *discordgo.Message) {
	if lastMessage != nil {
		editedMessage := discordgo.NewMessageEdit(bot.channelID, lastMessage.ID)
		editedMessage.SetContent(model.ToShoppingListTable(bot.shoppingList))
		if _, err := session.ChannelMessageEditComplex(editedMessage); err != nil {
			log.Error().Err(err).Msgf("Could not edit message %s", lastMessage.ID)
		}
		return
	}

	if _, err := session.ChannelMessageSendComplex(bot.channelID, &discordgo.MessageSend{
		Content:    model.ToShoppingListTable(bot.shoppingList),
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
