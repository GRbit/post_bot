package db

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/grbit/post_bot/model"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"golang.org/x/xerrors"
)

func init() {
	cache.Persons = make(map[string]*model.Address)
}

type everyone struct {
	Persons map[string]*model.Address
	Tgs     []string
	sync.RWMutex
}

func (e *everyone) Add(a *model.Address) {
	e.Lock()
	defer e.Unlock()

	e.Persons[a.Telegram] = a
	e.Tgs = append(e.Tgs, a.Telegram)
	e.Tgs = lo.Uniq(e.Tgs)
}

var cache everyone // almost not used, probably better remove

func InitDataUpdater(
	ctx context.Context,
	postgresURL string,
	spID string,
) error {
	if err := initGlobalDB(postgresURL); err != nil {
		return xerrors.Errorf("init global DB: %w", err)
	}

	if err := updateData(ctx, spID); err != nil {
		return xerrors.Errorf("update data: %w", err)
	}

	log.Debug().Msg("InitDataUpdater done")

	return nil
}

func RunDataUpdater(
	ctx context.Context,
	dataReloadTimeout time.Duration,
	spID string,
) {
	for range time.NewTicker(dataReloadTimeout).C {
		if err := updateData(ctx, spID); err != nil {
			log.Error().Err(err).Msg("updating data")
		}
	}
}

func updateData(ctx context.Context, spID string) error {
	aa, err := getAddresses(ctx, spID)
	if err != nil {
		return xerrors.Errorf("getting addresses: %w", err)
	}

	if err = globalRepo.upsertAddresses(aa); err != nil {
		return xerrors.Errorf("upserting addresses: %w", err)
	}

	log.Debug().Interface("lenAddresses", len(aa)).Msg("Addresses upserted")

	cache.Lock()
	defer cache.Unlock()

	for i := range aa {
		cache.Persons[aa[i].Telegram] = aa[i]
	}

	cache.Tgs = lo.Map(aa, func(a *model.Address, _ int) string {
		return a.Telegram
	})

	log.Debug().Msg("cache updated")

	return nil
}

func getAddresses(ctx context.Context, spreadsheetID string) ([]*model.Address, error) {
	srv, err := newSheetsService(ctx)
	if err != nil {
		return nil, xerrors.Errorf("creating new spreadsheets client: %w", err)
	}

	// range concepts https://developers.google.com/sheets/api/guides/concepts#expandable-1
	readRange := "Лист1!A:Z"

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, xerrors.Errorf("retrieving data from sheet %q: %w", spreadsheetID, err)
	}

	if len(resp.Values) == 0 {
		return nil, xerrors.Errorf("no data found in spreadsheet %q", spreadsheetID)
	}

	persons := []*model.Address{}

	for _, rowColumns := range resp.Values {
		row := make([]string, len(rowColumns))
		for i, column := range rowColumns {
			row[i] = column.(string)
		}

		if len(row) < 4 {
			log.Debug().Interface("row", row).Msg("Skipping too short row")

			continue
		}

		if lo.Contains(row, "Телеграм") {
			// Телеграм	Инстаграмм	Имя и фамилия	Адрес	Пожелания
			log.Debug().Interface("row", row).Msg("Skipping main row")

			continue
		}

		a := model.Address{}

		for i := range row {
			switch i {
			case 0:
				a.Telegram = strings.ToLower(row[i])
			case 1:
				a.Instagram = strings.ToLower(row[i])
			case 2:
				a.PersonName = row[i]
			case 3:
				a.Address = row[i]
			case 4:
				a.Wishes = row[i]
			}
		}

		persons = append(persons, &a)

		log.Trace().
			Interface("row", row).
			Interface("person", persons[len(persons)-1]).
			Msg("Row added")
	}

	log.Info().Int("persons", len(persons)).Msg("persons readed")

	return persons, nil
}

func prepareBuilding(b string) string {
	b = strings.ReplaceAll(b, " ", "")
	b = strings.ReplaceAll(b, "Зг", "ЗГ")
	b = regexp.MustCompile("([^0-9])([0-9])-([0-9])").ReplaceAllString(b, "${1}0${2}-${3}")

	return b
}
