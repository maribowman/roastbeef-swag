package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/rs/zerolog/log"
)

type SqlitePantryClient struct {
	sqlite    *sql.DB
	tableName string
}

func NewSqlitePantryClient(databaseClient model.DatabaseClient, tableName string) model.PantryClient {
	client := &SqlitePantryClient{
		sqlite:    databaseClient.GetDatabaseConnection(),
		tableName: tableName,
	}
	client.init()
	return client
}

func (client *SqlitePantryClient) init() {
	_, err := client.sqlite.Exec(fmt.Sprintf("create table if not exists %s(id integer primary key autoincrement, name text not null, quantity int not null, date int not null);", client.tableName))
	if err != nil {
		log.Fatal().Err(err).Msgf("Could not create pantry database table %s", client.tableName)
	}
}

func (client *SqlitePantryClient) AddItem(item model.PantryItem) (int, error) {
	stmt, err := client.sqlite.Prepare(fmt.Sprintf("insert into %s (name, quantity, date) values (?, ?, ?);", client.tableName))
	if err != nil {
		log.Error().Err(err).Msgf("Failed to prepare insert statement on table %s", client.tableName)
		return -1, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(item.Name, item.Quantity, item.Date.Unix())
	if err != nil {
		log.Error().Err(err).Msgf("Failed to insert item [%+v] into table %s", item, client.tableName)
		return -1, err
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}

func (client *SqlitePantryClient) UpdateItem(item model.PantryItem) {
	stmt, err := client.sqlite.Prepare(fmt.Sprintf("update %s set name=?, quantity=? where id=?;", client.tableName))
	if err != nil {
		log.Error().Err(err).Msgf("Failed to prepare update statement on table %s", client.tableName)
		return
	}
	defer stmt.Close()

	if _, err := stmt.Exec(item.Name, item.Quantity, item.ID); err != nil {
		log.Error().Err(err).Msgf("Failed to update item [%+v] in table %s", item, client.tableName)
	}
}

func (client *SqlitePantryClient) RemoveItem(id int) {
	stmt, err := client.sqlite.Prepare(fmt.Sprintf("delete from %s where id=?;", client.tableName))
	if err != nil {
		log.Error().Err(err).Msgf("Failed to prepare delete statement on table %s", client.tableName)
		return
	}
	defer stmt.Close()

	if _, err = stmt.Exec(id); err != nil {
		log.Error().Err(err).Msgf("Failed to delete item with ID %d from table %s", id, client.tableName)
	}
}

func (client *SqlitePantryClient) RemoveAllItems() {
	// Remove all pantry items
	stmt, err := client.sqlite.Prepare(fmt.Sprintf("delete from %s;", client.tableName))
	if err != nil {
		log.Error().Err(err).Msgf("Failed to prepare delete all statement on table %s", client.tableName)
		return
	}
	defer stmt.Close()

	if _, err = stmt.Exec(); err != nil {
		log.Error().Err(err).Msgf("Failed to delete all items from table %s", client.tableName)
	}
	// Reset ID offset (mainly for unit testing)
	stmt, err = client.sqlite.Prepare(fmt.Sprintf("delete from sqlite_sequence where name='%s';", client.tableName))
	if err != nil {
		log.Error().Err(err).Msgf("Failed to prepare reset sequence statement on table %s", client.tableName)
		return
	}
	defer stmt.Close()

	if _, err = stmt.Exec(); err != nil {
		log.Error().Err(err).Msgf("Failed to reset sequence from table %s", client.tableName)
	}
}

func (client *SqlitePantryClient) GetItems() []model.PantryItem {
	stmt, err := client.sqlite.Prepare(fmt.Sprintf("select * from %s order by date asc;", client.tableName))
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
		err := rows.Scan(&item.ID, &item.Name, &item.Quantity, &unixDate)
		if err != nil {
			log.Error().Err(err).Msg("Failed to map row to pantry item")
		}
		item.Date = time.Unix(unixDate, 0)
		items = append(items, item)
	}
	return items
}
