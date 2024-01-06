package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestUpdateFromShoppingList(t *testing.T) {
	// given
	tests := map[string]struct {
		shoppingList []GroceryItem
		update       string
		expected     []GroceryItem
	}{
		"simple quantity update": {
			shoppingList: []GroceryItem{
				{
					ID:     1,
					Item:   "bacon",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
			update: "[1] bacon\t\t, 3",
			expected: []GroceryItem{
				{
					ID:     1,
					Item:   "bacon",
					Amount: 3,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
		},
		"simple item update": {
			shoppingList: []GroceryItem{
				{
					ID:     1,
					Item:   "bac",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}},
			update: "[1] bacon\t\t, 3",
			expected: []GroceryItem{
				{
					ID:     1,
					Item:   "bacon",
					Amount: 3,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
		},
		"complex update": {
			shoppingList: []GroceryItem{
				{
					ID:     1,
					Item:   "coffee",
					Amount: 2,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:     2,
					Item:   "eggz",
					Amount: 4,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:     3,
					Item:   "milk",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
			update: "[1] bacon\n[2] eggs\t\t,2\n\n[3] milk",
			expected: []GroceryItem{
				{
					ID:     1,
					Item:   "bacon",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:     2,
					Item:   "eggs",
					Amount: 2,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:     3,
					Item:   "milk",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
		},
		"complex update + added items": {
			shoppingList: []GroceryItem{
				{
					ID:     1,
					Item:   "eggos",
					Amount: 4,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:     2,
					Item:   "milk",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
			update: "bacon\n[1] eggs\t\t,2\n[2] milk\nbeer,6",
			expected: []GroceryItem{
				{
					ID:     1,
					Item:   "eggs",
					Amount: 2,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:     2,
					Item:   "milk",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:     3,
					Item:   "bacon",
					Amount: 1,
					Date:   time.Now().Truncate(time.Minute),
				}, {
					ID:     4,
					Item:   "beer",
					Amount: 6,
					Date:   time.Now().Truncate(time.Minute),
				},
			},
		},
		//"remove item": {
		//	shoppingList: []GroceryItem{
		//		{
		//			ID:     1,
		//			Item:   "eggos",
		//			Amount: 4,
		//			Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
		//		}, {
		//			ID:     2,
		//			Item:   "milk",
		//			Amount: 1,
		//			Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
		//		},
		//	},
		//	update: "[1] eggs\t\t,2\n",
		//	expected: []GroceryItem{
		//		{
		//			ID:     1,
		//			Item:   "eggs",
		//			Amount: 2,
		//			Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
		//		},
		//	},
		//},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// when
			actual := UpdateFromShoppingList(test.shoppingList, test.update)

			// then
			assert.EqualValues(t, test.expected, actual)
		})
	}
}

func TestFromShoppingListTable(t *testing.T) {
	// given
	tests := map[string]struct {
		table    string
		expected []GroceryItem
	}{
		"simple conversion": {
			table: "```md\n" +
				"| # | ITEM | QTY | ADDED  |\n" +
				"|---|------|-----|--------|\n" +
				"| 1 | test | 3   | 27.12. |\n" +
				"```",
			expected: []GroceryItem{
				{
					ID:     1,
					Item:   "test",
					Amount: 3,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
		},
		"multi conversion": {
			table: "```md\n" +
				"| # |  ITEM  | QTY | ADDED  |\n" +
				"|---|--------|-----|--------|\n" +
				"| 1 | eggs   | 4   | 24.12. |\n" +
				"| 2 | coffee | 1   | 25.12. |\n" +
				"| 3 | bacon  | 3   | 26.12. |\n" +
				"| 4 | milk   | 1   | 27.12. |\n" +
				"```",
			expected: []GroceryItem{
				{
					ID:     1,
					Item:   "eggs",
					Amount: 4,
					Date:   time.Date(time.Now().Year(), 12, 24, 0, 0, 0, 0, time.Local),
				}, {
					ID:     2,
					Item:   "coffee",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 25, 0, 0, 0, 0, time.Local),
				}, {
					ID:     3,
					Item:   "bacon",
					Amount: 3,
					Date:   time.Date(time.Now().Year(), 12, 26, 0, 0, 0, 0, time.Local),
				}, {
					ID:     4,
					Item:   "milk",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// when
			actual := FromShoppingListTable(test.table)

			// then
			assert.EqualValues(t, test.expected, actual)
		})
	}
}
