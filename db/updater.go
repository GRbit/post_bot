package db

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"golang.org/x/xerrors"
	"github.com/grbit/post_bot/model"
	"math/rand"
)

func init() {
	cache.Persons = make(map[string]*model.Address)
}

type everyone struct {
	Persons map[string]*model.Address
	Tgs     []string
	sync.RWMutex
}

var cache everyone

func Search(req string) ([]string, error) {
	bb, err := globalRepo.searchAddress(req)
	if err != nil {
		return nil, xerrors.Errorf("searching in DB: %w", err)
	}

	var res []string
	for _, b := range bb {
		res = append(res, b.String())
	}

	return res, nil
}

func Random() string {
	cache.RLock()
	defer cache.RUnlock()

	log.Trace().
		Interface("telegrams", cache.Tgs).
		Interface("persons", cache.Persons).
		Msg("Random call")

	log.Debug().
		Interface("len(telegrams)", len(cache.Tgs)).
		Interface("len(persons)", len(cache.Persons)).
		Msg("Random call")

	k := rand.Intn(len(cache.Tgs))
	log.Debug().Str("random key", cache.Tgs[k]).Send()
	tg := cache.Tgs[k]
	log.Debug().Str("random telegram", tg).Send()

	return cache.Persons[tg].Address
}

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

	cache.Tgs = []string{}
	for i := range aa {
		cache.Tgs = append(cache.Tgs, aa[i].Telegram)
	}

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

func preparePhone(b string) string {
	l := len("79998148871")

	b = strings.ReplaceAll(b, `"`, ``)
	b = strings.ReplaceAll(b, " ", "")
	b = strings.ReplaceAll(b, "(", "")
	b = strings.ReplaceAll(b, ")", "")
	b = strings.ReplaceAll(b, "+", "")
	b = strings.ReplaceAll(b, "-", "")

	if len(b) == l && strings.HasPrefix(b, "89") {
		b = strings.Replace(b, "8", "7", 1)
	} else if len(b) == l-1 && strings.HasPrefix(b, "9") {
		b = "7" + b
	}

	return b
}
