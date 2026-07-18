package model

import "gorm.io/gorm"

type Task struct {
	gorm.Model
	Title   string   `json:"title"`
	UnitID  uint     `json:"unit_id"`
	Lessons []Lesson `json:"lessons" gorm:"foreignKey:TaskID"`
}
