package model

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

type PantryItem struct {
	ID     int
	Name   string
	Amount int
	Date   time.Time
}

func ToList(items []PantryItem) string {
	var shoppingList string
	for index, item := range items {
		if index != 0 {
			shoppingList += "\n"
		}
		shoppingList += fmt.Sprintf("[%d] %d %s", index+1, item.Amount, item.Name)
	}
	return shoppingList
}

func ToMarkdownTable(items []PantryItem, linebreak int, dateFormat string) string {
	var data [][]string
	for index, item := range items {
		data = append(data, processItemMarkdownLayout(index+1, item, linebreak, dateFormat)...)
	}
	return writeMarkdownTable(data)
}

func processItemMarkdownLayout(index int, item PantryItem, linebreak int, dateFormat string) [][]string {
	if len(item.Name) <= linebreak {
		return [][]string{{
			strconv.Itoa(index),
			item.Name,
			strconv.Itoa(item.Amount),
			item.Date.Format(dateFormat),
		},
		}
	}

	// Split item name in whitespace separated chunks
	itemNameSplit := strings.Split(item.Name, " ")
	nameLine := ""
	nameMultiLine := []string{}

	for idx, split := range itemNameSplit {
		if len(nameLine) > 0 {
			// Add whitespace separator to simply append stuff
			nameLine += " "
		}

		if len(split) > linebreak {
			// Split too long item word
			charsLeft := linebreak - len(nameLine) - 1
			nameMultiLine = append(nameMultiLine, fmt.Sprintf("%s%s-", nameLine, split[:charsLeft]))
			nameLine = split[charsLeft:]
			// Split a second time in case of a really long word (covers 99.9%)
			if len(nameLine) > linebreak {
				nameMultiLine = append(nameMultiLine, fmt.Sprintf("%s-", nameLine[:linebreak-1]))
				nameLine = nameLine[linebreak-1:]
			}
		} else if len(nameLine)+len(split) > linebreak {
			// Create newline before name line gets too long
			nameMultiLine = append(nameMultiLine, strings.TrimSpace(nameLine))
			// Assign leftover split to item line
			nameLine = split
		} else {
			nameLine += split
		}

		// Wrap up last line
		if idx == len(itemNameSplit)-1 {
			nameMultiLine = append(nameMultiLine, strings.TrimSpace(nameLine))
		}
	}

	var result [][]string
	for idx, line := range nameMultiLine {
		if idx == 0 {
			result = append(result, []string{
				strconv.Itoa(index),
				line,
				strconv.Itoa(item.Amount),
				item.Date.Format(dateFormat),
			})
			continue
		}

		result = append(result, []string{
			" ",
			line,
			" ",
			" ",
		})
	}
	return result
}

func writeMarkdownTable(data [][]string) string {
	writer := bytes.Buffer{}
	writer.WriteString("```md\n")

	table := tablewriter.NewTable(&writer,
		tablewriter.WithRenderer(renderer.NewMarkdown(
			tw.Rendition{
				Borders: tw.Border{
					Left:   tw.On,
					Top:    tw.Off,
					Right:  tw.On,
					Bottom: tw.Off,
				},
			},
		)),
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignNone},
			},
			Row: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignNone},
			},
		}),
	)

	table.Header([]string{"#", "ITEM", "QTY", "ADDED"})
	table.Bulk(data)
	table.Render()

	writer.WriteString("```")

	return writer.String()
}
