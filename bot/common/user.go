package common

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"bt-bot/database"
	"bt-bot/database/model"

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

func UserAndPermissions(userID int64) (*model.User, *model.Permissions, error) {
	uuid, ok, err := GetUserUUID(userID)
	if !ok {
		user, permissions, err := CreateUser(userID)
		if err != nil {
			return nil, nil, err
		}
		return user, permissions, nil
	}
	// 其他错误
	if err != nil {
		return nil, nil, err
	}

	user, err := GetUser(uuid)
	if err != nil {
		return nil, nil, err
	}
	permissions, err := GetPermissions(user.Premium)
	if err != nil {
		return nil, nil, err
	}

	return user, permissions, nil

}

func GetUserUUID(userID int64) (string, bool, error) {
	var userMap model.UserMap
	err := database.DB.Where("user_id = ?", userID).First(&userMap).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", true, err
	}
	return userMap.UUID, true, nil
}

func GetUser(uuid string) (*model.User, error) {
	var user model.User
	err := database.DB.Where("uuid = ?", uuid).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func GetPermissions(premium string) (*model.Permissions, error) {
	var permissions model.Permissions
	err := database.DB.Where("uuid = ?", premium).First(&permissions).Error
	if err != nil {
		return nil, err
	}

	return &permissions, nil
}

func CreateUser(userID int64) (*model.User, *model.Permissions, error) {
	permissionsUUID := uuid.New().String()
	permissions := model.BasicPermissions
	permissions.UUID = permissionsUUID

	userUUID := uuid.New().String()
	user := model.User{
		UUID:     userUUID,
		UserIds:  strings.Join([]string{strconv.FormatInt(userID, 10)}, ","),
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
