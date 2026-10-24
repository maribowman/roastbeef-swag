package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/rs/zerolog/log"
)

const (
	SuggestionAcceptButton = "suggestion-accept"
	SuggestionDenyButton   = "suggestion-deny"

	suggestionPrefix = "💡 Suggestion: "
)

// suggestion is a single `?` line extracted from a user message, together with its author.
type suggestion struct{ Text, AuthorID string }

var (
	// Detects suggestion lines like `? 2 milk` or `?milk`
	suggestionLineRegex = regexp.MustCompile(`^\?\s*(.*)$`)
	// Reads the item text back from a published suggestion message; the author suffix is optional
	suggestionContentRegex = regexp.MustCompile(`^` + regexp.QuoteMeta(suggestionPrefix) + `(.+?)(?: \(by <@\d+>\))?$`)
)

// extractSuggestions splits off all `?` lines from the input. It returns the remaining input
// (line order kept) and the trimmed suggestion bodies; empty suggestions are dropped.
func extractSuggestions(input string) (rest string, suggestions []string) {
	var restLines []string
	for line := range strings.Lines(input) {
		matches := suggestionLineRegex.FindStringSubmatch(strings.TrimSpace(line))
		if matches == nil {
			restLines = append(restLines, strings.TrimRight(line, "\r\n"))
			continue
		}

		body := strings.TrimSpace(matches[1])
		if body == "" {
			log.Debug().Msg("Dropping empty suggestion")
			continue
		}
		suggestions = append(suggestions, body)
	}
	return strings.Join(restLines, "\n"), suggestions
}

// formatSuggestion renders the content of a suggestion message. The item text can be read back
// with parseSuggestion, so open suggestions need no state besides the Discord message itself.
func formatSuggestion(item model.PantryItem, authorID string) string {
	content := fmt.Sprintf("%s%d %s", suggestionPrefix, item.Quantity, item.Name)
	if authorID != "" {
		content += fmt.Sprintf(" (by <@%s>)", authorID)
	}
	return content
}

// parseSuggestion reads the `<qty> <name>` item text back from a suggestion message content.
func parseSuggestion(content string) (string, bool) {
	matches := suggestionContentRegex.FindStringSubmatch(content)
	if matches == nil {
		return "", false
	}
	return matches[1], true
}

// isSuggestionMessage reports whether the message carries the suggestion accept button.
// Components unmarshalled from Discord are pointers, locally built ones are values; both are handled.
func isSuggestionMessage(msg *discordgo.Message) bool {
	for _, component := range msg.Components {
		var buttons []discordgo.MessageComponent
		switch row := component.(type) {
		case *discordgo.ActionsRow:
			buttons = row.Components
		case discordgo.ActionsRow:
			buttons = row.Components
		}

		for _, button := range buttons {
			switch b := button.(type) {
			case *discordgo.Button:
				if b.CustomID == SuggestionAcceptButton {
					return true
				}
			case discordgo.Button:
				if b.CustomID == SuggestionAcceptButton {
					return true
				}
			}
		}
	}
	return false
}

// createSuggestionButtons generates the accept and deny buttons for a suggestion message
func createSuggestionButtons() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Emoji: &discordgo.ComponentEmoji{
						Name: "👍",
					},
					CustomID: SuggestionAcceptButton,
					Style:    discordgo.SuccessButton,
				},
				discordgo.Button{
					Emoji: &discordgo.ComponentEmoji{
						Name: "👎",
					},
					CustomID: SuggestionDenyButton,
					Style:    discordgo.DangerButton,
				},
			},
		},
	}
}
