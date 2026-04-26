package db

import (
	"errors"

	"github.com/lionellc/fusion-gate/internal/config"
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
		d := &DB{db: db}
		cleanup := func() {
			_ = d.Close()
		}
		return d, cleanup, nil
	default:
		return nil, nil, errors.New("invalid database type")

	}
}

func (d *DB) Close() error {
	return nil
}
