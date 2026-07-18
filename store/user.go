package store

import (
	"duolabs/model"

	"gorm.io/gorm"
)

type UserStore interface {
	FindAll() ([]model.User, error)
	FindByName(name string) (*model.User, error)
	Create(*model.User) error
}

type SQLiteUserStore struct {
	db *gorm.DB
}

func NewSQLiteUserStore(db *gorm.DB) *SQLiteUserStore {
	return &SQLiteUserStore{db: db}
}

func (s *SQLiteUserStore) FindAll() ([]model.User, error) {
	var users []model.User
	err := s.db.Order("created_at desc").Find(&users).Error
	return users, err
}

func (s *SQLiteUserStore) FindByName(name string) (*model.User, error) {
	var user model.User
	err := s.db.Where("name = ?", name).First(&user).Error
	return &user, err
}

func (s *SQLiteUserStore) Create(user *model.User) error {
	return s.db.Create(user).Error
}
