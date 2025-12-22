package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/rs/zerolog/log"
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
	_, err := client.sqlite.Exec(fmt.Sprintf("create table if not exists %s(id integer primary key autoincrement, name text not null, amount int not null, date int not null);", client.tableName))
	if err != nil {
		log.Fatal().Err(err).Msgf("Could not create pantry database table %s", client.tableName)
	}
}

func (client *PantrySqliteClient) AddItem(item model.PantryItem) (int, error) {
	stmt, err := client.sqlite.Prepare(fmt.Sprintf("insert into %s (name, amount, date) values (?, ?, ?);", client.tableName))
	if err != nil {
		log.Error().Err(err).Msgf("Failed to prepare insert statement on table %s", client.tableName)
		return -1, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(item.Name, item.Amount, item.Date.Unix())
	if err != nil {
		log.Error().Err(err).Msgf("Failed to insert item [%+v] into table %s", item, client.tableName)
		return -1, err
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}

func (client *PantrySqliteClient) UpdateItem(item model.PantryItem) {
	stmt, err := client.sqlite.Prepare(fmt.Sprintf("update %s set name=?, amount=? where index=?;", client.tableName))
	if err != nil {
		log.Error().Err(err).Msgf("Failed to prepare update statement on table %s", client.tableName)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(item.Name, item.Amount, item.ID); err != nil {
		log.Error().Err(err).Msgf("Failed to update item [%+v] in table %s", item, client.tableName)
	}
}

func (client *PantrySqliteClient) RemoveItem(id int) {
	stmt, err := client.sqlite.Prepare(fmt.Sprintf("delete from %s where id=?;", client.tableName))
	if err != nil {
		log.Error().Err(err).Msgf("Failed to prepare delete statement on table %s", client.tableName)
	}
	defer stmt.Close()

	if _, err = stmt.Exec(id); err != nil {
		log.Error().Err(err).Msgf("Failed to delete item with ID %d from table %s", id, client.tableName)
	}
}

func (client *PantrySqliteClient) GetItems() []model.PantryItem {
	stmt, err := client.sqlite.Prepare(fmt.Sprintf("select * from %s;", client.tableName))
	if err != nil {
		log.Error().Err(err).Msgf("Failed to prepare select all statement on table %s", client.tableName)
		return []model.PantryItem{}
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		log.Error().Err(err).Msgf("Failed to select all items from table %s", client.tableName)
		return []model.PantryItem{}
	}

	var items []model.PantryItem
	defer rows.Close()
	for rows.Next() {
		var item model.PantryItem
		var unixDate int64
		err := rows.Scan(&item.Name, &item.Amount, &unixDate)
		if err != nil {
			log.Error().Err(err).Msg("Failed to map row to pantry item")
		}
		item.Date = time.Unix(unixDate, 0)
		items = append(items, item)
	}
	return items
}
