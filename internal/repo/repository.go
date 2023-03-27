package db

import (
	"context"
	"strings"
	"time"

	model2 "github.com/grbit/post_bot/internal/model"

	"github.com/samber/lo"
	"golang.org/x/xerrors"
	"google.golang.org/api/sheets/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repo struct {
	*gorm.DB
	sheets  *sheets.Service
	sheetID string
}

var globalRepo *Repo

const (
	poolSize        = 10
	connMaxLifetime = time.Minute
)

func initGlobalDB(ctx context.Context, postgresURL, spreadsheetID string) (*Repo, error) {
	pg := postgres.New(postgres.Config{DSN: postgresURL, PreferSimpleProtocol: true})
	db, err := gorm.Open(pg)
	if err != nil {
		return nil, xerrors.Errorf("opening form connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, xerrors.Errorf("getting sql.DB from gorm.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(poolSize)
	sqlDB.SetMaxIdleConns(poolSize)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	googleSheets, err := connectToGoogleSheetsService(ctx)
	if err != nil {
		return nil, xerrors.Errorf("connecting to google sheets: %w", err)
	}

	globalRepo = &Repo{
		DB:      db,
		sheets:  googleSheets,
		sheetID: spreadsheetID,
	}

	return globalRepo, nil
}

func (r *Repo) findUser(userID string) (*model2.User, error) {
	user := &model2.User{}
	if err := r.Take(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (r *Repo) upsertAddresses(addresses []*model2.Address) error {
	if len(addresses) == 0 {
		return nil
	}

	return r.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "telegram"}},
			Where:     clause.Where{Exprs: []clause.Expression{gorm.Expr("addresses.deleted_at IS NULL")}},
			UpdateAll: true,
		},
	).Create(addresses).Error
}

func (r *Repo) searchAddress(req string) ([]*model2.Address, error) {
	if req == "" {
		return nil, nil
	}

	phone := preparePhone(req)
	tg := prepareTelegram(req)
	inst := prepareInstagram(req)

	aa := []*model2.Address{}
	if err := r.Find(&aa,
		"phone = ? OR email = ? OR telegram = ? OR instagram = ?",
		phone, req, tg, inst).Error; err != nil {
		return nil, err
	}

	if len(aa) == 0 {
		if err := r.Find(&aa, "person_name ~* ?", req).Error; err != nil {
			return nil, err
		}

		lo.Filter(aa, func(a *model2.Address, _ int) bool {
			if a.PersonName == "" {
				return false
			}

			if strings.EqualFold(a.PersonName, req) {
				return true
			}

			ss := strings.Split(a.PersonName, " ")
			for _, s := range ss {
				if len(s) < 2 {
					continue
				}

				if strings.EqualFold(s, req) {
					return true
				}
			}

			return false
		})
	}

	return aa, nil
}
