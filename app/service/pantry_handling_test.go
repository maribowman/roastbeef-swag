package service

import (
	"fmt"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRemove(t *testing.T) {
	// given
	tests := map[string]struct {
		content  string
		expected []model.PantryItem
	}{
		"single remove": {
			content: "7",
			expected: []model.PantryItem{
				{1, "item", 1, time.Now().Truncate(time.Minute)},
				{2, "item", 2, time.Now().Truncate(time.Minute)},
				{3, "item", 3, time.Now().Truncate(time.Minute)},
				{4, "item", 4, time.Now().Truncate(time.Minute)},
				{5, "item", 5, time.Now().Truncate(time.Minute)},
				{6, "item", 6, time.Now().Truncate(time.Minute)},
				{7, "item", 8, time.Now().Truncate(time.Minute)},
				{8, "item", 9, time.Now().Truncate(time.Minute)},
			},
		},
		"multi remove": {
			content: "3 5 8",
			expected: []model.PantryItem{
				{1, "item", 1, time.Now().Truncate(time.Minute)},
				{2, "item", 2, time.Now().Truncate(time.Minute)},
				{3, "item", 4, time.Now().Truncate(time.Minute)},
				{4, "item", 6, time.Now().Truncate(time.Minute)},
				{5, "item", 7, time.Now().Truncate(time.Minute)},
				{6, "item", 9, time.Now().Truncate(time.Minute)},
			},
		},
		"single and range remove": {
			content: "1 4-7",
			expected: []model.PantryItem{
				{1, "item", 2, time.Now().Truncate(time.Minute)},
				{2, "item", 3, time.Now().Truncate(time.Minute)},
				{3, "item", 8, time.Now().Truncate(time.Minute)},
				{4, "item", 9, time.Now().Truncate(time.Minute)},
			},
		},
		"range remove": {
			content: "2-5",
			expected: []model.PantryItem{
				{1, "item", 1, time.Now().Truncate(time.Minute)},
				{2, "item", 6, time.Now().Truncate(time.Minute)},
				{3, "item", 7, time.Now().Truncate(time.Minute)},
				{4, "item", 8, time.Now().Truncate(time.Minute)},
				{5, "item", 9, time.Now().Truncate(time.Minute)},
			},
		},
		"remove all": {
			content:  "*",
			expected: []model.PantryItem{},
		},
		"remove all except single": {
			content: "* 5",
			expected: []model.PantryItem{
				{1, "item", 5, time.Now().Truncate(time.Minute)},
			},
		},
		"remove all except multi": {
			content: "* 5 2 8",
			expected: []model.PantryItem{
				{1, "item", 2, time.Now().Truncate(time.Minute)},
				{2, "item", 5, time.Now().Truncate(time.Minute)},
				{3, "item", 8, time.Now().Truncate(time.Minute)},
			},
		},
		"remove all except range": {
			content: "* 3-6",
			expected: []model.PantryItem{
				{1, "item", 3, time.Now().Truncate(time.Minute)},
				{2, "item", 4, time.Now().Truncate(time.Minute)},
				{3, "item", 5, time.Now().Truncate(time.Minute)},
				{4, "item", 6, time.Now().Truncate(time.Minute)},
			},
		},
		"remove all except single and range": {
			content: "* 7 1-3",
			expected: []model.PantryItem{
				{1, "item", 1, time.Now().Truncate(time.Minute)},
				{2, "item", 2, time.Now().Truncate(time.Minute)},
				{3, "item", 3, time.Now().Truncate(time.Minute)},
				{4, "item", 7, time.Now().Truncate(time.Minute)},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// and
			var items []model.PantryItem
			for i := 1; i < 10; i++ {
				items = add(items, fmt.Sprintf("item %d", i))
			}

			// when
			actual := remove(items, test.content)

			// then
			assert.EqualValues(t, test.expected, actual)
		})
	}
}

func TestAdd(t *testing.T) {
	// given
	tests := map[string]struct {
		content  string
		expected []model.PantryItem
	}{
		"simple add": {
			content:  "bacon",
			expected: []model.PantryItem{{1, "bacon", 1, time.Now().Truncate(time.Minute)}},
		},
		"simple multi word add": {
			content:  "butter scotch",
			expected: []model.PantryItem{{1, "butter scotch", 1, time.Now().Truncate(time.Minute)}},
		},
		"simple hyphened add": {
			content:  "dry-gin",
			expected: []model.PantryItem{{1, "dry-gin", 1, time.Now().Truncate(time.Minute)}},
		},
		"add with trailing quantity": {
			content:  "bacon 5",
			expected: []model.PantryItem{{1, "bacon", 5, time.Now().Truncate(time.Minute)}},
		},
		"add with leading quantity": {
			content:  "13 bacon",
			expected: []model.PantryItem{{1, "bacon", 13, time.Now().Truncate(time.Minute)}},
		},
		"add with numbered name": {
			content:  "2 monkey47",
			expected: []model.PantryItem{{1, "monkey47", 2, time.Now().Truncate(time.Minute)}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// when
			actual := add([]model.PantryItem{}, test.content)

			// then
			assert.EqualValues(t, test.expected, actual)
		})
	}
}
