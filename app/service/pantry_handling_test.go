package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/maribowman/roastbeef-swag/app/repository"
	"github.com/stretchr/testify/assert"
)

func TestUpdateItemsFromModal(t *testing.T) {
	// where
	tests := map[string]struct {
		pantryItems []model.PantryItem
		modalInput  string
		expected    []model.PantryItem
	}{
		"simple quantity update": {
			pantryItems: []model.PantryItem{
				{Name: "bacon", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
			},
			modalInput: "[1] 3 bacon",
			expected: []model.PantryItem{
				{ID: 1, Name: "bacon", Quantity: 3, Date: time.Now().Truncate(time.Minute)},
			},
		},
		"simple item update": {
			pantryItems: []model.PantryItem{
				{Name: "BAC", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
			},
			modalInput: "[1] 3 bacon",
			expected: []model.PantryItem{
				{ID: 1, Name: "bacon", Quantity: 3, Date: time.Now().Truncate(time.Minute)},
			},
		},
		"complex update": {
			pantryItems: []model.PantryItem{
				{Name: "coffee", Quantity: 2, Date: time.Now().Truncate(time.Minute)},
				{Name: "eggz", Quantity: 4, Date: time.Now().Truncate(time.Minute)},
				{Name: "milk", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
			},
			modalInput: "[1] 1 bacon\n[2] 2 eggs\n\n[3] milk",
			expected: []model.PantryItem{
				{ID: 1, Name: "bacon", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
				{ID: 2, Name: "eggs", Quantity: 2, Date: time.Now().Truncate(time.Minute)},
				{ID: 3, Name: "milk", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
			},
		},
		"complex update + added items": {
			pantryItems: []model.PantryItem{
				{Name: "eggos", Quantity: 4, Date: time.Now().Truncate(time.Minute)},
				{Name: "milk", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
			},
			modalInput: "bacon\n[1] 2 eggs\n[2] milk\n6 beer",
			expected: []model.PantryItem{
				{ID: 1, Name: "eggs", Quantity: 2, Date: time.Now().Truncate(time.Minute)},
				{ID: 2, Name: "milk", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
				{ID: 3, Name: "bacon", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
				{ID: 4, Name: "beer", Quantity: 6, Date: time.Now().Truncate(time.Minute)},
			},
		},
		"remove item": {
			pantryItems: []model.PantryItem{
				{Name: "eggos", Quantity: 4, Date: time.Now().Truncate(time.Minute)},
				{Name: "milk", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
			},
			modalInput: "[1] 2 eggs\n",
			expected: []model.PantryItem{
				{ID: 1, Name: "eggs", Quantity: 2, Date: time.Now().Truncate(time.Minute)},
			},
		},
	}

	// given
	pantryClient := repository.NewSqlitePantryClient(repository.NewDatabaseClient(), "unit_tests")

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// and
			for _, item := range test.pantryItems {
				pantryClient.AddItem(item)
			}

			// when
			UpdateItemsFromModal(pantryClient, test.modalInput)

			// then
			actual := pantryClient.GetItems()
			assert.EqualValues(t, test.expected, actual)

			// cleanup
			pantryClient.RemoveAllItems()
		})
	}
}

func TestUpdateItems(t *testing.T) {
	// where
	tests := map[string]struct {
		pantryItemCount int
		input           string
		expected        []model.PantryItem
	}{
		// EDIT TEST CASES
		"simple quantity update": {
			pantryItemCount: 1,
			input:           "1++",
			expected: []model.PantryItem{
				{ID: 1, Name: "Item #1", Quantity: 2, Date: time.Now().Truncate(time.Minute)},
			},
		},
		"advanced quantity update": {
			pantryItemCount: 2,
			input:           "2--3",
			expected: []model.PantryItem{
				{ID: 1, Name: "Item #1", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
				{ID: 2, Name: "Item #2", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
			},
		},
		"advanced negative quantity update exception": {
			pantryItemCount: 1,
			input:           "1--5",
			expected:        nil,
		},
		"quantity update to zero removes item": {
			pantryItemCount: 1,
			input:           "1--1",
			expected:        nil,
		},

		// REMOVE TEST CASES
		"single number remove": {
			pantryItemCount: 3,
			input:           "2",
			expected: []model.PantryItem{
				{ID: 1, Name: "Item #1", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
				{ID: 3, Name: "Item #3", Quantity: 9, Date: time.Now().Truncate(time.Minute)},
			},
		},
		"multi number remove": {
			pantryItemCount: 5,
			input:           "2 4 5",
			expected: []model.PantryItem{
				{ID: 1, Name: "Item #1", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
				{ID: 3, Name: "Item #3", Quantity: 9, Date: time.Now().Truncate(time.Minute)},
			},
		},
		"single range remove": {
			pantryItemCount: 5,
			input:           "2-5",
			expected: []model.PantryItem{
				{ID: 1, Name: "Item #1", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
			},
		},
		"multi range remove": {
			pantryItemCount: 10,
			input:           "1-3 5-9",
			expected: []model.PantryItem{
				{ID: 4, Name: "Item #4", Quantity: 16, Date: time.Now().Truncate(time.Minute)},
				{ID: 10, Name: "Item #10", Quantity: 100, Date: time.Now().Truncate(time.Minute)},
			},
		},
		"single number and single range remove": {
			pantryItemCount: 5,
			input:           "1 3-5",
			expected: []model.PantryItem{
				{ID: 2, Name: "Item #2", Quantity: 4, Date: time.Now().Truncate(time.Minute)},
			},
		},
		"multi number and multi range remove": {
			pantryItemCount: 15,
			input:           "1 3 5-10 12-15",
			expected: []model.PantryItem{
				{ID: 2, Name: "Item #2", Quantity: 4, Date: time.Now().Truncate(time.Minute)},
				{ID: 4, Name: "Item #4", Quantity: 16, Date: time.Now().Truncate(time.Minute)},
				{ID: 11, Name: "Item #11", Quantity: 121, Date: time.Now().Truncate(time.Minute)},
			},
		},
		"remove all": {
			pantryItemCount: 5,
			input:           "*",
			expected:        nil,
		},
		"remove all except single number": {
			pantryItemCount: 1,
			input:           "* 1",
			expected: []model.PantryItem{
				{ID: 1, Name: "Item #1", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
			},
		},
		"remove all except multi number": {
			pantryItemCount: 5,
			input:           "* 2 4",
			expected: []model.PantryItem{
				{ID: 2, Name: "Item #2", Quantity: 4, Date: time.Now().Truncate(time.Minute)},
				{ID: 4, Name: "Item #4", Quantity: 16, Date: time.Now().Truncate(time.Minute)},
			},
		},
		"remove all except single range": {
			pantryItemCount: 5,
			input:           "* 1-3",
			expected: []model.PantryItem{
				{ID: 1, Name: "Item #1", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
				{ID: 2, Name: "Item #2", Quantity: 4, Date: time.Now().Truncate(time.Minute)},
				{ID: 3, Name: "Item #3", Quantity: 9, Date: time.Now().Truncate(time.Minute)},
			},
		},
		"remove all except multi range": {
			pantryItemCount: 10,
			input:           "* 1-3 5-6",
			expected: []model.PantryItem{
				{ID: 1, Name: "Item #1", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
				{ID: 2, Name: "Item #2", Quantity: 4, Date: time.Now().Truncate(time.Minute)},
				{ID: 3, Name: "Item #3", Quantity: 9, Date: time.Now().Truncate(time.Minute)},
				{ID: 5, Name: "Item #5", Quantity: 25, Date: time.Now().Truncate(time.Minute)},
				{ID: 6, Name: "Item #6", Quantity: 36, Date: time.Now().Truncate(time.Minute)},
			},
		},
		"remove all except single number and single range": {
			pantryItemCount: 5,
			input:           "* 5 1-3",
			expected: []model.PantryItem{
				{ID: 1, Name: "Item #1", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
				{ID: 2, Name: "Item #2", Quantity: 4, Date: time.Now().Truncate(time.Minute)},
				{ID: 3, Name: "Item #3", Quantity: 9, Date: time.Now().Truncate(time.Minute)},
				{ID: 5, Name: "Item #5", Quantity: 25, Date: time.Now().Truncate(time.Minute)},
			},
		},
		"remove all except multi number and multi range": {
			pantryItemCount: 10,
			input:           "* 1 6 3-5 7-8",
			expected: []model.PantryItem{
				{ID: 1, Name: "Item #1", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
				{ID: 3, Name: "Item #3", Quantity: 9, Date: time.Now().Truncate(time.Minute)},
				{ID: 4, Name: "Item #4", Quantity: 16, Date: time.Now().Truncate(time.Minute)},
				{ID: 5, Name: "Item #5", Quantity: 25, Date: time.Now().Truncate(time.Minute)},
				{ID: 6, Name: "Item #6", Quantity: 36, Date: time.Now().Truncate(time.Minute)},
				{ID: 7, Name: "Item #7", Quantity: 49, Date: time.Now().Truncate(time.Minute)},
				{ID: 8, Name: "Item #8", Quantity: 64, Date: time.Now().Truncate(time.Minute)},
			},
		},

		// ADD TEST CASES
		"simple add": {
			pantryItemCount: 0,
			input:           "bacon",
			expected: []model.PantryItem{
				{ID: 1, Name: "bacon", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
			},
		},
		"simple multi word add": {
			pantryItemCount: 0,
			input:           "butter scotch",
			expected: []model.PantryItem{
				{ID: 1, Name: "butter scotch", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
			},
		},
		"simple hyphened add": {
			pantryItemCount: 0,
			input:           "dry-gin",
			expected: []model.PantryItem{
				{ID: 1, Name: "dry-gin", Quantity: 1, Date: time.Now().Truncate(time.Minute)}},
		},
		"add with trailing quantity": {
			pantryItemCount: 0,
			input:           "bacon 5",
			expected: []model.PantryItem{
				{ID: 1, Name: "bacon", Quantity: 5, Date: time.Now().Truncate(time.Minute)}},
		},
		"add with leading quantity": {
			pantryItemCount: 0,
			input:           "13 bacon",
			expected: []model.PantryItem{
				{ID: 1, Name: "bacon", Quantity: 13, Date: time.Now().Truncate(time.Minute)}},
		},
		"add with numbered name": {
			pantryItemCount: 0,
			input:           "2 monkey47",
			expected: []model.PantryItem{
				{ID: 1, Name: "monkey47", Quantity: 2, Date: time.Now().Truncate(time.Minute)}},
		},
		"multi line add": {
			pantryItemCount: 0,
			input:           "3 bacon\ncoffee 4",
			expected: []model.PantryItem{
				{ID: 1, Name: "bacon", Quantity: 3, Date: time.Now().Truncate(time.Minute)},
				{ID: 2, Name: "coffee", Quantity: 4, Date: time.Now().Truncate(time.Minute)},
			},
		},
	}

	// given
	pantryClient := repository.NewSqlitePantryClient(repository.NewDatabaseClient(), "unit_tests")

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// and
			for i := 1; i <= test.pantryItemCount; i++ {
				pantryClient.AddItem(model.PantryItem{
					Name:     fmt.Sprintf("Item #%d", i),
					Quantity: i * i,
					Date:     time.Now().Truncate(time.Minute),
				})
			}

			// when
			UpdateItems(pantryClient, test.input)

			// then
			actual := pantryClient.GetItems()
			assert.EqualValues(t, test.expected, actual)

			// cleanup
			pantryClient.RemoveAllItems()
		})
	}
}
