package service

import (
	"encoding/json"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/stretchr/testify/assert"
)

func TestExtractSuggestions(t *testing.T) {
	// where
	tests := map[string]struct {
		input               string
		expectedRest        string
		expectedSuggestions []string
	}{
		"no suggestions": {
			input:        "eggs 3\n2",
			expectedRest: "eggs 3\n2",
		},
		"only suggestion": {
			input:               "? 2 milk",
			expectedRest:        "",
			expectedSuggestions: []string{"2 milk"},
		},
		"suggestion without space": {
			input:               "?milk",
			expectedRest:        "",
			expectedSuggestions: []string{"milk"},
		},
		"bare question mark is dropped": {
			input:        "?\n  ?  ",
			expectedRest: "",
		},
		"indented suggestion": {
			input:               "  ? jam  ",
			expectedRest:        "",
			expectedSuggestions: []string{"jam"},
		},
		"mixed lines keep their order": {
			input:               "eggs\n? butter\n3 bagels\n? jam\n2-4",
			expectedRest:        "eggs\n3 bagels\n2-4",
			expectedSuggestions: []string{"butter", "jam"},
		},
		"question mark inside a line is no suggestion": {
			input:        "milk?",
			expectedRest: "milk?",
		},
		"windows line endings": {
			input:               "eggs\r\n? butter\r\nbagels",
			expectedRest:        "eggs\nbagels",
			expectedSuggestions: []string{"butter"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// when
			rest, suggestions := extractSuggestions(test.input)

			// then
			assert.Equal(t, test.expectedRest, rest)
			assert.Equal(t, test.expectedSuggestions, suggestions)
		})
	}
}

func TestFormatParseSuggestion(t *testing.T) {
	// where
	tests := map[string]struct {
		item            model.PantryItem
		authorID        string
		expectedContent string
		expectedText    string
	}{
		"simple item": {
			item:            model.PantryItem{Name: "milk", Quantity: 2},
			authorID:        "123456789012345678",
			expectedContent: "💡 Suggestion: 2 milk (by <@123456789012345678>)",
			expectedText:    "2 milk",
		},
		"name with spaces": {
			item:            model.PantryItem{Name: "oat milk barista", Quantity: 1},
			authorID:        "42",
			expectedContent: "💡 Suggestion: 1 oat milk barista (by <@42>)",
			expectedText:    "1 oat milk barista",
		},
		"name with umlauts": {
			item:            model.PantryItem{Name: "Käsespätzle", Quantity: 3},
			authorID:        "42",
			expectedContent: "💡 Suggestion: 3 Käsespätzle (by <@42>)",
			expectedText:    "3 Käsespätzle",
		},
		"name with colons and parentheses": {
			item:            model.PantryItem{Name: "tea: earl grey (by the box)", Quantity: 1},
			authorID:        "42",
			expectedContent: "💡 Suggestion: 1 tea: earl grey (by the box) (by <@42>)",
			expectedText:    "1 tea: earl grey (by the box)",
		},
		"without author": {
			item:            model.PantryItem{Name: "jam", Quantity: 1},
			expectedContent: "💡 Suggestion: 1 jam",
			expectedText:    "1 jam",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// when
			content := formatSuggestion(test.item, test.authorID)
			text, ok := parseSuggestion(content)

			// then
			assert.Equal(t, test.expectedContent, content)
			assert.True(t, ok)
			assert.Equal(t, test.expectedText, text)
			assert.Equal(t, test.item.Name, generateNewPantryItem(text).Name)
			assert.Equal(t, test.item.Quantity, generateNewPantryItem(text).Quantity)
		})
	}
}

func TestParseSuggestionRejectsOtherContent(t *testing.T) {
	// where
	tests := map[string]string{
		"empty":           "",
		"markdown table":  "```md\n| # | ITEM | QTY | ADDED |\n```",
		"missing item":    "💡 Suggestion: ",
		"wrong prefix":    "Suggestion: 2 milk",
		"plain user text": "2 milk",
	}

	for name, content := range tests {
		t.Run(name, func(t *testing.T) {
			// when
			_, ok := parseSuggestion(content)

			// then
			assert.False(t, ok)
		})
	}
}

func TestIsSuggestionMessage(t *testing.T) {
	// given
	// Message.Components is not marshalled (json:"-"), so build the Discord payload from the components
	fromDiscord := func(t *testing.T, components []discordgo.MessageComponent) *discordgo.Message {
		raw, err := json.Marshal(map[string]any{"components": components})
		assert.NoError(t, err)
		var msg discordgo.Message
		assert.NoError(t, json.Unmarshal(raw, &msg))
		return &msg
	}

	// where
	tests := map[string]struct {
		message  func(t *testing.T) *discordgo.Message
		expected bool
	}{
		"locally built suggestion": {
			message: func(t *testing.T) *discordgo.Message {
				return &discordgo.Message{Components: createSuggestionButtons()}
			},
			expected: true,
		},
		"suggestion unmarshalled from Discord": {
			message: func(t *testing.T) *discordgo.Message {
				return fromDiscord(t, createSuggestionButtons())
			},
			expected: true,
		},
		"table unmarshalled from Discord": {
			message: func(t *testing.T) *discordgo.Message {
				return fromDiscord(t, CreateMessageButtons())
			},
			expected: false,
		},
		"message without components": {
			message: func(t *testing.T) *discordgo.Message {
				return &discordgo.Message{Content: "? milk"}
			},
			expected: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// when
			actual := isSuggestionMessage(test.message(t))

			// then
			assert.Equal(t, test.expected, actual)
		})
	}
}
