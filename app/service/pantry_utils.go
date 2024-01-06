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
	TkGoodsChannelName   = "tkGoods"

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

func Remove(items []model.PantryItem, line string) []model.PantryItem {
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

func Add(items []model.PantryItem, line string) []model.PantryItem {
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

func PublishList(items []model.PantryItem, session *discordgo.Session, channelID, messageID string) {
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
