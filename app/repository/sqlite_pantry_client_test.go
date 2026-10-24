package repository

import (
	"testing"
	"time"

	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/stretchr/testify/assert"
)

func TestGetItemsOrdersTiedDatesByID(t *testing.T) {
	// given
	pantryClient := NewSqlitePantryClient(NewDatabaseClient(), "repository_order_tests")
	defer pantryClient.RemoveAllItems()
	date := time.Now().Truncate(time.Minute)
	for _, name := range []string{"milk", "eggs", "bread"} {
		_, err := pantryClient.AddItem(model.PantryItem{Name: name, Quantity: 1, Date: date})
		assert.NoError(t, err)
	}

	// when
	actual := pantryClient.GetItems()

	// then
	assert.Equal(t, []model.PantryItem{
		{ID: 1, Name: "milk", Quantity: 1, Date: date},
		{ID: 2, Name: "eggs", Quantity: 1, Date: date},
		{ID: 3, Name: "bread", Quantity: 1, Date: date},
	}, actual)
}

func TestSnapshotRestore(t *testing.T) {
	date := time.Now().Truncate(time.Minute)
	original := []model.PantryItem{
		{Name: "bacon", Quantity: 2, Date: date.Add(-2 * time.Hour)},
		{Name: "coffee", Quantity: 1, Date: date.Add(-time.Hour)},
		{Name: "milk", Quantity: 3, Date: date},
	}
	restoredOriginal := []model.PantryItem{
		{ID: 1, Name: "bacon", Quantity: 2, Date: date.Add(-2 * time.Hour)},
		{ID: 2, Name: "coffee", Quantity: 1, Date: date.Add(-time.Hour)},
		{ID: 3, Name: "milk", Quantity: 3, Date: date},
	}

	// where
	tests := map[string]struct {
		initial  []model.PantryItem
		mutate   func(model.PantryClient)
		expected []model.PantryItem
	}{
		"restore after add": {
			initial: original,
			mutate: func(client model.PantryClient) {
				_, _ = client.AddItem(model.PantryItem{Name: "eggs", Quantity: 6, Date: date})
			},
			expected: restoredOriginal,
		},
		"restore after remove": {
			initial: original,
			mutate: func(client model.PantryClient) {
				client.RemoveItem(2)
			},
			expected: restoredOriginal,
		},
		"restore after update": {
			initial: original,
			mutate: func(client model.PantryClient) {
				client.UpdateItem(model.PantryItem{ID: 3, Name: "oat milk", Quantity: 1})
			},
			expected: restoredOriginal,
		},
		"restore after remove all": {
			initial: original,
			mutate: func(client model.PantryClient) {
				client.RemoveAllItems()
			},
			expected: restoredOriginal,
		},
		"restore snapshot of empty table": {
			initial: nil,
			mutate: func(client model.PantryClient) {
				_, _ = client.AddItem(model.PantryItem{Name: "eggs", Quantity: 6, Date: date})
			},
			expected: nil,
		},
	}

	// given
	pantryClient := NewSqlitePantryClient(NewDatabaseClient(), "repository_snapshot_tests")

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// and
			for _, item := range test.initial {
				_, err := pantryClient.AddItem(item)
				assert.NoError(t, err)
			}
			pantryClient.CreateSnapshot()
			test.mutate(pantryClient)

			// when
			restored := pantryClient.RestoreSnapshot()

			// then
			assert.True(t, restored)
			assert.Equal(t, test.expected, pantryClient.GetItems())

			// and a second restore changes nothing
			assert.False(t, pantryClient.RestoreSnapshot())
			assert.Equal(t, test.expected, pantryClient.GetItems())

			// cleanup
			pantryClient.RemoveAllItems()
		})
	}
}

func TestRestoreSnapshotWithoutSnapshot(t *testing.T) {
	// given
	pantryClient := NewSqlitePantryClient(NewDatabaseClient(), "repository_no_snapshot_tests")
	defer pantryClient.RemoveAllItems()
	date := time.Now().Truncate(time.Minute)
	_, err := pantryClient.AddItem(model.PantryItem{Name: "bacon", Quantity: 1, Date: date})
	assert.NoError(t, err)

	// when
	restored := pantryClient.RestoreSnapshot()

	// then
	assert.False(t, restored)
	assert.Equal(t, []model.PantryItem{{ID: 1, Name: "bacon", Quantity: 1, Date: date}}, pantryClient.GetItems())
}

func TestSnapshotsAreIsolatedPerTable(t *testing.T) {
	// given
	databaseClient := NewDatabaseClient()
	groceries := NewSqlitePantryClient(databaseClient, "repository_snapshot_groceries")
	freezer := NewSqlitePantryClient(databaseClient, "repository_snapshot_freezer")
	defer groceries.RemoveAllItems()
	defer freezer.RemoveAllItems()
	date := time.Now().Truncate(time.Minute)

	// and
	_, _ = groceries.AddItem(model.PantryItem{Name: "milk", Quantity: 1, Date: date})
	_, _ = freezer.AddItem(model.PantryItem{Name: "peas", Quantity: 2, Date: date})
	groceries.CreateSnapshot()
	freezer.CreateSnapshot()
	_, _ = groceries.AddItem(model.PantryItem{Name: "eggs", Quantity: 6, Date: date})
	freezer.RemoveAllItems()

	// when
	restoredGroceries := groceries.RestoreSnapshot()

	// then
	assert.True(t, restoredGroceries)
	assert.Equal(t, []model.PantryItem{{ID: 1, Name: "milk", Quantity: 1, Date: date}}, groceries.GetItems())
	assert.Empty(t, freezer.GetItems())

	// and when
	restoredFreezer := freezer.RestoreSnapshot()

	// then
	assert.True(t, restoredFreezer)
	assert.Equal(t, []model.PantryItem{{ID: 1, Name: "peas", Quantity: 2, Date: date}}, freezer.GetItems())
	assert.Equal(t, []model.PantryItem{{ID: 1, Name: "milk", Quantity: 1, Date: date}}, groceries.GetItems())
}
