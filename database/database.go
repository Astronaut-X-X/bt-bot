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
	&model.TorrentInfo{},
	&model.TorrentFile{},
	&model.DownloadFileMessage{},
	&model.DownloadFileComment{},
}

func InitDatabase(config Config) error {
	db, err := gorm.Open(sqlite.Open(config.Path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}
	DB = db

	err = db.AutoMigrate(models...)
	if err != nil {
		return err
	}

	ResetPermissions()

	return nil
}

func ResetPermissions() {
	DB.Model(&model.Permissions{}).Where("type = ?", model.PermissionsTypeBasic).Update("async_download_remain", 1)
	DB.Model(&model.Permissions{}).Where("type = ?", model.PermissionsTypePremium).Update("async_download_remain", 3)
}
