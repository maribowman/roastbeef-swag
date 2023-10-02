package service

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseContent(t *testing.T) {
	// given
	bot := NewGroceryBot("_", "_")

	// and
	tests := map[string]struct {
		content  string
		expected string
	}{
		"simple add": {
			content:  "bacon",
			expected: add,
		},
		"add with quantity": {
			content:  "bacon 5",
			expected: add,
		},
		"update entry": {
			content:  "1 bacon",
			expected: update,
		},
		"update entry and quantity": {
			content:  "1 bacon 2",
			expected: update,
		},
		"simple remove": {
			content:  "0",
			expected: remove,
		},
		"multi remove": {
			content:  "0 1",
			expected: remove,
		},
		"remove all": {
			content:  "*",
			expected: remove,
		},
		"remove all except": {
			content:  "* 5 2 8",
			expected: remove,
		},
		"undo": {
			content:  "&",
			expected: undo,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// when
			actual := bot.ParseContent(test.content)

			// then
			assert.EqualValues(t, test.expected, actual)
		})
	}
}
