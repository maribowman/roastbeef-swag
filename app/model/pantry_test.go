package model

import (
	"testing"
	"time"
	"unicode/utf8"

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
		"word too long for remaining room starts on fresh line": {
			items: []PantryItem{
				{0, "abcdefghijklmnopqrs abcdefghijklmnopqrstuvwxyz", 1, time.Date(2023, 12, 27, 0, 0, 0, 0, time.Local)},
			},
			expected: "```md\n" +
				"| # |         ITEM         | QTY |  ADDED   |\n" +
				"|---|----------------------|-----|----------|\n" +
				"| 1 | abcdefghijklmnopqrs  | 1   | 27.12.23 |\n" +
				"|   | abcdefghijklmnopqrs- |     |          |\n" +
				"|   | tuvwxyz              |     |          |\n" +
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

func TestProcessItemMarkdownLayout(t *testing.T) {
	// given
	tests := map[string]struct {
		name     string
		expected []string
	}{
		"word too long for remaining room": {
			name:     "abcdefghijklmnopqrs abcdefghijklmnopqrstuvwxyz",
			expected: []string{"abcdefghijklmnopqrs", "abcdefghijklmnopqrs-", "tuvwxyz"},
		},
		"first word exactly fills line followed by long word": {
			name:     "abcdefghijklmnopqrst 1234512345123451234512345",
			expected: []string{"abcdefghijklmnopqrst", "1234512345123451234-", "512345"},
		},
		"single word longer than three line breaks": {
			name:     "12345123451234512345123451234512345123451234512345123451234512345",
			expected: []string{"1234512345123451234-", "5123451234512345123-", "4512345123451234512-", "34512345"},
		},
		"umlaut on byte split point": {
			name:     "Hähnchenschenkelsüßsauer",
			expected: []string{"Hähnchenschenkelsüß-", "sauer"},
		},
		"umlaut words with no room left": {
			name:     "Hähnchenbrustfilet Grünkohlsuppenmischung",
			expected: []string{"Hähnchenbrustfilet", "Grünkohlsuppenmisch-", "ung"},
		},
		"umlaut name fitting line in runes but not in bytes": {
			name:     "Bärlauchpesto Gläser",
			expected: []string{"Bärlauchpesto Gläser"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// when
			var rows [][]string
			assert.NotPanics(t, func() {
				rows = processItemMarkdownLayout(1, PantryItem{Name: test.name, Quantity: 1}, 20, "02.01.06")
			})

			// then
			var actual []string
			for _, row := range rows {
				actual = append(actual, row[1])
				assert.True(t, utf8.ValidString(row[1]), "invalid UTF-8 in row %q", row[1])
				assert.LessOrEqual(t, utf8.RuneCountInString(row[1]), 20, "row %q exceeds line break", row[1])
			}
			assert.Equal(t, test.expected, actual)
		})
	}
}
