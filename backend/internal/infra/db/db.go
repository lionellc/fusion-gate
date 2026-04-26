package db

import (
	"context"
	"errors"

	"github.com/lionellc/fusion-gate/internal/config"
	apikeyentity "github.com/lionellc/fusion-gate/internal/domain/apikey/entity"
	userentity "github.com/lionellc/fusion-gate/internal/domain/user/entity"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DB struct {
	db *gorm.DB
}

func NewDB(cfg *config.Config) (*DB, func(), error) {
	switch cfg.Database.Type {
	case "sqlite":
		db, err := gorm.Open(sqlite.Open("fusion.db"), &gorm.Config{})
		if err != nil {
			return nil, nil, err
		}

		db.AutoMigrate(&userentity.User{}, &apikeyentity.APIKey{})

		d := &DB{db: db}
		cleanup := func() {
			_ = d.Close()
		}
		return d, cleanup, nil
	default:
		return nil, nil, errors.New("invalid database type")

	}
}

func (d *DB) WithContext(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx)
}

func (d *DB) Close() error {
	return nil
}
