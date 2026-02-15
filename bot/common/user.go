package common

import (
	"bt-bot/database"
	"bt-bot/database/model"
	"errors"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func FullName(user *tgbotapi.User) string {
	if user.LastName == "" {
		return fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	}
	return user.FirstName
}

func GetUserAndPermissions(userID int64) (*model.User, *model.Permissions, error) {
	var userMap model.UserMap
	err := database.DB.Where("user_id = ?", userID).First(&userMap).Error
	if err != nil {
		return nil, nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return CreateUser(userID)
	}

	var user model.User
	err = database.DB.Where("uuid = ?", userMap.UUID).First(&user).Error
	if err != nil {
		return nil, nil, err
	}

	var permissions model.Permissions
	err = database.DB.Where("uuid = ?", user.Premium).First(&permissions).Error
	if err != nil {
		return nil, nil, err
	}

	return &user, &permissions, nil
}

func CreateUser(userID int64) (*model.User, *model.Permissions, error) {
	permissionsUUID := uuid.New().String()
	permissions := model.BasicPermissions
	permissions.UUID = permissionsUUID

	userUUID := uuid.New().String()
	user := model.User{
		UUID:     userUUID,
		UserIds:  model.UserIds{userID},
		Premium:  permissionsUUID,
		Language: "zh",
	}

	userMap := model.UserMap{
		UserID: userID,
		UUID:   userUUID,
	}

	err := database.DB.Create(&permissions).Error
	if err != nil {
		return nil, nil, err
	}
	err = database.DB.Create(&user).Error
	if err != nil {
		return nil, nil, err
	}
	err = database.DB.Create(&userMap).Error
	if err != nil {
		return nil, nil, err
	}
	return &user, &permissions, nil
}
