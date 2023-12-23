package service

import (
	"fmt"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAdd(t *testing.T) {
	// given
	tests := map[string]struct {
		content  string
		expected []model.ShoppingEntry
	}{
		"simple add": {
			content: "bacon",
			expected: []model.ShoppingEntry{{
				ID:     1,
				Item:   "bacon",
				Amount: 1,
				Date:   time.Now().Truncate(time.Minute),
			}},
		},
		"add with trailing quantity": {
			content: "bacon 5",
			expected: []model.ShoppingEntry{{
				ID:     1,
				Item:   "bacon",
				Amount: 5,
				Date:   time.Now().Truncate(time.Minute),
			}},
		},
		"add with leading quantity": {
			content: "13 bacon",
			expected: []model.ShoppingEntry{{
				ID:     1,
				Item:   "bacon",
				Amount: 13,
				Date:   time.Now().Truncate(time.Minute),
			}},
		},
		"add with numbered name": {
			content: "2 monkey47",
			expected: []model.ShoppingEntry{{
				ID:     1,
				Item:   "monkey47",
				Amount: 2,
				Date:   time.Now().Truncate(time.Minute),
			}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// and
			bot := GroceryBot{}

			// when
			bot.add(test.content)

			// then
			assert.EqualValues(t, test.expected, bot.shoppingList)
		})
	}
}

func TestRemove(t *testing.T) {
	// given
	tests := map[string]struct {
		content  string
		expected []int
	}{
		"single remove": {
			content:  "7",
			expected: []int{1, 2, 3, 4, 5, 6, 8, 9},
		},
		"multi remove": {
			content:  "3 5 8",
			expected: []int{1, 2, 4, 6, 7, 9},
		},
		"remove all": {
			content:  "*",
			expected: []int{},
		},
		"single and range remove": {
			content:  "1 4-7",
			expected: []int{2, 3, 8, 9},
		},
		"range remove": {
			content:  "2-5",
			expected: []int{1, 6, 7, 8, 9},
		},
		"multi range remove": {
			content:  "1-3 6-8",
			expected: []int{4, 5, 9},
		},
		"multi and range remove": {
			content:  "2 3 5-7 9",
			expected: []int{1, 4, 8},
		},
		"remove all except single": {
			content:  "* 5",
			expected: []int{5},
		},
		"remove all except multi": {
			content:  "* 5 2 8",
			expected: []int{2, 5, 8},
		},
		"remove all except range": {
			content:  "* 3-6",
			expected: []int{3, 4, 5, 6},
		},
		"remove all except single and range": {
			content:  "* 7 1-3",
			expected: []int{1, 2, 3, 7},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// and
			bot := GroceryBot{}
			for i := 0; i < 9; i++ {
				bot.add(fmt.Sprintf("item %d", i))
			}

			// when
			bot.remove(test.content)

			// and
			actual := []int{}
			for _, item := range bot.shoppingList {
				actual = append(actual, item.ID)
			}

			// then
			assert.EqualValues(t, test.expected, actual)
		})
	}
}
