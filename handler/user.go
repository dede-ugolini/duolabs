package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"duolabs/model"
	"duolabs/store"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandler struct {
	store store.UserStore
}

func NewUserHandler(s store.UserStore) *UserHandler {
	return &UserHandler{store: s}
}

// credentials is the body of both /register and /login.
type credentials struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

// Register creates a regular user. The role is never read from the request:
// admins only come from seed.EnsureAdmin.
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body credentials
	if !decodeJSON(w, r, &body) {
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" || body.Password == "" {
		http.Error(w, "name and password are required", http.StatusBadRequest)
		return
	}
	if _, err := h.store.FindByName(name); err == nil {
		http.Error(w, "name already taken", http.StatusConflict)
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		internalError(w, err)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		internalError(w, err)
		return
	}
	user := model.User{Name: name, Password: string(hash), Role: model.RoleUser}
	if err := h.store.Create(&user); err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

// Login checks the credentials and answers with a signed JWT.
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body credentials
	if !decodeJSON(w, r, &body) {
		return
	}
	user, err := h.store.FindByName(strings.TrimSpace(body.Name))
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)) != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	token, err := newToken(user)
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

// newToken signs a 24h JWT carrying the claims the middleware and the
// frontend read: user_id, name and role.
func newToken(user *model.User) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"name":    user.Name,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}).SignedString(jwtSecret)
}
