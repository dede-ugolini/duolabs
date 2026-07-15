package model

import "gorm.io/gorm"

type Course struct {
	gorm.Model
	Title string `json:"title"`
	Xp    int16  `json:"xp"`
}
