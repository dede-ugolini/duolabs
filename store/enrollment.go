package store

import (
	"duolabs/model"

	"gorm.io/gorm"
)

type EnrollmentStore interface {
	Enroll(enrollment *model.Enrollment) error
	Unenroll(id uint) error
	FindByUser(userID uint) ([]model.Course, error)
	FindByCourse(courseID uint) ([]model.User, error)
	FindAll() ([]model.Enrollment, error)
}

type SQLiteEnrollmentStore struct {
	db *gorm.DB
}

func NewSQLiteEnrollmentStore(db *gorm.DB) *SQLiteEnrollmentStore {
	return &SQLiteEnrollmentStore{db: db}
}

func (s *SQLiteEnrollmentStore) Enroll(enrollment *model.Enrollment) error {
	var existing model.Enrollment
	err := s.db.Where("user_id = ? AND course_id = ?", enrollment.UserID, enrollment.CourseID).First(&existing).Error
	if err == nil {
		return gorm.ErrDuplicatedKey
	}
	return s.db.Create(enrollment).Error
}

func (s *SQLiteEnrollmentStore) Unenroll(id uint) error {
	result := s.db.Delete(&model.Enrollment{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *SQLiteEnrollmentStore) FindByUser(userID uint) ([]model.Course, error) {
	var courses []model.Course
	err := s.db.Joins("JOIN enrollments ON enrollments.course_id = courses.id").
		Where("enrollments.user_id = ?", userID).
		Order("courses.created_at desc").
		Find(&courses).Error
	return courses, err
}

func (s *SQLiteEnrollmentStore) FindByCourse(courseID uint) ([]model.User, error) {
	var users []model.User
	err := s.db.Joins("JOIN enrollments ON enrollments.user_id = users.id").
		Where("enrollments.course_id = ?", courseID).
		Order("users.created_at desc").
		Find(&users).Error
	return users, err
}

func (s *SQLiteEnrollmentStore) FindAll() ([]model.Enrollment, error) {
	var enrollments []model.Enrollment
	err := s.db.Order("created_at desc").Find(&enrollments).Error
	return enrollments, err
}
