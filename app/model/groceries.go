package model

import (
	"bytes"
	"github.com/olekukonko/tablewriter"
	"strconv"
	"time"
)

type ShoppingEntry struct {
	ID     int
	Item   string
	Amount int
	Date   time.Time
}

func CreateShoppingListTable(shoppingList []ShoppingEntry) string {
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
	writer.WriteString("```md\n")

	table := tablewriter.NewWriter(&writer)
	table.SetHeader([]string{"#", "Item", "QTY", "Added"})
	table.SetHeaderAlignment(tablewriter.ALIGN_CENTER)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data)
	table.Render()

	writer.WriteString("```")

	return writer.String()
}
