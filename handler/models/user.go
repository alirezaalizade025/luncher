package model

import "gorm.io/gorm"

type User struct {
	gorm.Model

	Name         string `json:"name" gorm:"type:varchar(50)"`
	Username     string `json:"username" gorm:"type:varchar(50)"`
	TelegramID   int64  `json:"telegram_id" gorm:"unique"`
	AlwaysLunch  bool   `json:"always_lunch" gorm:"default:false"`
	AlwaysDinner bool   `json:"always_dinner" gorm:"default:false"`

	Reserves []Reserve `json:"reserves" gorm:"foreignKey:UserID"`
}
