package model

import "gorm.io/gorm"

type User struct {
	gorm.Model

	Username     string `json:"username" gorm:"type:varchar(50)"`
	AlwaysLunch  bool   `json:"always_lunch" gorm:"default:false"`
	AlwaysDinner bool   `json:"always_dinner" gorm:"default:false"`

	Reserves []Reserve `json:"reserves" gorm:"foreignKey:UserID"`
}
