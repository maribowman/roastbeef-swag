package service

import (
	"testing"
	"time"

	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/maribowman/roastbeef-swag/app/repository"
	"github.com/stretchr/testify/assert"
)

func TestUpdateFromModal(t *testing.T) {
	// where
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
		"simple item update": {
			pantryItems: []model.PantryItem{
				{
					Name:   "BAC",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}},
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
		"complex update": {
			pantryItems: []model.PantryItem{
				{
					Name:   "coffee",
					Amount: 2,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					Name:   "eggz",
					Amount: 4,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					Name:   "milk",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
			modalInput: "[1] 1 bacon\n[2] 2 eggs\n\n[3] milk",
			expected: []model.PantryItem{
				{
					ID:     1,
					Name:   "bacon",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:     2,
					Name:   "eggs",
					Amount: 2,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:     3,
					Name:   "milk",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
		},
		"complex update + added items": {
			pantryItems: []model.PantryItem{
				{
					Name:   "eggos",
					Amount: 4,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					Name:   "milk",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
			modalInput: "bacon\n[1] 2 eggs\n[2] milk\n6 beer",
			expected: []model.PantryItem{
				{
					ID:     1,
					Name:   "eggs",
					Amount: 2,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:     2,
					Name:   "milk",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					ID:     3,
					Name:   "bacon",
					Amount: 1,
					Date:   time.Now().Truncate(time.Minute),
				}, {
					ID:     4,
					Name:   "beer",
					Amount: 6,
					Date:   time.Now().Truncate(time.Minute),
				},
			},
		},
		"remove item": {
			pantryItems: []model.PantryItem{
				{
					Name:   "eggos",
					Amount: 4,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				}, {
					Name:   "milk",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
				},
			},
			modalInput: "[1] 2 eggs\n",
			expected: []model.PantryItem{
				{
					ID:     1,
					Name:   "eggs",
					Amount: 2,
					Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
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

func TestDetermineRemovableIndices(t *testing.T) {
	// where
	tests := map[string]struct {
		line     string
		expected []int
	}{
		"single number remove": {
			line:     "7",
			expected: []int{7},
		},
		"multi number remove": {
			line:     "3 5 8",
			expected: []int{3, 5, 8},
		},
		"single range remove": {
			line:     "2-5",
			expected: []int{2, 3, 4, 5},
		},
		"multi range remove": {
			line:     "1-3 7-9",
			expected: []int{1, 2, 3, 7, 8, 9},
		},
		"single number and single range remove": {
			line:     "1 4-7",
			expected: []int{1, 4, 5, 6, 7},
		},
		"multi number and multi range remove": {
			line:     "1 3 5-7 9-11",
			expected: []int{1, 3, 5, 6, 7, 9, 10, 11},
		},
		//		"remove all": {
		//			line:  "*",
		//			expected: []model.PantryItem{},
		//		},
		//		"remove all except single": {
		//			line: "* 5",
		//			expected: []model.PantryItem{
		//				{ID: 1, Name: "item", Amount: 5, Date: time.Now().Truncate(time.Minute)},
		//			},
		//		},
		//		"remove all except multi": {
		//			line: "* 5 2 8",
		//			expected: []model.PantryItem{
		//				{ID: 1, Name: "item", Amount: 2, Date: time.Now().Truncate(time.Minute)},
		//				{ID: 2, Name: "item", Amount: 5, Date: time.Now().Truncate(time.Minute)},
		//				{ID: 3, Name: "item", Amount: 8, Date: time.Now().Truncate(time.Minute)},
		//			},
		//		},
		//		"remove all except range": {
		//			line: "* 3-6",
		//			expected: []model.PantryItem{
		//				{ID: 1, Name: "item", Amount: 3, Date: time.Now().Truncate(time.Minute)},
		//				{ID: 2, Name: "item", Amount: 4, Date: time.Now().Truncate(time.Minute)},
		//				{ID: 3, Name: "item", Amount: 5, Date: time.Now().Truncate(time.Minute)},
		//				{ID: 4, Name: "item", Amount: 6, Date: time.Now().Truncate(time.Minute)},
		//			},
		//		},
		//		"remove all except single and range": {
		//			line: "* 7 1-3",
		//			expected: []model.PantryItem{
		//				{ID: 1, Name: "item", Amount: 1, Date: time.Now().Truncate(time.Minute)},
		//				{ID: 2, Name: "item", Amount: 2, Date: time.Now().Truncate(time.Minute)},
		//				{ID: 3, Name: "item", Amount: 3, Date: time.Now().Truncate(time.Minute)},
		//				{ID: 4, Name: "item", Amount: 7, Date: time.Now().Truncate(time.Minute)},
		//			},
		//		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// when
			actual := determineRemovableIndices(test.line)

			// then
			assert.EqualValues(t, test.expected, actual)
		})
	}
}

func TestGenerateNewPantryItem(t *testing.T) {
	// where
	tests := map[string]struct {
		line     string
		expected model.PantryItem
	}{
		"simple add": {
			line:     "bacon",
			expected: model.PantryItem{Name: "bacon", Amount: 1, Date: time.Now().Truncate(time.Minute)},
		},
		"simple multi word add": {
			line:     "butter scotch",
			expected: model.PantryItem{Name: "butter scotch", Amount: 1, Date: time.Now().Truncate(time.Minute)},
		},
		"simple hyphened add": {
			line:     "dry-gin",
			expected: model.PantryItem{Name: "dry-gin", Amount: 1, Date: time.Now().Truncate(time.Minute)},
		},
		"add with trailing quantity": {
			line:     "bacon 5",
			expected: model.PantryItem{Name: "bacon", Amount: 5, Date: time.Now().Truncate(time.Minute)},
		},
		"add with leading quantity": {
			line:     "13 bacon",
			expected: model.PantryItem{Name: "bacon", Amount: 13, Date: time.Now().Truncate(time.Minute)},
		},
		"add with numbered name": {
			line:     "2 monkey47",
			expected: model.PantryItem{Name: "monkey47", Amount: 2, Date: time.Now().Truncate(time.Minute)},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// when
			actual := generateNewPantryItem(test.line)

			// then
			assert.EqualValues(t, test.expected, actual)
		})
	}
}
