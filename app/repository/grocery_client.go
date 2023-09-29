package repository

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/model"
)

type GroceryClient struct {
}

func NewGroceryClient(session *discordgo.Session) model.GroceryClient {
	session.AddHandler(messageCreate)
	return &GroceryClient{}
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == session.State.User.ID {
		return
	}

	if message.Content == "!gophers" {
		if true {
			_, err := session.ChannelMessageSend(message.ChannelID, "")
			//_, err = session.ChannelFileSend(message.ChannelID, "random-gopher.png", response.Body)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("Error: Can't get list of Gophers! :-(")
		}
	}
}
