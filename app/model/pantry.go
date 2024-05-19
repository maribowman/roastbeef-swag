package model

import (
	"bytes"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"strconv"
	"strings"
	"time"
)

type PantryItem struct {
	ID     int
	Number int
	Item   string
	Amount int
	Date   time.Time
}

func (item *PantryItem) ToString() string {
	return fmt.Sprintf("id: `%d`; number:`%d`' item: `%s`; amount: `%d`; date: `%s`", item.ID, item.Number, item.Item, item.Amount, item.Date.Format("02.01.06"))
}

func ToList(items []PantryItem) string {
	var shoppingList string
	for index, item := range items {
		if index != 0 {
			shoppingList += "\n"
		}
		shoppingList += fmt.Sprintf("[%d] %d %s", index+1, item.Amount, item.Item)
	}
	return shoppingList
}

func ToMarkdownTable(items []PantryItem, linebreak int, dateFormat string) string {
	var data [][]string
	for _, item := range items {
		tableItemLines := []string{}

		if len(item.Item) < linebreak {
			tableItemLines = append(tableItemLines, item.Item)
		} else {
			// split item in white space separated chunks
			tableItemLine := ""
			itemSplit := strings.Split(item.Item, " ")

			for index, split := range itemSplit {
				if len(tableItemLine) != 0 {
					tableItemLine += " "
				}
				if len(split) > linebreak {
					// split too long item word
					charsLeft := linebreak - len(tableItemLine) - 1
					tableItemLines = append(tableItemLines, tableItemLine+split[:charsLeft]+"-")
					tableItemLine = split[charsLeft:]
					// split a second time in rare case of a mega long word
					if len(tableItemLine) > linebreak {
						tableItemLines = append(tableItemLines, tableItemLine[:linebreak-1]+"-")
						tableItemLine = tableItemLine[linebreak-1:]
					}
				} else if len(tableItemLine)+len(split) > linebreak {
					// create new line before table item line gets too long
					tableItemLines = append(tableItemLines, strings.TrimSpace(tableItemLine))
					// reset table item line
					tableItemLine = split
				} else {
					tableItemLine += split
				}
				// wrap up last line
				if index == len(itemSplit)-1 {
					tableItemLines = append(tableItemLines, strings.TrimSpace(tableItemLine))
					tableItemLine = ""
				}
			}
		}

		for index, tableItemLine := range tableItemLines {
			if index == 0 {
				data = append(data, []string{
					strconv.Itoa(item.Number),
					tableItemLine,
					strconv.Itoa(item.Amount),
					item.Date.Format(dateFormat)},
				)
			} else {
				data = append(data, []string{
					strconv.Itoa(item.Number),
					tableItemLine,
					"",
					""},
				)
			}
		}
	}

	writer := bytes.Buffer{}
	writer.WriteString("```md\n")

	table := tablewriter.NewWriter(&writer)
	table.SetHeader([]string{"#", "ITEM", "QTY", "ADDED"})
	table.SetHeaderAlignment(tablewriter.ALIGN_CENTER)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.SetAutoMergeCellsByColumnIndex([]int{0})
	table.AppendBulk(data)
	table.Render()

	writer.WriteString("```")

	return writer.String()
}

func FromMarkdownTable(table string, dateFormat string) []PantryItem {
	var result []PantryItem
	splitTable := strings.Split(table, "\n")

	for index, item := range splitTable {
		if index <= 2 || index == len(splitTable)-1 {
			continue
		}

		splitItem := strings.Split(item, "|")
		number, err := strconv.Atoi(strings.TrimSpace(splitItem[1]))
		if err != nil {
			// overwriting last item -> assuming it is a multi-line item because it does not have a number
			lastItem := result[len(result)-1]
			if strings.HasSuffix(lastItem.Item, "-") {
				lastItem.Item = strings.TrimSuffix(lastItem.Item, "-") + strings.TrimSpace(splitItem[2])
			} else {
				lastItem.Item += " " + strings.TrimSpace(splitItem[2])
			}
			result[len(result)-1] = lastItem
			continue
		}
		amount, _ := strconv.Atoi(strings.TrimSpace(splitItem[3]))
		date, _ := time.Parse(dateFormat, strings.TrimSpace(splitItem[4]))
		if date.Year() <= 0 {
			date = time.Date(time.Now().Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
		}

		result = append(result, PantryItem{
			Number: number,
			Item:   strings.TrimSpace(splitItem[2]),
			Amount: amount,
			Date:   date,
		})
	}
	return result
}
