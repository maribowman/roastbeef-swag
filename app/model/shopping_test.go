package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFromShoppingListTable(t *testing.T) {
	// given
	tests := map[string]struct {
		table    string
		expected []ShoppingListItem
	}{
		"simple conversion": {
			table: "```md\n" +
				"| # | ITEM | QTY | ADDED  |\n" +
				"|---|------|-----|--------|\n" +
				"| 1 | test | 3   | 27.12. |\n" +
				"```",
			expected: []ShoppingListItem{{
				ID:     1,
				Item:   "test",
				Amount: 3,
				Date:   time.Date(time.Now().Year(), 12, 27, 0, 0, 0, 0, time.Local),
			}},
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
			expected: []ShoppingListItem{{
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
			}},
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
