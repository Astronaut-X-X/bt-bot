package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type UserIds []int64

var _ sql.Scanner = (*UserIds)(nil)
var _ driver.Valuer = (*UserIds)(nil)

func (a *UserIds) Scan(value any) error {
	value_, ok := value.(string)
	if !ok {
		return errors.New("type assertion to string failed")
	}
	return json.Unmarshal([]byte(value_), &a)
}

func (a *UserIds) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	data, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

type User struct {
	UUID     string  `gorm:"column:uuid;type:varchar(255);primaryKey"`
	UserIds  UserIds `gorm:"column:user_ids;type:text"`
	Premium  string  `gorm:"column:premium;type:varchar(255);default:basic"`
	Language string  `gorm:"column:language;type:varchar(255);default:zh"`
}
