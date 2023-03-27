package db

import (
	"context"
	"time"

	model2 "github.com/grbit/post_bot/internal/model"

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

	return r.
		Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "telegram"}},
				Where:     clause.Where{Exprs: []clause.Expression{gorm.Expr("addresses.deleted_at IS NULL")}},
				UpdateAll: true,
			},
		).
		Create(addresses).Error
}

func (r *Repo) searchAddress(req string) ([]*model2.Address, error) {
	phone := preparePhone(req)
	tg := prepareTelegram(req)

	aa := []*model2.Address{}
	if err := r.
		Find(&aa,
			"phone = ? OR email = ? OR telegram = ? OR instagram = ?",
			phone, req, tg, req).Error; err != nil {
		return nil, err
	}

	return aa, nil
}
