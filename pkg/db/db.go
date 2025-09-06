package db

import (
	"fmt"
	"telecommunications_repair_hub/config"
	"telecommunications_repair_hub/models"

	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
	config *config.Config
}

func New(config *config.Config) (*DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		config.Database.Host, config.Database.User, config.Database.Password, config.Database.Database, config.Database.Port)
	dbInstance, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, errors.WithMessage(err, "failed to open database")
	}

	db := &DB{
		DB:     dbInstance,
		config: config,
	}

	if err := db.Migrate(); err != nil {
		return nil, errors.WithMessage(err, "failed to migrate database")
	}

	return db, nil
}

func (d *DB) Migrate() error {
	return d.AutoMigrate(
		&models.User{},
	)
}
