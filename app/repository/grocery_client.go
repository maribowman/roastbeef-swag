package repository

import (
	"bytes"
	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/config"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog/log"
	"regexp"
	"strconv"
	"time"
)

const groceries_channel_name = "groceries"

type GroceryClient struct {
	groceryChannelID string
	shoppingList     []model.ShoppingEntry
}

func NewGroceryClient(session *discordgo.Session) model.GroceryClient {
	log.Debug().Msg("registering grocery client handler")
	client := GroceryClient{}
	for _, channel := range config.Config.Discord.Channels {
		if channel.Name == groceries_channel_name {
			client.groceryChannelID = channel.ID
			break
		}
	}
	session.AddHandler(client.groceryAction)
	return &client
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func (client *GroceryClient) groceryAction(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == session.State.User.ID || message.ChannelID != client.groceryChannelID {
		// TODO remove
		log.Info().Msgf("CHANNEL ID: %s", message.ChannelID, message)
		return
	}
	content := message.Content
	if regexp.MustCompile(`\d`).MatchString(content) {
		number, _ := strconv.Atoi(message.Content)
		for index, entry := range client.shoppingList {
			if entry.ID == number {
				client.shoppingList = append(client.shoppingList[:index], client.shoppingList[index+1:]...)
			}
		}
	} else {
		client.shoppingList = append(client.shoppingList, model.ShoppingEntry{
			ID:     len(client.shoppingList),
			Item:   message.Content,
			Amount: 1,
			Date:   time.Now(),
		})
	}

	shoppingListTable := createShoppingListTable(client.shoppingList)
	if _, err := session.ChannelMessageSend(message.ChannelID, shoppingListTable); err != nil {
		log.Error().Err(err).Msg("could not send message")
	}
}

func createShoppingListTable(shoppingList []model.ShoppingEntry) string {
	var data [][]string
	for _, entry := range shoppingList {
		data = append(data, []string{
			strconv.Itoa(entry.ID),
			entry.Item,
			strconv.Itoa(entry.Amount),
			entry.Date.Format("02.02.")},
		)
	}

	writer := bytes.Buffer{}
	table := tablewriter.NewWriter(&writer)
	table.SetHeader([]string{"ID", "Item", "Amount", "Added"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data)
	table.Render()

	return writer.String()
}
