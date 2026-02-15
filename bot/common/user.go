package common

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"bt-bot/database"
	"bt-bot/database/model"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

func FullName(user *tgbotapi.User) string {
	if user.LastName == "" {
		return fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	}
	return user.FirstName
}

func GetUserUUID(userID int64) (string, bool, error) {
	var userMap model.UserMap
	err := database.DB.Where("user_id = ?", userID).First(&userMap).Error
	if err != nil {
		log.Println("GetUserUUID error:", err)
		return "", false, err
	}
	return userMap.UUID, true, nil
}

func GetUserAndPermissions(UUID string) (*model.User, *model.Permissions, error) {
	var user model.User
	err := database.DB.Where("uuid = ?", UUID).First(&user).Error
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
