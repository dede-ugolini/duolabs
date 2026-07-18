package handler

import (
	"encoding/json"
	"net/http"

	"duolabs/model"
	"duolabs/store"
)

type UserHandler struct {
	store store.UserStore
}

func NewUserHandler(s store.UserStore) *UserHandler {
	return &UserHandler{store: s}
}

func (h *UserHandler) GetAll(w http.ResponseWriter, _ *http.Request) {
	users, err := h.store.FindAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := h.store.Create(&user); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(user)
}
