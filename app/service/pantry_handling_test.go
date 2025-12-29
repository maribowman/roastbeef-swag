package service

import (
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
				{
					Name:     "bacon",
					Quantity: 1,
					Date:     time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
			modalInput: "[1] 3 bacon",
			expected: []model.PantryItem{
				{
					ID:       1,
					Name:     "bacon",
					Quantity: 3,
					Date:     time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
		},
		"simple item update": {
			pantryItems: []model.PantryItem{
				{
					Name:     "BAC",
					Quantity: 1,
					Date:     time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}},
			modalInput: "[1] 3 bacon",
			expected: []model.PantryItem{
				{
					ID:       1,
					Name:     "bacon",
					Quantity: 3,
					Date:     time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
		},
		"complex update": {
			pantryItems: []model.PantryItem{
				{
					Name:     "coffee",
					Quantity: 2,
					Date:     time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					Name:     "eggz",
					Quantity: 4,
					Date:     time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					Name:     "milk",
					Quantity: 1,
					Date:     time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
			modalInput: "[1] 1 bacon\n[2] 2 eggs\n\n[3] milk",
			expected: []model.PantryItem{
				{
					ID:       1,
					Name:     "bacon",
					Quantity: 1,
					Date:     time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:       2,
					Name:     "eggs",
					Quantity: 2,
					Date:     time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:       3,
					Name:     "milk",
					Quantity: 1,
					Date:     time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
		},
		"complex update + added items": {
			pantryItems: []model.PantryItem{
				{
					Name:     "eggos",
					Quantity: 4,
					Date:     time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					Name:     "milk",
					Quantity: 1,
					Date:     time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
			modalInput: "bacon\n[1] 2 eggs\n[2] milk\n6 beer",
			expected: []model.PantryItem{
				{
					ID:       1,
					Name:     "eggs",
					Quantity: 2,
					Date:     time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:       2,
					Name:     "milk",
					Quantity: 1,
					Date:     time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:       3,
					Name:     "bacon",
					Quantity: 1,
					Date:     time.Now().Truncate(time.Minute),
				}, {
					ID:       4,
					Name:     "beer",
					Quantity: 6,
					Date:     time.Now().Truncate(time.Minute),
				},
			},
		},
		"remove item": {
			pantryItems: []model.PantryItem{
				{
					Name:     "eggos",
					Quantity: 4,
					Date:     time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					Name:     "milk",
					Quantity: 1,
					Date:     time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
			modalInput: "[1] 2 eggs\n",
			expected: []model.PantryItem{
				{
					ID:       1,
					Name:     "eggs",
					Quantity: 2,
					Date:     time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
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
		pantryItems []model.PantryItem
		input       string
		expected    []model.PantryItem
	}{
		// EDIT TEST CASES
		"simple quantity update": {
			pantryItems: []model.PantryItem{
				{
					Name:     "bacon",
					Quantity: 3,
					Date:     time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
			input: "1++",
			expected: []model.PantryItem{
				{
					ID:       1,
					Name:     "bacon",
					Quantity: 4,
					Date:     time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
		},
		"advanced quantity update": {
			pantryItems: []model.PantryItem{
				{
					Name:     "bacon",
					Quantity: 4,
					Date:     time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
			input: "1--2",
			expected: []model.PantryItem{
				{
					ID:       1,
					Name:     "bacon",
					Quantity: 2,
					Date:     time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
		},
		"advanced quantity update exception": {
			pantryItems: []model.PantryItem{
				{
					Name:     "bacon",
					Quantity: 5,
					Date:     time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
			input:    "1--5",
			expected: nil,
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
			UpdateItems(pantryClient, test.input)

			// then
			actual := pantryClient.GetItems()
			assert.EqualValues(t, test.expected, actual)

			// cleanup
			pantryClient.RemoveAllItems()
		})
	}
}

func TestDetermineIndices(t *testing.T) {
	// where
	tests := map[string]struct {
		input    string
		expected []int
	}{
		"single number remove": {
			input:    "7",
			expected: []int{7},
		},
		"multi number remove": {
			input:    "3 5 8",
			expected: []int{3, 5, 8},
		},
		"single range remove": {
			input:    "2-5",
			expected: []int{2, 3, 4, 5},
		},
		"multi range remove": {
			input:    "1-3 7-9",
			expected: []int{1, 2, 3, 7, 8, 9},
		},
		"single number and single range remove": {
			input:    "1 4-7",
			expected: []int{1, 4, 5, 6, 7},
		},
		"multi number and multi range remove": {
			input:    "1 3 5-7 9-11",
			expected: []int{1, 3, 5, 6, 7, 9, 10, 11},
		},
		//		"remove all": {
		//			input:  "*",
		//			expected: []model.PantryItem{},
		//		},
		//		"remove all except single": {
		//			input: "* 5",
		//			expected: []model.PantryItem{
		//				{ID: 1, Name: "item", Quantity: 5, Date: time.Now().Truncate(time.Minute)},
		//			},
		//		},
		//		"remove all except multi": {
		//			input: "* 5 2 8",
		//			expected: []model.PantryItem{
		//				{ID: 1, Name: "item", Quantity: 2, Date: time.Now().Truncate(time.Minute)},
		//				{ID: 2, Name: "item", Quantity: 5, Date: time.Now().Truncate(time.Minute)},
		//				{ID: 3, Name: "item", Quantity: 8, Date: time.Now().Truncate(time.Minute)},
		//			},
		//		},
		//		"remove all except range": {
		//			input: "* 3-6",
		//			expected: []model.PantryItem{
		//				{ID: 1, Name: "item", Quantity: 3, Date: time.Now().Truncate(time.Minute)},
		//				{ID: 2, Name: "item", Quantity: 4, Date: time.Now().Truncate(time.Minute)},
		//				{ID: 3, Name: "item", Quantity: 5, Date: time.Now().Truncate(time.Minute)},
		//				{ID: 4, Name: "item", Quantity: 6, Date: time.Now().Truncate(time.Minute)},
		//			},
		//		},
		//		"remove all except single and range": {
		//			input: "* 7 1-3",
		//			expected: []model.PantryItem{
		//				{ID: 1, Name: "item", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
		//				{ID: 2, Name: "item", Quantity: 2, Date: time.Now().Truncate(time.Minute)},
		//				{ID: 3, Name: "item", Quantity: 3, Date: time.Now().Truncate(time.Minute)},
		//				{ID: 4, Name: "item", Quantity: 7, Date: time.Now().Truncate(time.Minute)},
		//			},
		//		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// when
			actual := determineIndices(test.input)

			// then
			assert.EqualValues(t, test.expected, actual)
		})
	}
}

func TestGenerateNewPantryItem(t *testing.T) {
	// where
	tests := map[string]struct {
		input    string
		expected model.PantryItem
	}{
		"simple add": {
			input:    "bacon",
			expected: model.PantryItem{Name: "bacon", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
		},
		"simple multi word add": {
			input:    "butter scotch",
			expected: model.PantryItem{Name: "butter scotch", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
		},
		"simple hyphened add": {
			input:    "dry-gin",
			expected: model.PantryItem{Name: "dry-gin", Quantity: 1, Date: time.Now().Truncate(time.Minute)},
		},
		"add with trailing quantity": {
			input:    "bacon 5",
			expected: model.PantryItem{Name: "bacon", Quantity: 5, Date: time.Now().Truncate(time.Minute)},
		},
		"add with leading quantity": {
			input:    "13 bacon",
			expected: model.PantryItem{Name: "bacon", Quantity: 13, Date: time.Now().Truncate(time.Minute)},
		},
		"add with numbered name": {
			input:    "2 monkey47",
			expected: model.PantryItem{Name: "monkey47", Quantity: 2, Date: time.Now().Truncate(time.Minute)},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// when
			actual := generateNewPantryItem(test.input)

			// then
			assert.EqualValues(t, test.expected, actual)
		})
	}
}
