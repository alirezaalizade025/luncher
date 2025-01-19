package model

type Meal struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	Lunch  *string `json:"lunch" gorm:"type:varchar(50)"`
	Dinner *string `json:"dinner" gorm:"type:varchar(50)"`
}
