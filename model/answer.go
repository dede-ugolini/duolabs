package model

import "gorm.io/gorm"

// Answer is a user's latest attempt at a lesson. Once Correct is true it
// stays true: the lesson counts as completed.
type Answer struct {
	gorm.Model
	UserID   uint `json:"user_id" gorm:"uniqueIndex:idx_answer_user_lesson"`
	LessonID uint `json:"lesson_id" gorm:"uniqueIndex:idx_answer_user_lesson"`
	ChoiceID uint `json:"choice_id"`
	Correct  bool `json:"correct"`
}

type CourseProgress struct {
	CourseID           uint   `json:"course_id"`
	TotalLessons       int    `json:"total_lessons"`
	CompletedLessons   int    `json:"completed_lessons"`
	Completed          bool   `json:"completed"`
	XpEarned           int    `json:"xp_earned"`
	CompletedLessonIDs []uint `json:"completed_lesson_ids"`
}
