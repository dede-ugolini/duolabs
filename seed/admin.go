package seed

import (
	"errors"

	"duolabs/model"
	"duolabs/store"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// EnsureAdmin creates the admin account, or resets an existing account with
// that name to the given password and the admin role. Run on every startup:
// without the reset, whoever registered the admin name first would keep it.
func EnsureAdmin(s store.UserStore, name, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user, err := s.FindByName(name)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.Create(&model.User{Name: name, Password: string(hash), Role: model.RoleAdmin})
	}
	if err != nil {
		return err
	}
	user.Password = string(hash)
	user.Role = model.RoleAdmin
	return s.Save(user)
}
