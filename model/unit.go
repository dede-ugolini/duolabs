package model

import "gorm.io/gorm"

type Unit struct {
	gorm.Model
	Title    string `json:"title"`
	CourseID uint   `json:"course_id"`
	Tasks    []Task `json:"tasks" gorm:"foreignKey:UnitID"`
}
