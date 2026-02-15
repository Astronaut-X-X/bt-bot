package database

import (
	"bt-bot/database/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

type Config struct {
	Path  string `yaml:"path"`
	Debug bool   `yaml:"debug"`
}

var models = []any{
	&model.User{},
	&model.UserMap{},
	&model.Permissions{},
}

func InitDatabase(config Config) error {
	db, err := gorm.Open(sqlite.Open(config.Path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}
	DB = db

	return db.AutoMigrate(models...)
}
