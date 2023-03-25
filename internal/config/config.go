package config

import (
	"sync"
	"time"

	"github.com/jessevdk/go-flags"
	"golang.org/x/xerrors"
)

var (
	cfg   Values
	parse sync.Once
)

type Values struct {
	BotToken          string        `long:"bot-token" env:"BOT_TOKEN"`
	LogLevel          string        `long:"log-level" default:"info" env:"LOG_LEVEL"`
	Debug             bool          `long:"debug" env:"DEBUG"`
	JSON              bool          `long:"json" env:"JSON" description:"write logs in json format"`
	DataReloadTimeout time.Duration `long:"data-reload-timeout" default:"30s" env:"DATA_RELOAD_TIMEOUT" description:"time between data reloads from google table"`
	PostgresURL       string        `long:"postgres-url" env:"POSTGRES_URL"`
	SpreadsheetID     string        `long:"spreadsheet-id" env:"SPREADSHEET_ID"`
}

func Get() Values {
	parse.Do(func() {
		if _, err := flags.Parse(&cfg); err != nil {
			panic(xerrors.Errorf("parsing Config: %w", err))
		}
	})

	return cfg
}
