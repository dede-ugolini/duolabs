package model

import "gorm.io/gorm"

type Enrollment struct {
	gorm.Model
	UserID   uint `json:"user_id"`
	CourseID uint `json:"course_id"`
}
