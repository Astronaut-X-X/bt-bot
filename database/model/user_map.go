package model

type UserMap struct {
	UserID int64  `gorm:"column:user_id;primaryKey"`
	UUID   string `gorm:"column:uuid;"`
}
