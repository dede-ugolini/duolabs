package store

import (
	"duolabs/model"

	"gorm.io/gorm"
)

type UserStore interface {
	FindByName(name string) (*model.User, error)
	Create(*model.User) error
	Save(*model.User) error
}

type SQLiteUserStore struct {
	db *gorm.DB
}

func NewSQLiteUserStore(db *gorm.DB) *SQLiteUserStore {
	return &SQLiteUserStore{db: db}
}

func (s *SQLiteUserStore) FindByName(name string) (*model.User, error) {
	var user model.User
	err := s.db.Where("name = ?", name).First(&user).Error
	return &user, err
}

func (s *SQLiteUserStore) Create(user *model.User) error {
	return s.db.Create(user).Error
}

func (s *SQLiteUserStore) Save(user *model.User) error {
	return s.db.Save(user).Error
}
