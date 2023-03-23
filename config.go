package main

import (
	"time"

	"github.com/jessevdk/go-flags"
	"golang.org/x/xerrors"
)

var cfg config

type config struct {
	BotToken          string        `long:"bot-token" env:"BOT_TOKEN"`
	LogLevel          string        `long:"log-level" default:"info" env:"LOG_LEVEL"`
	Debug             bool          `long:"debug" env:"DEBUG"`
	JSON              bool          `long:"json" env:"JSON" description:"write logs in json format"`
	DataReloadTimeout time.Duration `long:"data-reload-timeout" default:"30s" env:"DATA_RELOAD_TIMEOUT" description:"time between data reloads from google table"`
	PostgresURL       string        `long:"postgres-url" env:"POSTGRES_URL"`
	SpreadsheetID     string        `long:"spreadsheet-id" env:"SPREADSHEET_ID"`
}

func parseCfg() error {
	if _, err := flags.Parse(&cfg); err != nil {
		return xerrors.Errorf("parsing config: %w", err)
	}

	return nil
}
