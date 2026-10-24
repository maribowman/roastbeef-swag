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

	// suggestionMsg mimics a suggestion message as returned by Discord (pointer components).
	suggestionMsg := func(id, content string, ageMinutes int) *discordgo.Message {
		m := msg(id, botID, content, ageMinutes)
		m.Components = []discordgo.MessageComponent{
			&discordgo.ActionsRow{Components: []discordgo.MessageComponent{
				&discordgo.Button{CustomID: SuggestionAcceptButton},
				&discordgo.Button{CustomID: SuggestionDenyButton},
			}},
		}
		return m
	}

	tests := map[string]struct {
		messages            []*discordgo.Message
		suggestionsEnabled  bool
		expectedKeptID      string
		expectedRemovable   []string
		expectedInput       string
		expectedSuggestions []suggestion
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
			expectedInput:     "eggs\nmilk",
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
			expectedInput:     "bread\nbutter",
		},
		"newest-first user input is applied oldest first": {
			// ChannelMessages returns newest first: `1--` was sent after `butter`.
			messages: []*discordgo.Message{
				msg("user-2", "alice", "1--", 10),
				msg("bot-old", botID, "", 0),
				msg("user-1", "alice", "butter", 5),
			},
			expectedKeptID:    "bot-old",
			expectedRemovable: []string{"user-2", "user-1"},
			expectedInput:     "butter\n1--",
		},
		"three user messages are ordered oldest first": {
			messages: []*discordgo.Message{
				msg("user-3", "alice", "2", 30),
				msg("user-2", "bob", "3", 20),
				msg("user-1", "alice", "milk", 10),
				msg("bot-old", botID, "", 0),
			},
			expectedKeptID:    "bot-old",
			expectedRemovable: []string{"user-3", "user-2", "user-1"},
			expectedInput:     "milk\n3\n2",
		},
		"equal timestamps fall back to numeric snowflake ID order": {
			// "10" < "9" as plain strings, but snowflake 9 was created first.
			messages: []*discordgo.Message{
				msg("10", "alice", "1--", 5),
				msg("9", "alice", "butter", 5),
				msg("8", botID, "", 0),
			},
			expectedKeptID:    "8",
			expectedRemovable: []string{"10", "9"},
			expectedInput:     "butter\n1--",
		},
		"suggestion messages are neither kept nor removable": {
			messages: []*discordgo.Message{
				suggestionMsg("sugg-2", "💡 Suggestion: 1 jam", 30),
				msg("user-1", "alice", "eggs", 25),
				suggestionMsg("sugg-1", "💡 Suggestion: 2 milk", 20),
				msg("bot-new", botID, "", 10),
				msg("bot-old", botID, "", 0),
			},
			suggestionsEnabled: true,
			expectedKeptID:     "bot-old",
			expectedRemovable:  []string{"user-1", "bot-new"},
			expectedInput:      "eggs",
		},
		"suggestion older than the table is not picked as the table": {
			// e.g. the table was deleted by hand and republished below an open suggestion
			messages: []*discordgo.Message{
				msg("bot-table", botID, "", 10),
				suggestionMsg("sugg-old", "💡 Suggestion: 2 milk", 0),
			},
			suggestionsEnabled: true,
			expectedKeptID:     "bot-table",
			expectedRemovable:  nil,
			expectedInput:      "",
		},
		"only suggestion messages keep nothing and remove nothing": {
			messages: []*discordgo.Message{
				suggestionMsg("sugg-1", "💡 Suggestion: 2 milk", 0),
			},
			suggestionsEnabled: true,
			expectedKeptID:     "",
			expectedRemovable:  nil,
			expectedInput:      "",
		},
		"suggestion lines are split off per message with their author, oldest first": {
			messages: []*discordgo.Message{
				msg("user-2", "bob", "? jam\n2", 20),
				msg("user-1", "alice", "eggs\n? 2 milk\n?butter", 10),
				msg("bot-old", botID, "", 0),
			},
			suggestionsEnabled: true,
			expectedKeptID:     "bot-old",
			expectedRemovable:  []string{"user-2", "user-1"},
			expectedInput:      "eggs\n2",
			expectedSuggestions: []suggestion{
				{Text: "2 milk", AuthorID: "alice"},
				{Text: "butter", AuthorID: "alice"},
				{Text: "jam", AuthorID: "bob"},
			},
		},
		"suggestion-only message adds no input": {
			messages: []*discordgo.Message{
				msg("user-1", "alice", "? 2 milk", 10),
				msg("bot-old", botID, "", 0),
			},
			suggestionsEnabled:  true,
			expectedKeptID:      "bot-old",
			expectedRemovable:   []string{"user-1"},
			expectedInput:       "",
			expectedSuggestions: []suggestion{{Text: "2 milk", AuthorID: "alice"}},
		},
		"suggestion lines stay input when suggestions are disabled": {
			messages: []*discordgo.Message{
				msg("user-1", "alice", "? fish", 10),
				msg("bot-old", botID, "", 0),
			},
			suggestionsEnabled: false,
			expectedKeptID:     "bot-old",
			expectedRemovable:  []string{"user-1"},
			expectedInput:      "? fish",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// when
			keptID, input, removable, suggestions := partitionChannelMessages(test.messages, botID, test.suggestionsEnabled)

			// then
			assert.Equal(t, test.expectedKeptID, keptID)
			assert.Equal(t, test.expectedInput, input)
			assert.Equal(t, test.expectedRemovable, removable)
			assert.Equal(t, test.expectedSuggestions, suggestions)

			// regression guard: every bot message except the kept one and open
			// suggestions must be removed, so no stale "standing posts" can linger.
			for _, m := range test.messages {
				if isSuggestionMessage(m) {
					assert.NotContains(t, removable, m.ID, "suggestion message must survive the sync")
					continue
				}
				if m.Author.ID == botID && m.ID != keptID {
					assert.Contains(t, removable, m.ID, "stale bot message must be removable")
				}
			}
		})
	}
}

