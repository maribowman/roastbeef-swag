package service

import (
	"fmt"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestUpdateFromList(t *testing.T) {
	// given
	tests := map[string]struct {
		shoppingList []model.PantryItem
		update       string
		expected     []model.PantryItem
	}{
		"simple quantity update": {
			shoppingList: []model.PantryItem{
				{
					ID:     0,
					Number: 1,
					Item:   "bacon",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
			update: "[1] 3 bacon\n",
			expected: []model.PantryItem{
				{
					ID:     0,
					Number: 1,
					Item:   "bacon",
					Amount: 3,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
		},
		"simple item update": {
			shoppingList: []model.PantryItem{
				{
					ID:     0,
					Number: 1,
					Item:   "bac",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}},
			update: "[1] 3 bacon\n",
			expected: []model.PantryItem{
				{
					ID:     0,
					Number: 1,
					Item:   "bacon",
					Amount: 3,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
		},
		"complex update": {
			shoppingList: []model.PantryItem{
				{
					ID:     0,
					Number: 1,
					Item:   "coffee",
					Amount: 2,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:     0,
					Number: 2,
					Item:   "eggz",
					Amount: 4,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:     0,
					Number: 3,
					Item:   "milk",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
			update: "[1] 1 bacon\n[2] 2 eggs\n\n[3] milk",
			expected: []model.PantryItem{
				{
					ID:     0,
					Number: 1,
					Item:   "bacon",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:     0,
					Number: 2,
					Item:   "eggs",
					Amount: 2,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:     0,
					Number: 3,
					Item:   "milk",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
		},
		"complex update + added items": {
			shoppingList: []model.PantryItem{
				{
					ID:     0,
					Number: 1,
					Item:   "eggos",
					Amount: 4,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:     0,
					Number: 2,
					Item:   "milk",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
			update: "bacon\n[1] 2 eggs\n[2] milk\n6 beer",
			expected: []model.PantryItem{
				{
					ID:     0,
					Number: 1,
					Item:   "eggs",
					Amount: 2,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:     0,
					Number: 2,
					Item:   "milk",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:     0,
					Number: 3,
					Item:   "bacon",
					Amount: 1,
					Date:   time.Now().Truncate(time.Minute),
				}, {
					ID:     0,
					Number: 4,
					Item:   "beer",
					Amount: 6,
					Date:   time.Now().Truncate(time.Minute),
				},
			},
		},
		"remove item": {
			shoppingList: []model.PantryItem{
				{
					ID:     0,
					Number: 1,
					Item:   "eggos",
					Amount: 4,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:     0,
					Number: 2,
					Item:   "milk",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
			update: "[1] 2 eggs\n",
			expected: []model.PantryItem{
				{
					ID:     0,
					Number: 1,
					Item:   "eggs",
					Amount: 2,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// when
			actual := UpdateItemsFromList(test.shoppingList, test.update)

			// then
			assert.EqualValues(t, test.expected, actual)
		})
	}
}

func TestRemove(t *testing.T) {
	// given
	tests := map[string]struct {
		content  string
		expected []model.PantryItem
	}{
		"single remove": {
			content: "7",
			expected: []model.PantryItem{
				{0, 1, "item", 1, time.Now().Truncate(time.Minute)},
				{0, 2, "item", 2, time.Now().Truncate(time.Minute)},
				{0, 3, "item", 3, time.Now().Truncate(time.Minute)},
				{0, 4, "item", 4, time.Now().Truncate(time.Minute)},
				{0, 5, "item", 5, time.Now().Truncate(time.Minute)},
				{0, 6, "item", 6, time.Now().Truncate(time.Minute)},
				{0, 7, "item", 8, time.Now().Truncate(time.Minute)},
				{0, 8, "item", 9, time.Now().Truncate(time.Minute)},
			},
		},
		"multi remove": {
			content: "3 5 8",
			expected: []model.PantryItem{
				{0, 1, "item", 1, time.Now().Truncate(time.Minute)},
				{0, 2, "item", 2, time.Now().Truncate(time.Minute)},
				{0, 3, "item", 4, time.Now().Truncate(time.Minute)},
				{0, 4, "item", 6, time.Now().Truncate(time.Minute)},
				{0, 5, "item", 7, time.Now().Truncate(time.Minute)},
				{0, 6, "item", 9, time.Now().Truncate(time.Minute)},
			},
		},
		"single and range remove": {
			content: "1 4-7",
			expected: []model.PantryItem{
				{0, 1, "item", 2, time.Now().Truncate(time.Minute)},
				{0, 2, "item", 3, time.Now().Truncate(time.Minute)},
				{0, 3, "item", 8, time.Now().Truncate(time.Minute)},
				{0, 4, "item", 9, time.Now().Truncate(time.Minute)},
			},
		},
		"range remove": {
			content: "2-5",
			expected: []model.PantryItem{
				{0, 1, "item", 1, time.Now().Truncate(time.Minute)},
				{0, 2, "item", 6, time.Now().Truncate(time.Minute)},
				{0, 3, "item", 7, time.Now().Truncate(time.Minute)},
				{0, 4, "item", 8, time.Now().Truncate(time.Minute)},
				{0, 5, "item", 9, time.Now().Truncate(time.Minute)},
			},
		},
		"remove all": {
			content:  "*",
			expected: []model.PantryItem{},
		},
		"remove all except single": {
			content: "* 5",
			expected: []model.PantryItem{
				{0, 1, "item", 5, time.Now().Truncate(time.Minute)},
			},
		},
		"remove all except multi": {
			content: "* 5 2 8",
			expected: []model.PantryItem{
				{0, 1, "item", 2, time.Now().Truncate(time.Minute)},
				{0, 2, "item", 5, time.Now().Truncate(time.Minute)},
				{0, 3, "item", 8, time.Now().Truncate(time.Minute)},
			},
		},
		"remove all except range": {
			content: "* 3-6",
			expected: []model.PantryItem{
				{0, 1, "item", 3, time.Now().Truncate(time.Minute)},
				{0, 2, "item", 4, time.Now().Truncate(time.Minute)},
				{0, 3, "item", 5, time.Now().Truncate(time.Minute)},
				{0, 4, "item", 6, time.Now().Truncate(time.Minute)},
			},
		},
		"remove all except single and range": {
			content: "* 7 1-3",
			expected: []model.PantryItem{
				{0, 1, "item", 1, time.Now().Truncate(time.Minute)},
				{0, 2, "item", 2, time.Now().Truncate(time.Minute)},
				{0, 3, "item", 3, time.Now().Truncate(time.Minute)},
				{0, 4, "item", 7, time.Now().Truncate(time.Minute)},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// and
			var items []model.PantryItem
			for i := 1; i < 10; i++ {
				items = add(items, fmt.Sprintf("item %d", i), time.Now().Truncate(time.Minute))
			}

			// when
			actual := remove(items, test.content)

			// then
			assert.EqualValues(t, test.expected, actual)
		})
	}
}

func TestAdd(t *testing.T) {
	// given
	tests := map[string]struct {
		content  string
		expected []model.PantryItem
	}{
		"simple add": {
			content:  "bacon",
			expected: []model.PantryItem{{0, 1, "bacon", 1, time.Now().Truncate(time.Minute)}},
		},
		"simple multi word add": {
			content:  "butter scotch",
			expected: []model.PantryItem{{0, 1, "butter scotch", 1, time.Now().Truncate(time.Minute)}},
		},
		"simple hyphened add": {
			content:  "dry-gin",
			expected: []model.PantryItem{{0, 1, "dry-gin", 1, time.Now().Truncate(time.Minute)}},
		},
		"add with trailing quantity": {
			content:  "bacon 5",
			expected: []model.PantryItem{{0, 1, "bacon", 5, time.Now().Truncate(time.Minute)}},
		},
		"add with leading quantity": {
			content:  "13 bacon",
			expected: []model.PantryItem{{0, 1, "bacon", 13, time.Now().Truncate(time.Minute)}},
		},
		"add with numbered name": {
			content:  "2 monkey47",
			expected: []model.PantryItem{{0, 1, "monkey47", 2, time.Now().Truncate(time.Minute)}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// when
			actual := add([]model.PantryItem{}, test.content, time.Now().Truncate(time.Minute))

			// then
			assert.EqualValues(t, test.expected, actual)
		})
	}
}
