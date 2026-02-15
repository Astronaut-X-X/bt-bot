package model

type UserMap struct {
	UserID int64  `gorm:"column:user_id;type:int64;primaryKey"`
	UUID   string `gorm:"column:uuid;type:varchar(255)"`
}
