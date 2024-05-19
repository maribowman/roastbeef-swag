package main

import (
	"context"
	"errors"
	"github.com/maribowman/roastbeef-swag/app"
	"github.com/maribowman/roastbeef-swag/app/config"
	"github.com/maribowman/roastbeef-swag/app/model"
	"github.com/maribowman/roastbeef-swag/app/repository"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var databaseClient model.DatabaseClient

func init() {
	initLogger()
	databaseClient = repository.NewDatabaseClient()
}

func initLogger() {
	var logger zerolog.Logger
	if config.Config.Logging.OutputFormat == "TEXT" {
		logFormat := zerolog.ConsoleWriter{Out: os.Stdout}
		logger = log.Output(logFormat).With().Timestamp().Caller().Logger()
	} else {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	}
	level, err := zerolog.ParseLevel(config.Config.Logging.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	log.Logger = logger
	log.Info().Msgf("Logging on %v level", level)
}

func main() {
	server, bot, err := app.InitServer(databaseClient)
	log.Info().Msgf("Running server on port %d", config.Config.Server.Port)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to init server")
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("Failed to boot server")
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	databaseClient.CloseDatabaseConnections()
	bot.CloseSession()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}
}
