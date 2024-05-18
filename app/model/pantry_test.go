package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestToMarkdownTable(t *testing.T) {
	// given
	tests := map[string]struct {
		items    []PantryItem
		expected string
	}{
		"no conversion": {
			items: []PantryItem{
				{1, "12345 12345 12345", 1, time.Date(2023, 12, 27, 0, 0, 0, 0, time.Local)},
			},
			expected: "```md\n" +
				"| # |       ITEM        | QTY |  ADDED   |\n" +
				"|---|-------------------|-----|----------|\n" +
				"| 1 | 12345 12345 12345 | 1   | 27.12.23 |\n" +
				"```",
		},
		"simple conversion": {
			items: []PantryItem{
				{1, "12345 12345 12345 12345 12345", 1, time.Date(2023, 12, 27, 0, 0, 0, 0, time.Local)},
			},
			expected: "```md\n" +
				"| # |       ITEM        | QTY |  ADDED   |\n" +
				"|---|-------------------|-----|----------|\n" +
				"| 1 | 12345 12345 12345 | 1   | 27.12.23 |\n" +
				"|   | 12345 12345       |     |          |\n" +
				"```",
		},
		"single too large item": {
			items: []PantryItem{
				{1, "1234512345123451234512345", 1, time.Date(2023, 12, 27, 0, 0, 0, 0, time.Local)},
			},
			expected: "```md\n" +
				"| # |         ITEM         | QTY |  ADDED   |\n" +
				"|---|----------------------|-----|----------|\n" +
				"| 1 | 1234512345123451234- | 1   | 27.12.23 |\n" +
				"|   | 512345               |     |          |\n" +
				"```",
		},
		"too large item": {
			items: []PantryItem{
				{1, "12345 1234512345123451234512345", 1, time.Date(2023, 12, 27, 0, 0, 0, 0, time.Local)},
			},
			expected: "```md\n" +
				"| # |         ITEM         | QTY |  ADDED   |\n" +
				"|---|----------------------|-----|----------|\n" +
				"| 1 | 12345 1234512345123- | 1   | 27.12.23 |\n" +
				"|   | 451234512345         |     |          |\n" +
				"```",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// when
			actual := ToMarkdownTable(test.items, 20, "02.01.06")

			// then
			assert.EqualValues(t, test.expected, actual)
		})
	}
}

func TestFromMarkdownTable(t *testing.T) {
	// given
	tests := map[string]struct {
		table    string
		expected []PantryItem
	}{
		"simple conversion": {
			table: "```md\n" +
				"| # | ITEM | QTY | ADDED  |\n" +
				"|---|------|-----|--------|\n" +
				"| 1 | test | 3   | 27.12. |\n" +
				"```",
			expected: []PantryItem{
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
			expected: []PantryItem{
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
		"multi-line conversion": {
			table: "```md\n" +
				"| # |     ITEM     | QTY | ADDED  |\n" +
				"|---|--------------|-----|--------|\n" +
				"| 1 | eggs         | 4   | 24.12. |\n" +
				"| 2 | coffee and   | 1   | 25.12. |\n" +
				"|   | more coffee  |     |        |\n" +
				"```",
			expected: []PantryItem{
				{
					ID:     1,
					Item:   "eggs",
					Amount: 4,
					Date:   time.Date(time.Now().Year(), 12, 24, 0, 0, 0, 0, time.Local),
				}, {
					ID:     2,
					Item:   "coffee and more coffee",
					Amount: 1,
					Date:   time.Date(time.Now().Year(), 12, 25, 0, 0, 0, 0, time.Local),
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// when
			actual := FromMarkdownTable(test.table, "02.01.")

			// then
			assert.EqualValues(t, test.expected, actual)
		})
	}
}
