package store

import (
	"errors"

	"duolabs/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrNotEnrolled   = errors.New("not enrolled in this course")
	ErrInvalidChoice = errors.New("choice does not belong to this lesson")
)

type ProgressStore interface {
	// SubmitAnswer grades the answer and records it. It returns
	// gorm.ErrRecordNotFound for an unknown lesson, ErrNotEnrolled when the
	// user is not enrolled in the lesson's course and ErrInvalidChoice when
	// the choice is not one of the lesson's choices.
	SubmitAnswer(userID, lessonID, choiceID uint) (correct bool, err error)
	// CourseProgress returns gorm.ErrRecordNotFound for an unknown course.
	CourseProgress(userID, courseID uint) (model.CourseProgress, error)
}

type SQLiteProgressStore struct {
	db *gorm.DB
}

func NewSQLiteProgressStore(db *gorm.DB) *SQLiteProgressStore {
	return &SQLiteProgressStore{db: db}
}

// lessonsOfCourse selects the live lessons that belong to a course.
func (s *SQLiteProgressStore) lessonsOfCourse(courseID uint) *gorm.DB {
	return s.db.Table("lessons").
		Joins("JOIN tasks ON tasks.id = lessons.task_id AND tasks.deleted_at IS NULL").
		Joins("JOIN units ON units.id = tasks.unit_id AND units.deleted_at IS NULL").
		Where("units.course_id = ? AND lessons.deleted_at IS NULL", courseID)
}

func (s *SQLiteProgressStore) SubmitAnswer(userID, lessonID, choiceID uint) (bool, error) {
	var courseID uint
	err := s.db.Table("lessons").
		Select("units.course_id").
		Joins("JOIN tasks ON tasks.id = lessons.task_id AND tasks.deleted_at IS NULL").
		Joins("JOIN units ON units.id = tasks.unit_id AND units.deleted_at IS NULL").
		Where("lessons.id = ? AND lessons.deleted_at IS NULL", lessonID).
		Scan(&courseID).Error
	if err != nil {
		return false, err
	}
	if courseID == 0 {
		return false, gorm.ErrRecordNotFound
	}

	var enrolled int64
	if err := s.db.Model(&model.Enrollment{}).
		Where("user_id = ? AND course_id = ?", userID, courseID).
		Count(&enrolled).Error; err != nil {
		return false, err
	}
	if enrolled == 0 {
		return false, ErrNotEnrolled
	}

	var choice model.Choice
	if err := s.db.Where("id = ? AND lesson_id = ?", choiceID, lessonID).First(&choice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ErrInvalidChoice
		}
		return false, err
	}

	// Single atomic upsert: concurrent submits of the same lesson cannot
	// collide on the unique (user_id, lesson_id) index, and a correct answer
	// is never overwritten by a later attempt.
	answer := model.Answer{UserID: userID, LessonID: lessonID, ChoiceID: choiceID, Correct: choice.Correct}
	err = s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "lesson_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"choice_id":  gorm.Expr("CASE WHEN answers.correct THEN answers.choice_id ELSE excluded.choice_id END"),
			"correct":    gorm.Expr("answers.correct OR excluded.correct"),
			"updated_at": gorm.Expr("excluded.updated_at"),
		}),
	}).Create(&answer).Error
	if err != nil {
		return false, err
	}

	var stored model.Answer
	if err := s.db.Where("user_id = ? AND lesson_id = ?", userID, lessonID).First(&stored).Error; err != nil {
		return false, err
	}
	return stored.Correct, nil
}

func (s *SQLiteProgressStore) CourseProgress(userID, courseID uint) (model.CourseProgress, error) {
	var course model.Course
	if err := s.db.Select("id", "xp").First(&course, courseID).Error; err != nil {
		return model.CourseProgress{}, err
	}

	progress := model.CourseProgress{CourseID: courseID, CompletedLessonIDs: []uint{}}

	var total int64
	if err := s.lessonsOfCourse(courseID).Count(&total).Error; err != nil {
		return progress, err
	}
	progress.TotalLessons = int(total)

	if err := s.lessonsOfCourse(courseID).
		Joins("JOIN answers ON answers.lesson_id = lessons.id AND answers.deleted_at IS NULL").
		Where("answers.user_id = ? AND answers.correct = ?", userID, true).
		Order("lessons.id").
		Pluck("lessons.id", &progress.CompletedLessonIDs).Error; err != nil {
		return progress, err
	}
	progress.CompletedLessons = len(progress.CompletedLessonIDs)
	progress.Completed = progress.TotalLessons > 0 && progress.CompletedLessons == progress.TotalLessons
	if progress.Completed {
		progress.XpEarned = int(course.Xp)
	}
	return progress, nil
}
