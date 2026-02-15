package model

type User struct {
	UUID     string `gorm:"column:uuid;primaryKey"`
	UserIds  string `gorm:"column:user_ids"`
	Premium  string `gorm:"column:premium;default:basic"`
	Language string `gorm:"column:language;default:zh"`
}
