package db

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grbit/post_bot/internal/model"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"golang.org/x/xerrors"
	"google.golang.org/api/sheets/v4"
)

// range concepts https://developers.google.com/sheets/api/guides/concepts#expandable-1
const (
	page      = "Лист1"
	readRange = page + "!A:Z"
)

func init() {
	cache.Persons = make(map[string]*model.Address)
}

type everyone struct {
	Persons map[string]*model.Address
	sync.RWMutex
}

func (e *everyone) Add(a *model.Address) {
	e.Lock()
	defer e.Unlock()

	e.Persons[a.Telegram] = a
}

var cache everyone // almost not used, probably better remove

func InitDataUpdater(
	ctx context.Context,
	postgresURL string,
	spID string,
) (*Repo, error) {
	repo, err := initGlobalDB(ctx, postgresURL, spID)
	if err != nil {
		return nil, xerrors.Errorf("init global DB: %w", err)
	}

	if err := repo.updateData(ctx); err != nil {
		return nil, xerrors.Errorf("update data: %w", err)
	}

	log.Debug().Msg("InitDataUpdater done")

	return repo, nil
}

func (r *Repo) RunDataUpdater(
	ctx context.Context,
	dataReloadTimeout time.Duration,
) {
	for range time.NewTicker(dataReloadTimeout).C {
		if err := r.updateData(ctx); err != nil {
			log.Error().Err(err).Msg("updating data")
		}
	}
}

func (r *Repo) updateData(ctx context.Context) error {
	aa, err := r.getAddressesFromGoogle(ctx)
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

	log.Debug().Msg("cache updated")

	return nil
}

func (r *Repo) getAddressesFromGoogle(ctx context.Context) ([]*model.Address, error) {
	resp, err := r.sheets.Spreadsheets.Values.Get(globalRepo.sheetID, readRange).Do()
	if err != nil {
		return nil, xerrors.Errorf("retrieving data from sheet %q: %w", globalRepo.sheetID, err)
	}

	if len(resp.Values) == 0 {
		return nil, xerrors.Errorf("no data found in spreadsheet %q", globalRepo.sheetID)
	}

	persons := []*model.Address{}

	for _, rowColumns := range resp.Values {
		row := make([]string, len(rowColumns))
		for i, column := range rowColumns {
			row[i] = column.(string)
		}

		if len(row) < 2 {
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
			case 5:
				a.Approved = parseBool(row[i])
			}
		}

		persons = append(persons, &a)

		log.Trace().
			Interface("row", row).
			Interface("person", a).
			Msg("Row added")
	}

	log.Info().Int("persons", len(persons)).Msg("persons readed")

	return persons, nil
}

var trueValues = []string{"1", "да", "Да", "ДА", "ок", "Ок", "ОК", "ок", "ОК", "Ок", "true", "True", "TRUE", "yes", "Yes", "YES"}

func parseBool(s string) bool {
	b, _ := strconv.ParseBool(s)

	return b || lo.Contains(trueValues, strings.TrimSpace(s))
}

func (r *Repo) pushAddressToGoogleTable(a *model.Address) error {
	resp, err := r.sheets.Spreadsheets.Values.Get(globalRepo.sheetID, readRange).Do()
	if err != nil {
		return xerrors.Errorf("retrieving data from sheet %q: %w", globalRepo.sheetID, err)
	}

	if len(resp.Values) == 0 {
		return xerrors.Errorf("no data found in spreadsheet %q", globalRepo.sheetID)
	}

	for i, rowColumns := range resp.Values {
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

		if row[0] == a.Telegram {
			log.Debug().Interface("row", row).Interface("telegram", a.Telegram).Msg("row found, updating")

			// updating all the data
			vr := &sheets.ValueRange{
				Values: [][]interface{}{
					{
						a.Telegram,
						a.Instagram,
						a.PersonName,
						a.Address,
						a.Wishes,
					},
				},
			}

			updRange := fmt.Sprintf("%s!%d:%d", page, i+1, i+1)

			_, err = r.sheets.Spreadsheets.Values.Update(globalRepo.sheetID, updRange, vr).
				ValueInputOption("USER_ENTERED").Do()
			if err != nil {
				return xerrors.Errorf("updating data to sheet %q: %w", globalRepo.sheetID, err)
			}

			return nil
		}
	}

	log.Debug().Interface("telegram", a.Telegram).Msg("row not found, appending")

	// append to google table
	vr := &sheets.ValueRange{
		Values: [][]interface{}{
			{
				a.Telegram,
				a.Instagram,
				a.PersonName,
				a.Address,
				a.Wishes,
			},
		},
	}

	_, err = r.sheets.Spreadsheets.Values.Append(globalRepo.sheetID, readRange, vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return xerrors.Errorf("appending data to sheet %q: %w", globalRepo.sheetID, err)
	}

	log.Debug().Interface("telegram", a.Telegram).Msg("row appended")

	return nil
}
