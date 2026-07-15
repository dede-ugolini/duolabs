package store

import (
	"duolabs/model"

	"gorm.io/gorm"
)

type CourseStore interface {
	FindAll() ([]model.Course, error)
	Create(course *model.Course) error
}

type SQLiteCourseStore struct {
	db *gorm.DB
}

func NewSQLiteCourseStore(db *gorm.DB) *SQLiteCourseStore {
	return &SQLiteCourseStore{db: db}
}

func (s *SQLiteCourseStore) FindAll() ([]model.Course, error) {
	var courses []model.Course
	err := s.db.Order("created_at desc").Find(&courses).Error
	return courses, err
}

func (s *SQLiteCourseStore) Create(course *model.Course) error {
	return s.db.Create(course).Error
}
