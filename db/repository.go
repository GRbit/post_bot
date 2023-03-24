package db

import (
	"time"

	"golang.org/x/xerrors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"github.com/grbit/post_bot/model"
)

type Repo struct {
	*gorm.DB
}

var globalRepo *Repo

const (
	poolSize        = 10
	connMaxLifetime = time.Minute
)

func initGlobalDB(postgresURL string) error {
	db, err := ConnectDB(postgresURL, poolSize, connMaxLifetime)
	if err != nil {
		return xerrors.Errorf("connecting to DB: %w", err)
	}

	globalRepo = &Repo{DB: db}

	return nil
}

func (r *Repo) findUser(userID string) (*model.User, error) {
	user := &model.User{}
	if err := r.Take(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (r *Repo) searchAddress(req string) ([]*model.Address, error) {
	phone := preparePhone(req)
	tg := prepareTelegram(req)

	aa := []*model.Address{}
	if err := r.
		Find(&aa,
			"phone = ? OR email = ? OR telegram = ? OR instagram = ?",
			phone, req, tg, req).Error; err != nil {
		return nil, err
	}

	return aa, nil
}

func (r *Repo) upsertAddresses(addresses []*model.Address) error {
	if len(addresses) == 0 {
		return nil
	}

	return r.
		Clauses(
			clause.OnConflict{
				Columns: []clause.Column{{Name: "telegram"}},
				Where:   clause.Where{Exprs: []clause.Expression{gorm.Expr("addresses.deleted_at IS NULL")}},
				DoUpdates: clause.AssignmentColumns([]string{
					"instagram",
					"person_name",
					"address",
					"wishes",
					"updated_at",
				}),
			},
		).
		Create(addresses).Error
}

func ConnectDB(url string, poolSize int, connMaxLifetime time.Duration) (*gorm.DB, error) {
	pg := postgres.New(postgres.Config{DSN: url, PreferSimpleProtocol: true})
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

	return db, nil
}
