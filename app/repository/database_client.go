package repository

import (
	"database/sql"
	"github.com/maribowman/roastbeef-swag/app/config"
	"github.com/maribowman/roastbeef-swag/app/model"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
	"os"
)

type DatabaseClient struct {
	sqlite *sql.DB
}

func NewDatabaseClient() model.DatabaseClient {
	return &DatabaseClient{
		sqlite: initSqliteConnection(),
	}
}

func initSqliteConnection() *sql.DB {
	_, err := os.Stat(config.Config.Database.Sqlite)
	if os.IsNotExist(err) {
		log.Info().Msg("No sqlite file present -> creating one")
		_ = os.MkdirAll(config.Config.Database.Sqlite, os.ModePerm)
		if file, err := os.Create(config.Config.Database.Sqlite); err != nil {
			log.Fatal().Err(err).Msg("Could not create sqlite file")
		} else {
			defer file.Close()
		}
	}

	log.Debug().Msg("Opening sqlite connection")
	sqlite, err := sql.Open("sqlite3", config.Config.Database.Sqlite)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not open sqlite")
	}

	log.Debug().Msg("Testing sqlite connection")
	err = sqlite.Ping()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not establish connection to sqlite")
	}

	return sqlite
}

func (client *DatabaseClient) CloseDatabaseConnections() {
	if err := client.sqlite.Close(); err != nil {
		log.Warn().Err(err).Msg("Unable to close database connection")
	}
}
