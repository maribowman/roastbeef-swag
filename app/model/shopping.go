package model

import (
	"bytes"
	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
	"time"
)

type ShoppingListItem struct {
	ID     int
	Item   string
	Amount int
	Date   time.Time
}

func ToShoppingList(items []ShoppingListItem) string {
	var shoppingList string
	for index, item := range items {
		if index != 0 {
			shoppingList += "\n"
		}
		shoppingList += item.Item
		if item.Amount > 1 {
			shoppingList += "\t" + "\t" + strconv.Itoa(item.Amount)
		}
	}
	log.Info().Msg(shoppingList)
	return shoppingList
}

func FromShoppingList(shoppingList string) []ShoppingListItem {
	return []ShoppingListItem{}
}

func ToShoppingListTable(items []ShoppingListItem) string {
	var data [][]string
	for _, item := range items {
		data = append(data, []string{
			strconv.Itoa(item.ID),
			item.Item,
			strconv.Itoa(item.Amount),
			item.Date.Format("02.01.")},
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

func FromShoppingListTable(table string) []ShoppingListItem {
	var result []ShoppingListItem
	splitTable := strings.Split(table, "\n")

	for index, item := range splitTable {
		if index <= 2 || index == len(splitTable)-1 {
			continue
		}

		splitItem := strings.Split(item, "|")
		id, _ := strconv.Atoi(strings.TrimSpace(splitItem[1]))
		amount, _ := strconv.Atoi(strings.TrimSpace(splitItem[3]))
		date, _ := time.Parse("02.01.", strings.TrimSpace(splitItem[4]))

		result = append(result, ShoppingListItem{
			ID:     id,
			Item:   strings.TrimSpace(splitItem[2]),
			Amount: amount,
			Date:   time.Date(time.Now().Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local),
		})
	}
	return result
}
