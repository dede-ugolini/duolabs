package model

import "gorm.io/gorm"

type Course struct {
	gorm.Model
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	Difficulty  Difficulty `json:"difficulty"`
	Xp          int16      `json:"xp"`
	Units       []Unit     `json:"units" gorm:"foreignKey:CourseID"`
}
