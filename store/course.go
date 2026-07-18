// Package store provides the store
package store

import (
	"duolabs/model"

	"gorm.io/gorm"
)

type CourseStore interface {
	FindAll() ([]model.Course, error)
	FindByID(id int) (model.Course, error)
	Create(course *model.Course) error
	Delete(id int) error
	Update(course *model.Course, id int) error
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

func (s *SQLiteCourseStore) FindByID(id int) (model.Course, error) {
	var course model.Course
	err := s.db.Preload("Units.Tasks.Lessons.Choices").First(&course, id).Error
	return course, err
}

func (s *SQLiteCourseStore) Create(course *model.Course) error {
	return s.db.Create(course).Error
}

func (s *SQLiteCourseStore) Delete(id int) error {
	result := s.db.Delete(&model.Course{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *SQLiteCourseStore) Update(course *model.Course, id int) error {
	result := s.db.Model(&model.Course{}).Where("id = ?", id).Updates(course)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
