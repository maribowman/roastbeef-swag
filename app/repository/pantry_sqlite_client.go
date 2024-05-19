package repository

import (
	"database/sql"
	"fmt"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/rs/zerolog/log"
	"time"
)

type PantrySqliteClient struct {
	sqlite    *sql.DB
	tableName string
}

func NewPantrySqliteClient(databaseClient model.DatabaseClient, tableName string) model.PantryClient {
	client := &PantrySqliteClient{
		sqlite:    databaseClient.GetDatabaseConnection(),
		tableName: tableName,
	}
	client.init()
	return client
}

func (client *PantrySqliteClient) init() {
	_, err := client.sqlite.Exec(fmt.Sprintf("create table if not exists %s(id integer primary key autoincrement, number integer not null unique, item text not null, amount int not null, date int not null);", client.tableName))
	if err != nil {
		log.Fatal().Err(err).Msgf("Could not create database pantry table %s", client.tableName)
	}
}

func (client *PantrySqliteClient) AddItem(item model.PantryItem) (int, error) {
	stmt, err := client.sqlite.Prepare("insert into ? values (?, ?, ?, ?);")
	if err != nil {
		log.Error().Err(err).Msgf("Failed to prepare insert statement on table %s", client.tableName)
		return -1, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(client.tableName, item.Number, item.Item, item.Amount, item.Date.Unix())
	if err != nil {
		log.Error().Err(err).Msgf("Failed to insert item [%s] into %s table", item.ToString(), client.tableName)
		return -1, err
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}

func (client *PantrySqliteClient) UpdateItem(item model.PantryItem) error {
	stmt, err := client.sqlite.Prepare("update ? set number=?, item=?, amount=? where id=?;")
	if err != nil {
		log.Error().Err(err).Msgf("Failed to prepare update statement on table %s", client.tableName)
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(client.tableName, item.Number, item.Item, item.Amount, item.ID); err != nil {
		log.Error().Err(err).Msgf("Failed to update item [%s] in %s table", item.ToString(), client.tableName)
		return err
	}
	return nil
}

func (client *PantrySqliteClient) RemoveItem(id int) error {
	stmt, err := client.sqlite.Prepare("delete from ? where id=?;")
	if err != nil {
		log.Error().Err(err).Msgf("Failed to prepare delete statement on table %s", client.tableName)
		return err
	}
	defer stmt.Close()

	if _, err = stmt.Exec(client.tableName, id); err != nil {
		log.Error().Err(err).Msgf("Failed to delete item [id: `%d`] in %s table", id, client.tableName)
	}
	return nil
}

func (client *PantrySqliteClient) GetItems() ([]model.PantryItem, error) {
	stmt, err := client.sqlite.Prepare("select * from ?;")
	if err != nil {
		log.Error().Err(err).Msgf("Failed to prepare select all statement on table %s", client.tableName)
		return []model.PantryItem{}, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(client.tableName)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to select all items from %s table", client.tableName)
		return []model.PantryItem{}, err
	}

	var items []model.PantryItem
	defer rows.Close()
	for rows.Next() {
		var item model.PantryItem
		var unixDate int64
		err := rows.Scan(&item.ID, &item.Number, &item.Item, &item.Amount, &unixDate)
		if err != nil {
			log.Error().Err(err).Msg("Failed to map row to pantry item")
		}
		item.Date = time.Unix(unixDate, 0)
		items = append(items, item)
	}
	return items, nil
}