func TestUndoLastAction(t *testing.T) {
	// freeze the clock so production and expected timestamps share one instant
	fixed := time.Now()
	now = func() time.Time { return fixed }
	t.Cleanup(func() { now = time.Now })

	today := now().Truncate(time.Minute)
	yesterday := today.Add(-24 * time.Hour)
	lastWeek := today.Add(-7 * 24 * time.Hour)

	type step struct {
		kind  string // "input", "modal" or "undo"
		input string
	}
	apply := func(client model.PantryClient, s step) {
		switch s.kind {
		case "input":
			ApplyUserInput(client, s.input)
		case "modal":
			ApplyModalInput(client, s.input)
		case "undo":
			// mirrors the UndoButton action: restore only, never snapshot
			client.RestoreSnapshot()
		}
	}

	initialItems := []model.PantryItem{
		{Name: "bacon", Quantity: 2, Date: lastWeek},
		{Name: "eggs", Quantity: 6, Date: yesterday},
		{Name: "milk", Quantity: 1, Date: today},
	}
	initialList := []model.PantryItem{
		{ID: 1, Name: "bacon", Quantity: 2, Date: lastWeek},
		{ID: 2, Name: "eggs", Quantity: 6, Date: yesterday},
		{ID: 3, Name: "milk", Quantity: 1, Date: today},
	}

	// where
	tests := map[string]struct {
		steps    []step
		expected []model.PantryItem
	}{
		"add then undo": {
			steps:    []step{{kind: "input", input: "3 coffee"}, {kind: "undo"}},
			expected: initialList,
		},
		"remove all then undo": {
			steps:    []step{{kind: "input", input: "*"}, {kind: "undo"}},
			expected: initialList,
		},
		"modal edit then undo": {
			steps:    []step{{kind: "modal", input: "[1] 5 bacon\n[3] oat milk\nbeer"}, {kind: "undo"}},
			expected: initialList,
		},
		"two actions then undo reverts only the second": {
			steps: []step{{kind: "input", input: "2"}, {kind: "input", input: "1++"}, {kind: "undo"}},
			expected: []model.PantryItem{
				{ID: 1, Name: "bacon", Quantity: 2, Date: lastWeek},
				{ID: 3, Name: "milk", Quantity: 1, Date: today},
			},
		},
		"undo twice changes nothing on the second click": {
			steps: []step{{kind: "input", input: "2"}, {kind: "input", input: "beer"}, {kind: "undo"}, {kind: "undo"}},
			expected: []model.PantryItem{
				{ID: 1, Name: "bacon", Quantity: 2, Date: lastWeek},
				{ID: 3, Name: "milk", Quantity: 1, Date: today},
			},
		},
		"undo without prior action changes nothing": {
			steps:    []step{{kind: "undo"}},
			expected: initialList,
		},
	}

	tableIndex := 0
	for name, test := range tests {
		tableIndex++
		t.Run(name, func(t *testing.T) {
			// given: a fresh table per case, so no snapshot leaks between cases
			pantryClient := repository.NewSqlitePantryClient(repository.NewDatabaseClient(), fmt.Sprintf("undo_tests_%d", tableIndex))
			for _, item := range initialItems {
				pantryClient.AddItem(item)
			}

			// when
			for _, s := range test.steps {
				apply(pantryClient, s)
			}

			// then
			assert.EqualValues(t, test.expected, pantryClient.GetItems())
		})
	}
}

func TestApplySuggestion(t *testing.T) {
	// freeze the clock so production and expected timestamps share one instant
	fixed := time.Now()
	now = func() time.Time { return fixed }
	t.Cleanup(func() { now = time.Now })

	today := now().Truncate(time.Minute)
	yesterday := today.Add(-24 * time.Hour)
	initialList := []model.PantryItem{
		{ID: 1, Name: "eggs", Quantity: 6, Date: yesterday},
	}

	// where
	tests := map[string]struct {
		suggestion        string
		expectedAfterAdd  []model.PantryItem
		expectedAfterUndo []model.PantryItem
	}{
		"leading quantity is added": {
			suggestion: "2 milk",
			expectedAfterAdd: []model.PantryItem{
				{ID: 1, Name: "eggs", Quantity: 6, Date: yesterday},
				{ID: 2, Name: "milk", Quantity: 2, Date: today},
			},
			expectedAfterUndo: initialList,
		},
		"name without quantity defaults to 1": {
			suggestion: "oat milk",
			expectedAfterAdd: []model.PantryItem{
				{ID: 1, Name: "eggs", Quantity: 6, Date: yesterday},
				{ID: 2, Name: "oat milk", Quantity: 1, Date: today},
			},
			expectedAfterUndo: initialList,
		},
	}

	// TestUndoLastAction uses undo_tests_1..n, so start high to keep snapshots apart
	tableIndex := 100
	for name, test := range tests {
		tableIndex++
		t.Run(name, func(t *testing.T) {
			// given
			pantryClient := repository.NewSqlitePantryClient(repository.NewDatabaseClient(), fmt.Sprintf("undo_tests_%d", tableIndex))
			pantryClient.AddItem(model.PantryItem{Name: "eggs", Quantity: 6, Date: yesterday})

			// when
			ApplySuggestion(pantryClient, test.suggestion)

			// then
			assert.EqualValues(t, test.expectedAfterAdd, pantryClient.GetItems())

			// when: undo, mirroring the UndoButton action
			pantryClient.RestoreSnapshot()

			// then
			assert.EqualValues(t, test.expectedAfterUndo, pantryClient.GetItems())
		})
	}
}
