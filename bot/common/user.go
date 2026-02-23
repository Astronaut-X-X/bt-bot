package common

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"bt-bot/database"
	"bt-bot/database/model"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func ParseFullName(update *tgbotapi.Update) string {
	user := update.Message.From
	if user.LastName != "" {
		return fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	}
	return user.FirstName
}

func ParseCallbackQueryFullName(update *tgbotapi.Update) string {
	user := update.CallbackQuery.From
	if user.LastName == "" {
		return fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	}
	return user.FirstName
}

func ParseUserId(update *tgbotapi.Update) int64 {
	msg := update.Message
	if msg == nil {
		return 0
	}
	return msg.From.ID
}

func User(userID int64) (*model.User, error) {
	uuid, ok, err := userUUID(userID)
	if !ok {
		user, _, err := CreateUserPermissions(userID)
		if err != nil {
			return nil, err
		}
		return user, nil
	}
	if err != nil {
		return nil, err
	}

	user, err := user(uuid)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func Permissions(userID int64) (*model.Permissions, error) {
	uuid, ok, err := userUUID(userID)
	if !ok {
		_, permissions, err := CreateUserPermissions(userID)
		if err != nil {
			return nil, err
		}
		return permissions, nil
	}
	if err != nil {
		return nil, err
	}

	user, err := user(uuid)
	if err != nil {
		return nil, err
	}

	permissions, err := permissions(user.Premium)
	if err != nil {
		return nil, err
	}

	return permissions, nil
}

func userUUID(userID int64) (string, bool, error) {
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

func user(uuid string) (*model.User, error) {
	var user model.User
	err := database.DB.Where("uuid = ?", uuid).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func permissions(premium string) (*model.Permissions, error) {
	var permissions model.Permissions
	err := database.DB.Where("uuid = ?", premium).First(&permissions).Error
	if err != nil {
		return nil, err
	}

	return &permissions, nil
}

func CreateUserPermissions(userID int64) (*model.User, *model.Permissions, error) {
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

// 减少用户可下载文件数
func DecrementDownloadCount(premium string) (bool, error) {
	permissions, err := permissions(premium)
	if err != nil {
		return false, err
	}

	if permissions.AsyncDownloadRemain <= 0 {
		return false, errors.New("下载数量不足")
	}

	permissions.AsyncDownloadRemain = permissions.AsyncDownloadRemain - 1
	if err = SetPermissions(premium, permissions); err != nil {
		return false, err
	}

	return true, nil
}

// 下载完成后恢复可下载文件数
func IncrementDownloadCount(premium string) error {
	permissions, err := permissions(premium)
	if err != nil {
		return err
	}

	permissions.AsyncDownloadRemain = min(permissions.AsyncDownloadRemain+1, permissions.AsyncDownloadQuantity)
	if err = SetPermissions(premium, permissions); err != nil {
		return err
	}
	return nil
}

func RemainDailyDownload(premium string) (bool, error) {
	permissions, err := permissions(premium)
	if err != nil {
		return false, err
	}

	date := time.Unix(permissions.DailyDownloadDate, 0)
	if !date.Equal(time.Now().Truncate(24 * time.Hour)) {
		permissions.DailyDownloadRemain = permissions.DailyDownloadQuantity
		permissions.DailyDownloadDate = time.Now().Truncate(24 * time.Hour).Unix()
		if err = SetPermissions(premium, permissions); err != nil {
			return false, err
		}
	}

	if permissions.DailyDownloadRemain <= 0 {
		return false, nil
	}
	return true, nil
}

func DecrementDailyDownloadQuantity(premium string) error {
	permissions, err := permissions(premium)
	if err != nil {
		return err
	}

	permissions.DailyDownloadRemain = min(permissions.DailyDownloadRemain-1, 0)
	if err = SetPermissions(premium, permissions); err != nil {
		return err
	}
	return nil
}

var (
	SetPermissionsLock sync.Mutex
)

func SetPermissions(premium string, permissions *model.Permissions) error {
	SetPermissionsLock.Lock()
	defer SetPermissionsLock.Unlock()
	return database.DB.Save(permissions).Error
}
