package model

import (
	"bytes"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var idPrefixRegex = regexp.MustCompile(`^\[(\d+)]\s`)

type PantryItem struct {
	ID     int
	Item   string
	Amount int
	Date   time.Time
}

func ToList(items []PantryItem) string {
	var shoppingList string
	for index, item := range items {
		if index != 0 {
			shoppingList += "\n"
		}
		shoppingList += fmt.Sprintf("[%d] %s", index+1, item.Item)
		if item.Amount > 1 {
			shoppingList += "\t" + "\t, " + strconv.Itoa(item.Amount)
		}
	}
	return shoppingList
}

func UpdateFromList(items []PantryItem, updatedList string) []PantryItem {
	for _, update := range strings.Split(updatedList, "\n") {
		if strings.TrimSpace(update) == "" {
			continue
		}

		updateSplit := strings.Split(update, ",")
		rawID := idPrefixRegex.FindStringSubmatch(updateSplit[0])
		item := strings.TrimSpace(idPrefixRegex.ReplaceAllString(updateSplit[0], ""))

		var id int
		if len(rawID) == 2 { // matches full string + capture group
			id, _ = strconv.Atoi(rawID[1])
		} else {
			id = len(items) + 1
			items = append(items, PantryItem{
				ID:   id,
				Date: time.Now().Truncate(time.Minute),
			})
		}

		items[id-1].Item = item
		items[id-1].Amount = 1
		if len(updateSplit) == 2 {
			if amount, err := strconv.Atoi(strings.TrimSpace(updateSplit[1])); err == nil {
				items[id-1].Amount = amount
			}
		}
	}
	return items
}

func ToMarkdownTable(items []PantryItem, linebreak int, dateFormat string) string {
	var data [][]string
	for _, item := range items {
		tableItem := item.Item
		if len(item.Item) > 20 {
			tableItem = ""
			itemSplit := strings.Split(item.Item, " ")
			for _, split := range itemSplit {
				if len(split) > 20 {

				}
				if len(tableItem)+len(split)+1 > 20 {
					tableItem += "\n"
				}
				tableItem += split + " "
			}
		}

		data = append(data, []string{
			strconv.Itoa(item.ID),
			tableItem,
			strconv.Itoa(item.Amount),
			item.Date.Format(dateFormat)},
		)
	}

	writer := bytes.Buffer{}
	writer.WriteString("```md\n")

	table := tablewriter.NewWriter(&writer)
	table.SetHeader([]string{"#", "ITEM", "QTY", "ADDED"})
	table.SetHeaderAlignment(tablewriter.ALIGN_CENTER)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data)
	table.Render()

	writer.WriteString("```")

	return writer.String()
}

func FromMarkdownTable(table string) []PantryItem {
	var result []PantryItem
	splitTable := strings.Split(table, "\n")

	for index, item := range splitTable {
		if index <= 2 || index == len(splitTable)-1 {
			continue
		}

		splitItem := strings.Split(item, "|")
		id, _ := strconv.Atoi(strings.TrimSpace(splitItem[1]))
		amount, _ := strconv.Atoi(strings.TrimSpace(splitItem[3]))
		date, _ := time.Parse("02.01.", strings.TrimSpace(splitItem[4]))

		result = append(result, PantryItem{
			ID:     id,
			Item:   strings.TrimSpace(splitItem[2]),
			Amount: amount,
			Date:   time.Date(time.Now().Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local),
		})
	}
	return result
}
