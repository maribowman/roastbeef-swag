package service

import (
	"testing"
	"time"

	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/maribowman/roastbeef-swag/app/repository"
	"github.com/stretchr/testify/assert"
)

func TestUpdateFromModal(t *testing.T) {
	// given
	tests := map[string]struct {
		pantryItems []model.PantryItem
		modalInput  string
		expected    []model.PantryItem
	}{
		"simple quantity update": {
			pantryItems: []model.PantryItem{
				{
					Name:   "bacon",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
			modalInput: "[1] 3 bacon",
			expected: []model.PantryItem{
				{
					ID:     1,
					Name:   "bacon",
					Amount: 3,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
		},
		//		"simple item update": {
		//			shoppingList: []model.PantryItem{
		//				{
		//					ID:     0,
		//					Name:   "bac",
		//					Amount: 1,
		//					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
		//				}},
		//			update: "[1] 3 bacon\n",
		//			expected: []model.PantryItem{
		//				{
		//					ID:     0,
		//					Name:   "bacon",
		//					Amount: 3,
		//					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
		//				},
		//			},
		//		},
		//		"complex update": {
		//			shoppingList: []model.PantryItem{
		//				{
		//					ID:     0,
		//					Name:   "coffee",
		//					Amount: 2,
		//					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
		//				}, {
		//					ID:     0,
		//					Name:   "eggz",
		//					Amount: 4,
		//					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
		//				}, {
		//					ID:     0,
		//					Name:   "milk",
		//					Amount: 1,
		//					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
		//				},
		//			},
		//			update: "[1] 1 bacon\n[2] 2 eggs\n\n[3] milk",
		//			expected: []model.PantryItem{
		//				{
		//					ID:     0,
		//					Name:   "bacon",
		//					Amount: 1,
		//					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
		//				}, {
		//					ID:     0,
		//					Name:   "eggs",
		//					Amount: 2,
		//					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
		//				}, {
		//					ID:     0,
		//					Name:   "milk",
		//					Amount: 1,
		//					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
		//				},
		//			},
		//		},
		//		"complex update + added items": {
		//			shoppingList: []model.PantryItem{
		//				{
		//					ID:     0,
		//					Name:   "eggos",
		//					Amount: 4,
		//					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
		//				}, {
		//					ID:     0,
		//					Name:   "milk",
		//					Amount: 1,
		//					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
		//				},
		//			},
		//			update: "bacon\n[1] 2 eggs\n[2] milk\n6 beer",
		//			expected: []model.PantryItem{
		//				{
		//					ID:     0,
		//					Name:   "eggs",
		//					Amount: 2,
		//					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
		//				}, {
		//					ID:     0,
		//					Name:   "milk",
		//					Amount: 1,
		//					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
		//				}, {
		//					ID:     0,
		//					Name:   "bacon",
		//					Amount: 1,
		//					Date:   time.Now().Truncate(time.Minute),
		//				}, {
		//					ID:     0,
		//					Name:   "beer",
		//					Amount: 6,
		//					Date:   time.Now().Truncate(time.Minute),
		//				},
		//			},
		//		},
		//		"remove item": {
		//			shoppingList: []model.PantryItem{
		//				{
		//					ID:     0,
		//					Name:   "eggos",
		//					Amount: 4,
		//					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
		//				}, {
		//					ID:     0,
		//					Name:   "milk",
		//					Amount: 1,
		//					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
		//				},
		//			},
		//			update: "[1] 2 eggs\n",
		//			expected: []model.PantryItem{
		//				{
		//					ID:     0,
		//					Name:   "eggs",
		//					Amount: 2,
		//					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
		//				},
		//			},
		//		},
	}

	// and
	pantryClient := repository.NewSqlitePantryClient(repository.NewDatabaseClient(), "unit_tests")

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// and
			for _, item := range test.pantryItems {
				pantryClient.AddItem(item)
			}

			// when
			UpdateItemsFromModal(pantryClient, test.modalInput)
			actual := pantryClient.GetItems()

			// then
			assert.EqualValues(t, test.expected, actual)
		})
	}
}

//func TestRemove(t *testing.T) {
//	// given
//	tests := map[string]struct {
//		content  string
//		expected []model.PantryItem
//	}{
//		"single remove": {
//			content: "7",
//			expected: []model.PantryItem{
//				{ID: 1, Name: "item", Amount: 1, Date: time.Now().Truncate(time.Minute)},
//				{ID: 2, Name: "item", Amount: 2, Date: time.Now().Truncate(time.Minute)},
//				{ID: 3, Name: "item", Amount: 3, Date: time.Now().Truncate(time.Minute)},
//				{ID: 4, Name: "item", Amount: 4, Date: time.Now().Truncate(time.Minute)},
//				{ID: 5, Name: "item", Amount: 5, Date: time.Now().Truncate(time.Minute)},
//				{ID: 6, Name: "item", Amount: 6, Date: time.Now().Truncate(time.Minute)},
//				{ID: 7, Name: "item", Amount: 8, Date: time.Now().Truncate(time.Minute)},
//				{ID: 8, Name: "item", Amount: 9, Date: time.Now().Truncate(time.Minute)},
//			},
//		},
//		"multi remove": {
//			content: "3 5 8",
//			expected: []model.PantryItem{
//				{ID: 1, Name: "item", Amount: 1, Date: time.Now().Truncate(time.Minute)},
//				{ID: 2, Name: "item", Amount: 2, Date: time.Now().Truncate(time.Minute)},
//				{ID: 3, Name: "item", Amount: 4, Date: time.Now().Truncate(time.Minute)},
//				{ID: 4, Name: "item", Amount: 6, Date: time.Now().Truncate(time.Minute)},
//				{ID: 5, Name: "item", Amount: 7, Date: time.Now().Truncate(time.Minute)},
//				{ID: 6, Name: "item", Amount: 9, Date: time.Now().Truncate(time.Minute)},
//			},
//		},
//		"single and range remove": {
//			content: "1 4-7",
//			expected: []model.PantryItem{
//				{ID: 1, Name: "item", Amount: 2, Date: time.Now().Truncate(time.Minute)},
//				{ID: 2, Name: "item", Amount: 3, Date: time.Now().Truncate(time.Minute)},
//				{ID: 3, Name: "item", Amount: 8, Date: time.Now().Truncate(time.Minute)},
//				{ID: 4, Name: "item", Amount: 9, Date: time.Now().Truncate(time.Minute)},
//			},
//		},
//		"range remove": {
//			content: "2-5",
//			expected: []model.PantryItem{
//				{ID: 1, Name: "item", Amount: 1, Date: time.Now().Truncate(time.Minute)},
//				{ID: 2, Name: "item", Amount: 6, Date: time.Now().Truncate(time.Minute)},
//				{ID: 3, Name: "item", Amount: 7, Date: time.Now().Truncate(time.Minute)},
//				{ID: 4, Name: "item", Amount: 8, Date: time.Now().Truncate(time.Minute)},
//				{ID: 5, Name: "item", Amount: 9, Date: time.Now().Truncate(time.Minute)},
//			},
//		},
//		"remove all": {
//			content:  "*",
//			expected: []model.PantryItem{},
//		},
//		"remove all except single": {
//			content: "* 5",
//			expected: []model.PantryItem{
//				{ID: 1, Name: "item", Amount: 5, Date: time.Now().Truncate(time.Minute)},
//			},
//		},
//		"remove all except multi": {
//			content: "* 5 2 8",
//			expected: []model.PantryItem{
//				{ID: 1, Name: "item", Amount: 2, Date: time.Now().Truncate(time.Minute)},
//				{ID: 2, Name: "item", Amount: 5, Date: time.Now().Truncate(time.Minute)},
//				{ID: 3, Name: "item", Amount: 8, Date: time.Now().Truncate(time.Minute)},
//			},
//		},
//		"remove all except range": {
//			content: "* 3-6",
//			expected: []model.PantryItem{
//				{ID: 1, Name: "item", Amount: 3, Date: time.Now().Truncate(time.Minute)},
//				{ID: 2, Name: "item", Amount: 4, Date: time.Now().Truncate(time.Minute)},
//				{ID: 3, Name: "item", Amount: 5, Date: time.Now().Truncate(time.Minute)},
//				{ID: 4, Name: "item", Amount: 6, Date: time.Now().Truncate(time.Minute)},
//			},
//		},
//		"remove all except single and range": {
//			content: "* 7 1-3",
//			expected: []model.PantryItem{
//				{ID: 1, Name: "item", Amount: 1, Date: time.Now().Truncate(time.Minute)},
//				{ID: 2, Name: "item", Amount: 2, Date: time.Now().Truncate(time.Minute)},
//				{ID: 3, Name: "item", Amount: 3, Date: time.Now().Truncate(time.Minute)},
//				{ID: 4, Name: "item", Amount: 7, Date: time.Now().Truncate(time.Minute)},
//			},
//		},
//	}
//
//	for name, test := range tests {
//		t.Run(name, func(t *testing.T) {
//			// and
//			var items []model.PantryItem
//			for i := 1; i < 10; i++ {
//				items = add(items, fmt.Sprintf("item %d", i), time.Now().Truncate(time.Minute))
//			}
//
//			// when
//			actual := remove(items, test.content)
//
//			// then
//			assert.EqualValues(t, test.expected, actual)
//		})
//	}
//}
//
//func TestAdd(t *testing.T) {
//	// given
//	tests := map[string]struct {
//		content  string
//		expected []model.PantryItem
//	}{
//		"simple add": {
//			content:  "bacon",
//			expected: []model.PantryItem{{ID: 1, Name: "bacon", Amount: 1, Date: time.Now().Truncate(time.Minute)}},
//		},
//		"simple multi word add": {
//			content:  "butter scotch",
//			expected: []model.PantryItem{{ID: 1, Name: "butter scotch", Amount: 1, Date: time.Now().Truncate(time.Minute)}},
//		},
//		"simple hyphened add": {
//			content:  "dry-gin",
//			expected: []model.PantryItem{{ID: 1, Name: "dry-gin", Amount: 1, Date: time.Now().Truncate(time.Minute)}},
//		},
//		"add with trailing quantity": {
//			content:  "bacon 5",
//			expected: []model.PantryItem{{ID: 1, Name: "bacon", Amount: 5, Date: time.Now().Truncate(time.Minute)}},
//		},
//		"add with leading quantity": {
//			content:  "13 bacon",
//			expected: []model.PantryItem{{ID: 1, Name: "bacon", Amount: 13, Date: time.Now().Truncate(time.Minute)}},
//		},
//		"add with numbered name": {
//			content:  "2 monkey47",
//			expected: []model.PantryItem{{ID: 1, Name: "monkey47", Amount: 2, Date: time.Now().Truncate(time.Minute)}},
//		},
//	}
//
//	for name, test := range tests {
//		t.Run(name, func(t *testing.T) {
//			// when
//			actual := add([]model.PantryItem{}, test.content, time.Now().Truncate(time.Minute))
//
//			// then
//			assert.EqualValues(t, test.expected, actual)
//		})
//	}
//}
