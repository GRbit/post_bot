package main

import (
	"context"
	"os"
	"time"

	"github.com/grbit/post_bot/internal/bot"
	"github.com/grbit/post_bot/internal/config"
	"github.com/grbit/post_bot/internal/repo"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/xerrors"
)

func main() {
	ctx := context.TODO()
	cfg := config.Get()

	if err := configureLogging(cfg); err != nil {
		log.Panic().Err(err).Msgf("can't configure logging: %+v", err)
	}

	log.Info().Interface("config", cfg).Send()

	repo, err := db.InitDataUpdater(ctx, cfg.PostgresURL, cfg.SpreadsheetID)
	if err != nil {
		log.Panic().Err(err).Msgf("can't run data updater: %+v", err)
	}

	go repo.RunDataUpdater(ctx, cfg.DataReloadTimeout)

	b, err := bot.New(cfg)
	if err != nil {
		log.Panic().Err(err).Msgf("can't create bot: %+v", err)
	}

	for {
		if err := b.StartBot(); err != nil {
			log.Error().Err(err).Msgf("error from start bot function: %+v", err)
		}
	}
}

func configureLogging(cfg config.Values) error {
	log.Logger = log.
		With().Timestamp().
		Logger()

	lvl, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		return xerrors.Errorf("parsing log level: %w", err)
	}

	if cfg.Debug && lvl != zerolog.TraceLevel {
		lvl = zerolog.DebugLevel
	}

	log.Logger = log.Level(lvl)

	if !cfg.JSON {
		log.Logger = log.
			Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.StampMicro}).
			With().Caller().Timestamp().Logger()
		zerolog.TimeFieldFormat = time.StampMicro
	}

	log.Info().Msg("Logger configured")

	return nil
}
