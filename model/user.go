package model

import "gorm.io/gorm"

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

type User struct {
	gorm.Model
	Name     string `json:"name"`
	Password string `json:"-"`
	Role     string `json:"role" gorm:"default:user"`
}
