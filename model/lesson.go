package model

import "gorm.io/gorm"

type Lesson struct {
	gorm.Model
	Question string   `json:"question"`
	TaskID   uint     `json:"task_id"`
	Choices  []Choice `json:"choices" gorm:"foreignKey:LessonID"`
}
