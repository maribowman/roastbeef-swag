package model

import (
	"bytes"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/ini.v1"
	"strconv"
	"strings"
	"time"
)

type ShoppingEntry struct {
	ID     int
	Item   string
	Amount int
	Date   time.Time
}

func ToShoppingListTable(shoppingList []ShoppingEntry) string {
	var data [][]string
	for _, entry := range shoppingList {
		data = append(data, []string{
			strconv.Itoa(entry.ID),
			entry.Item,
			strconv.Itoa(entry.Amount),
			entry.Date.Format("02.01.")},
		)
	}

	writer := bytes.Buffer{}
	writer.WriteString("```md" + ini.LineBreak)

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

func FromShoppingListTable(table string) []ShoppingEntry {
	var result []ShoppingEntry
	splitTable := strings.Split(table, ini.LineBreak)

	for index, item := range splitTable {
		if index <= 2 || index == len(splitTable)-1 {
			continue
		}

		splitItem := strings.Split(item, "|")
		id, _ := strconv.Atoi(strings.TrimSpace(splitItem[1]))
		amount, _ := strconv.Atoi(strings.TrimSpace(splitItem[3]))
		date, _ := time.Parse("02.01.", strings.TrimSpace(splitItem[4]))

		result = append(result, ShoppingEntry{
			ID:     id,
			Item:   strings.TrimSpace(splitItem[2]),
			Amount: amount,
			Date:   time.Date(time.Now().Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local),
		})
	}
	return result
}
