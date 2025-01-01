package model

import "time"

type Reserve struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Date      time.Time `json:"date" gorm:"not null"`
	UserID    uint      `json:"user_id"`
	HasLunch  bool      `json:"has_lunch" gorm:"default:false"`
	HasDinner bool      `json:"has_dinner" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
	UpdateAt  time.Time `json:"update_at"`

	User User `json:"user" gorm:"foreignKey:UserID"`
}
