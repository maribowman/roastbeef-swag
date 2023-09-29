package service

import (
	"github.com/maribowman/roastbeef-swag/app/model"
)

type DiscordBot struct {
	groceryClient model.GroceryClient
}

type Wiring struct {
	GroceryClient model.GroceryClient
}

func NewScreenCastService(wiring *Wiring) model.DiscordBot {
	return &DiscordBot{
		groceryClient: wiring.GroceryClient,
	}
}
