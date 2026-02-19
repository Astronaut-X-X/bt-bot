package database

import (
	"bt-bot/database/model"
	"errors"
	"fmt"
	"testing"

	"gorm.io/gorm"
)

func TestUser(t *testing.T) {
	InitDatabase(Config{
		Path:  "test.db",
		Debug: true,
	})

	var user model.User
	err := DB.Where("uuid = ?", "1234567890").First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Println("user not found")
			t.Fatal("user not found")
		}
	}

	fmt.Println(user)
}
