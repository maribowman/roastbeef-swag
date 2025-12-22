package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToMarkdownTable(t *testing.T) {
	// given
	tests := map[string]struct {
		items    []PantryItem
		expected string
	}{
		"no conversion": {
			items: []PantryItem{
				{0, "12345 12345 12345", 1, time.Date(2023, 12, 27, 0, 0, 0, 0, time.Local)},
			},
			expected: "```md\n" +
				"| # |       ITEM        | QTY |  ADDED   |\n" +
				"|---|-------------------|-----|----------|\n" +
				"| 1 | 12345 12345 12345 | 1   | 27.12.23 |\n" +
				"```",
		},
		"simple conversion": {
			items: []PantryItem{
				{0, "12345 12345 12345 12345 12345", 1, time.Date(2023, 12, 27, 0, 0, 0, 0, time.Local)},
			},
			expected: "```md\n" +
				"| # |       ITEM        | QTY |  ADDED   |\n" +
				"|---|-------------------|-----|----------|\n" +
				"| 1 | 12345 12345 12345 | 1   | 27.12.23 |\n" +
				"|   | 12345 12345       |     |          |\n" +
				"```",
		},
		"too large name with single word split": {
			items: []PantryItem{
				{0, "1234512345123451234512345", 1, time.Date(2023, 12, 27, 0, 0, 0, 0, time.Local)},
			},
			expected: "```md\n" +
				"| # |         ITEM         | QTY |  ADDED   |\n" +
				"|---|----------------------|-----|----------|\n" +
				"| 1 | 1234512345123451234- | 1   | 27.12.23 |\n" +
				"|   | 512345               |     |          |\n" +
				"```",
		},
		"too large name with double word split": {
			items: []PantryItem{
				{0, "1234512345123451234512345123451234512345", 1, time.Date(2023, 12, 27, 0, 0, 0, 0, time.Local)},
			},
			expected: "```md\n" +
				"| # |         ITEM         | QTY |  ADDED   |\n" +
				"|---|----------------------|-----|----------|\n" +
				"| 1 | 1234512345123451234- | 1   | 27.12.23 |\n" +
				"|   | 5123451234512345123- |     |          |\n" +
				"|   | 45                   |     |          |\n" +
				"```",
		},
		"too large name with whitespace and word split": {
			items: []PantryItem{
				{0, "12345 1234512345123451234512345", 1, time.Date(2023, 12, 27, 0, 0, 0, 0, time.Local)},
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
