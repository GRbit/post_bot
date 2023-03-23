package main

import (
	"context"
	"os"
	"time"

	"github.com/grbit/post_bot/db"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/xerrors"
)

const (
	senMsgRetries = 10
)

func main() {
	ctx := context.TODO()

	if err := parseCfg(); err != nil {
		log.Panic().Err(err).Msgf("can't parse config: %+v", err)
	}

	if err := configureLogging(); err != nil {
		log.Panic().Err(err).Msgf("can't configure logging: %+v", err)
	}

	log.Info().Interface("config", cfg).Send()

	if err := db.InitDataUpdater(ctx, cfg.PostgresURL, cfg.SpreadsheetID); err != nil {
		log.Panic().Err(err).Msgf("can't run data updater: %+v", err)
	}

	go db.RunDataUpdater(ctx, cfg.DataReloadTimeout, cfg.SpreadsheetID)

	bot, err := newBot()
	if err != nil {
		log.Panic().Err(err).Msgf("can't create bot: %+v", err)
	}

	bot.configure(makeHandlers())

	for {
		if err := bot.startBot(); err != nil {
			log.Error().Err(err).Msgf("error from start bot function: %+v", err)
		}
	}
}

func configureLogging() error {
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
