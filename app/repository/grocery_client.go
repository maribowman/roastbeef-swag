package repository

import "github.com/maribowman/roastbeef-swag/app/model"

type GroceryClient struct {
}

func NewGroceryClient() model.GroceryClient {
	return &GroceryClient{}
}
