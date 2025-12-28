package repository

import (
	"database/sql"
	"github.com/maribowman/roastbeef-swag/app/config"
	"github.com/maribowman/roastbeef-swag/app/model"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
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
	if config.Config.Database.Sqlite == ":memory:" {
		log.Info().Msg("Connecting to configured in-memory sqlite")
	} else if _, err := os.Stat(config.Config.Database.Sqlite); os.IsNotExist(err) {
		log.Info().Msg("No sqlite file present -> creating one")
		_ = os.MkdirAll(filepath.Dir(config.Config.Database.Sqlite), 0755)
		if file, err := os.Create(config.Config.Database.Sqlite); err != nil {
			log.Fatal().Err(err).Msg("Could not create sqlite file")
		} else {
			defer file.Close()
		}
	} else {
		log.Info().Msg("Sqlite file exists, using existing one")
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

func (client *DatabaseClient) GetDatabaseConnection() *sql.DB {
	return client.sqlite
}

func (client *DatabaseClient) CloseDatabaseConnection() {
	if err := client.sqlite.Close(); err != nil {
		log.Warn().Err(err).Msg("Unable to close database connection")
	}
}
