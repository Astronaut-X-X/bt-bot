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

func ParseCallbackQueryUserId(update *tgbotapi.Update) int64 {
	callbackQuery := update.CallbackQuery
	if callbackQuery == nil {
		return 0
	}
	if callbackQuery.From == nil {
		return 0
	}
	return callbackQuery.From.ID
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

// 剩余每日下载数量
func RemainDailyDownloadQuantity(premium string) (int, error) {
	// 获取用户权限信息
	permissions, err := permissions(premium)
	if err != nil {
		return 0, err
	}

	// 取出上次记录的每日下载日期
	date := time.Unix(permissions.DailyDownloadDate, 0)
	// 如果不是今天，重置每日下载次数和日期
	if !date.Equal(time.Now().Truncate(24 * time.Hour)) {
		permissions.DailyDownloadRemain = permissions.DailyDownloadQuantity        // 重置剩余下载次数为每日最大下载数
		permissions.DailyDownloadDate = time.Now().Truncate(24 * time.Hour).Unix() // 更新为今天的日期
		if err = setPermissions(permissions); err != nil {
			return 0, err
		}
	}

	// 返回今天剩余的下载次数
	return permissions.DailyDownloadRemain, nil
}

// 减少每日下载数量
func DecrementDailyDownloadQuantity(premium string) error {
	permissions, err := permissions(premium)
	if err != nil {
		return err
	}

	permissions.DailyDownloadRemain--
	if permissions.DailyDownloadRemain < 0 {
		permissions.DailyDownloadRemain = 0
	}

	if err = setPermissions(permissions); err != nil {
		return err
	}
	return nil
}

var (
	SetPermissionsLock sync.Mutex
)

// 设置权限
func setPermissions(permissions *model.Permissions) error {
	SetPermissionsLock.Lock()
	defer SetPermissionsLock.Unlock()
	return database.DB.Save(permissions).Error
}
