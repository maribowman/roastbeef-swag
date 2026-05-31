package service

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/maribowman/roastbeef-swag/app/repository"
	"github.com/stretchr/testify/assert"
)

func TestUpdateItemsFromModal(t *testing.T) {
	// freeze the clock so production and expected timestamps share one instant
	fixed := time.Now()
	now = func() time.Time { return fixed }
	t.Cleanup(func() { now = time.Now })

	// where
	tests := map[string]struct {
		pantryItems []model.PantryItem
		modalInput  string
		expected    []model.PantryItem
	}{
		"simple quantity update": {
			pantryItems: []model.PantryItem{
				{Name: "bacon", Quantity: 1, Date: now().Truncate(time.Minute)},
			},
			modalInput: "[1] 3 bacon",
			expected: []model.PantryItem{
				{ID: 1, Name: "bacon", Quantity: 3, Date: now().Truncate(time.Minute)},
			},
		},
		"simple item update": {
			pantryItems: []model.PantryItem{
				{Name: "BAC", Quantity: 1, Date: now().Truncate(time.Minute)},
			},
			modalInput: "[1] 3 bacon",
			expected: []model.PantryItem{
				{ID: 1, Name: "bacon", Quantity: 3, Date: now().Truncate(time.Minute)},
			},
		},
		"complex update": {
			pantryItems: []model.PantryItem{
				{Name: "coffee", Quantity: 2, Date: now().Truncate(time.Minute)},
				{Name: "eggz", Quantity: 4, Date: now().Truncate(time.Minute)},
				{Name: "milk", Quantity: 1, Date: now().Truncate(time.Minute)},
			},
			modalInput: "[1] 1 bacon\n[2] 2 eggs\n\n[3] milk",
			expected: []model.PantryItem{
				{ID: 1, Name: "bacon", Quantity: 1, Date: now().Truncate(time.Minute)},
				{ID: 2, Name: "eggs", Quantity: 2, Date: now().Truncate(time.Minute)},
				{ID: 3, Name: "milk", Quantity: 1, Date: now().Truncate(time.Minute)},
			},
		},
		"complex update + added items": {
			pantryItems: []model.PantryItem{
				{Name: "eggos", Quantity: 4, Date: now().Truncate(time.Minute)},
				{Name: "milk", Quantity: 1, Date: now().Truncate(time.Minute)},
			},
			modalInput: "bacon\n[1] 2 eggs\n[2] milk\n6 beer",
			expected: []model.PantryItem{
				{ID: 1, Name: "eggs", Quantity: 2, Date: now().Truncate(time.Minute)},
				{ID: 2, Name: "milk", Quantity: 1, Date: now().Truncate(time.Minute)},
				{ID: 3, Name: "bacon", Quantity: 1, Date: now().Truncate(time.Minute)},
				{ID: 4, Name: "beer", Quantity: 6, Date: now().Truncate(time.Minute)},
			},
		},
		"remove item": {
			pantryItems: []model.PantryItem{
				{Name: "eggos", Quantity: 4, Date: now().Truncate(time.Minute)},
				{Name: "milk", Quantity: 1, Date: now().Truncate(time.Minute)},
			},
			modalInput: "[1] 2 eggs\n",
			expected: []model.PantryItem{
				{ID: 1, Name: "eggs", Quantity: 2, Date: now().Truncate(time.Minute)},
			},
		},
		"out-of-range modal index adds a new item": {
			pantryItems: []model.PantryItem{
				{Name: "milk", Quantity: 1, Date: now().Truncate(time.Minute)},
			},
			modalInput: "[1] milk\n[5] foo",
			expected: []model.PantryItem{
				{ID: 1, Name: "milk", Quantity: 1, Date: now().Truncate(time.Minute)},
				{ID: 2, Name: "foo", Quantity: 1, Date: now().Truncate(time.Minute)},
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
	// freeze the clock so production and expected timestamps share one instant
	fixed := time.Now()
	now = func() time.Time { return fixed }
	t.Cleanup(func() { now = time.Now })

	// where
	tests := map[string]struct {
		pantryItemCount int
		input           string
		expected        []model.PantryItem
	}{
		// EDIT TEST CASES
		"simple quantity update": {
			pantryItemCount: 1,
			input:           "1++",
			expected: []model.PantryItem{
				{ID: 1, Name: "Item #1", Quantity: 2, Date: now().Truncate(time.Minute)},
			},
		},
		"advanced quantity update": {
			pantryItemCount: 2,
			input:           "2--3",
			expected: []model.PantryItem{
				{ID: 1, Name: "Item #1", Quantity: 1, Date: now().Truncate(time.Minute)},
				{ID: 2, Name: "Item #2", Quantity: 1, Date: now().Truncate(time.Minute)},
			},
		},
		"advanced negative quantity update exception": {
			pantryItemCount: 1,
			input:           "1--5",
			expected:        nil,
		},
		"quantity update to zero removes item": {
			pantryItemCount: 1,
			input:           "1--1",
			expected:        nil,
		},

		// REMOVE TEST CASES
		"single number remove": {
			pantryItemCount: 3,
			input:           "2",
			expected: []model.PantryItem{
				{ID: 1, Name: "Item #1", Quantity: 1, Date: now().Truncate(time.Minute)},
				{ID: 3, Name: "Item #3", Quantity: 9, Date: now().Truncate(time.Minute)},
			},
		},
		"multi number remove": {
			pantryItemCount: 5,
			input:           "2 4 5",
			expected: []model.PantryItem{
				{ID: 1, Name: "Item #1", Quantity: 1, Date: now().Truncate(time.Minute)},
				{ID: 3, Name: "Item #3", Quantity: 9, Date: now().Truncate(time.Minute)},
			},
		},
		"single range remove": {
			pantryItemCount: 5,
			input:           "2-5",
			expected: []model.PantryItem{
				{ID: 1, Name: "Item #1", Quantity: 1, Date: now().Truncate(time.Minute)},
			},
		},
		"multi range remove": {
			pantryItemCount: 10,
			input:           "1-3 5-9",
			expected: []model.PantryItem{
				{ID: 4, Name: "Item #4", Quantity: 16, Date: now().Truncate(time.Minute)},
				{ID: 10, Name: "Item #10", Quantity: 100, Date: now().Truncate(time.Minute)},
			},
		},
		"single number and single range remove": {
			pantryItemCount: 5,
			input:           "1 3-5",
			expected: []model.PantryItem{
				{ID: 2, Name: "Item #2", Quantity: 4, Date: now().Truncate(time.Minute)},
			},
		},
		"multi number and multi range remove": {
			pantryItemCount: 15,
			input:           "1 3 5-10 12-15",
			expected: []model.PantryItem{
				{ID: 2, Name: "Item #2", Quantity: 4, Date: now().Truncate(time.Minute)},
				{ID: 4, Name: "Item #4", Quantity: 16, Date: now().Truncate(time.Minute)},
				{ID: 11, Name: "Item #11", Quantity: 121, Date: now().Truncate(time.Minute)},
			},
		},
		"remove all": {
			pantryItemCount: 5,
			input:           "*",
			expected:        nil,
		},
		"remove all except single number": {
			pantryItemCount: 1,
			input:           "* 1",
			expected: []model.PantryItem{
				{ID: 1, Name: "Item #1", Quantity: 1, Date: now().Truncate(time.Minute)},
			},
		},
		"remove all except multi number": {
			pantryItemCount: 5,
			input:           "* 2 4",
			expected: []model.PantryItem{
				{ID: 2, Name: "Item #2", Quantity: 4, Date: now().Truncate(time.Minute)},
				{ID: 4, Name: "Item #4", Quantity: 16, Date: now().Truncate(time.Minute)},
			},
		},
		"remove all except single range": {
			pantryItemCount: 5,
			input:           "* 1-3",
			expected: []model.PantryItem{
				{ID: 1, Name: "Item #1", Quantity: 1, Date: now().Truncate(time.Minute)},
				{ID: 2, Name: "Item #2", Quantity: 4, Date: now().Truncate(time.Minute)},
				{ID: 3, Name: "Item #3", Quantity: 9, Date: now().Truncate(time.Minute)},
			},
		},
		"remove all except multi range": {
			pantryItemCount: 10,
			input:           "* 1-3 5-6",
			expected: []model.PantryItem{
				{ID: 1, Name: "Item #1", Quantity: 1, Date: now().Truncate(time.Minute)},
				{ID: 2, Name: "Item #2", Quantity: 4, Date: now().Truncate(time.Minute)},
				{ID: 3, Name: "Item #3", Quantity: 9, Date: now().Truncate(time.Minute)},
				{ID: 5, Name: "Item #5", Quantity: 25, Date: now().Truncate(time.Minute)},
				{ID: 6, Name: "Item #6", Quantity: 36, Date: now().Truncate(time.Minute)},
			},
		},
		"remove all except single number and single range": {
			pantryItemCount: 5,
			input:           "* 5 1-3",
			expected: []model.PantryItem{
				{ID: 1, Name: "Item #1", Quantity: 1, Date: now().Truncate(time.Minute)},
				{ID: 2, Name: "Item #2", Quantity: 4, Date: now().Truncate(time.Minute)},
				{ID: 3, Name: "Item #3", Quantity: 9, Date: now().Truncate(time.Minute)},
				{ID: 5, Name: "Item #5", Quantity: 25, Date: now().Truncate(time.Minute)},
			},
		},
		"remove all except multi number and multi range": {
			pantryItemCount: 10,
			input:           "* 1 6 3-5 7-8",
			expected: []model.PantryItem{
				{ID: 1, Name: "Item #1", Quantity: 1, Date: now().Truncate(time.Minute)},
				{ID: 3, Name: "Item #3", Quantity: 9, Date: now().Truncate(time.Minute)},
				{ID: 4, Name: "Item #4", Quantity: 16, Date: now().Truncate(time.Minute)},
				{ID: 5, Name: "Item #5", Quantity: 25, Date: now().Truncate(time.Minute)},
				{ID: 6, Name: "Item #6", Quantity: 36, Date: now().Truncate(time.Minute)},
				{ID: 7, Name: "Item #7", Quantity: 49, Date: now().Truncate(time.Minute)},
				{ID: 8, Name: "Item #8", Quantity: 64, Date: now().Truncate(time.Minute)},
			},
		},

		// ADD TEST CASES
		"simple add": {
			pantryItemCount: 0,
			input:           "bacon",
			expected: []model.PantryItem{
				{ID: 1, Name: "bacon", Quantity: 1, Date: now().Truncate(time.Minute)},
			},
		},
		"simple multi word add": {
			pantryItemCount: 0,
			input:           "butter scotch",
			expected: []model.PantryItem{
				{ID: 1, Name: "butter scotch", Quantity: 1, Date: now().Truncate(time.Minute)},
			},
		},
		"simple hyphened add": {
			pantryItemCount: 0,
			input:           "dry-gin",
			expected: []model.PantryItem{
				{ID: 1, Name: "dry-gin", Quantity: 1, Date: now().Truncate(time.Minute)}},
		},
		"add with trailing quantity": {
			pantryItemCount: 0,
			input:           "bacon 5",
			expected: []model.PantryItem{
				{ID: 1, Name: "bacon", Quantity: 5, Date: now().Truncate(time.Minute)}},
		},
		"add with leading quantity": {
			pantryItemCount: 0,
			input:           "13 bacon",
			expected: []model.PantryItem{
				{ID: 1, Name: "bacon", Quantity: 13, Date: now().Truncate(time.Minute)}},
		},
		"add with numbered name": {
			pantryItemCount: 0,
			input:           "2 monkey47",
			expected: []model.PantryItem{
				{ID: 1, Name: "monkey47", Quantity: 2, Date: now().Truncate(time.Minute)}},
		},
		"multi line add": {
			pantryItemCount: 0,
			input:           "3 bacon\ncoffee 4",
			expected: []model.PantryItem{
				{ID: 1, Name: "bacon", Quantity: 3, Date: now().Truncate(time.Minute)},
				{ID: 2, Name: "coffee", Quantity: 4, Date: now().Truncate(time.Minute)},
			},
		},
	}

	// given
	pantryClient := repository.NewSqlitePantryClient(repository.NewDatabaseClient(), "unit_tests")

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// and
			for i := 1; i <= test.pantryItemCount; i++ {
				pantryClient.AddItem(model.PantryItem{
					Name:     fmt.Sprintf("Item #%d", i),
					Quantity: i * i,
					Date:     now().Truncate(time.Minute),
				})
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

func TestSplitMarkdownTable(t *testing.T) {
	// where
	const open = "```md\n"
	const close = "```"

	body := func(table string) string {
		return strings.TrimSuffix(strings.TrimPrefix(table, open), close)
	}

	smallTable := open + "| 1 | milk | 1 |\n" + close
	largeTable := open
	for i := 1; i <= 20; i++ {
		largeTable += fmt.Sprintf("| %d | item-%d | %d |\n", i, i, i)
	}
	largeTable += close

	tests := map[string]struct {
		table     string
		max       int
		minChunks int
	}{
		"single chunk when within limit": {
			table:     smallTable,
			max:       2000,
			minChunks: 1,
		},
		"splits into multiple chunks when over limit": {
			table:     largeTable,
			max:       80,
			minChunks: 2,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// when
			chunks := splitMarkdownTable(test.table, test.max)

			// then
			assert.GreaterOrEqual(t, len(chunks), test.minChunks)

			var rejoined string
			for _, chunk := range chunks {
				assert.LessOrEqual(t, len(chunk), test.max)
				assert.True(t, strings.HasPrefix(chunk, open), "chunk must open a code block")
				assert.True(t, strings.HasSuffix(chunk, close), "chunk must close a code block")
				assert.NotContains(t, chunk, "...", "no truncation marker should remain")
				rejoined += body(chunk)
			}

			// no rows dropped: rejoined chunk bodies reproduce the original table body
			assert.Equal(t, body(test.table), rejoined)
		})
	}

	t.Run("single chunk equals input", func(t *testing.T) {
		chunks := splitMarkdownTable(smallTable, 2000)
		assert.Equal(t, []string{smallTable}, chunks)
	})
}

func TestPartitionChannelMessages(t *testing.T) {
	// where
	const botID = "bot"

	base := time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)
	msg := func(id, authorID, content string, ageMinutes int) *discordgo.Message {
		return &discordgo.Message{
			ID:        id,
			Author:    &discordgo.User{ID: authorID},
			Content:   content,
			Timestamp: base.Add(time.Duration(ageMinutes) * time.Minute),
		}
	}

	tests := map[string]struct {
		messages          []*discordgo.Message
		expectedKeptID    string
		expectedRemovable []string
		expectedInput     string
	}{
		"keeps oldest bot message and removes the rest": {
			// Discord returns newest first; the oldest bot message must be reused.
			messages: []*discordgo.Message{
				msg("bot-new", botID, "", 20),
				msg("bot-mid", botID, "", 10),
				msg("bot-old", botID, "", 0),
			},
			expectedKeptID:    "bot-old",
			expectedRemovable: []string{"bot-new", "bot-mid"},
			expectedInput:     "",
		},
		"removes user messages and accumulates their input": {
			messages: []*discordgo.Message{
				msg("user-2", "alice", "milk", 30),
				msg("bot-old", botID, "", 10),
				msg("user-1", "bob", "eggs", 5),
			},
			expectedKeptID:    "bot-old",
			expectedRemovable: []string{"user-2", "user-1"},
			expectedInput:     "\nmilk\neggs",
		},
		"single bot message is kept with nothing removable": {
			messages: []*discordgo.Message{
				msg("bot-only", botID, "", 0),
			},
			expectedKeptID:    "bot-only",
			expectedRemovable: nil,
			expectedInput:     "",
		},
		"no bot message keeps nothing and removes all user messages": {
			messages: []*discordgo.Message{
				msg("user-1", "alice", "bread", 0),
				msg("user-2", "bob", "butter", 5),
			},
			expectedKeptID:    "",
			expectedRemovable: []string{"user-1", "user-2"},
			expectedInput:     "\nbread\nbutter",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// when
			keptID, input, removable := partitionChannelMessages(test.messages, botID)

			// then
			assert.Equal(t, test.expectedKeptID, keptID)
			assert.Equal(t, test.expectedInput, input)
			assert.Equal(t, test.expectedRemovable, removable)

			// regression guard: every bot message except the kept one must be removed,
			// so no stale "standing posts" can linger.
			for _, m := range test.messages {
				if m.Author.ID == botID && m.ID != keptID {
					assert.Contains(t, removable, m.ID, "stale bot message must be removable")
				}
			}
		})
	}
}
