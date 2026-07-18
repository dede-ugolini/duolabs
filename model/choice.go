package model

import "gorm.io/gorm"

type Choice struct {
	gorm.Model
	Label    string `json:"label"`
	LessonID uint   `json:"lesson_id"`
}
