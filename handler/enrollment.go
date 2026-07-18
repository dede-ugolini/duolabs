package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"duolabs/model"
	"duolabs/store"

	"gorm.io/gorm"
)

type EnrollmentHandler struct {
	store store.EnrollmentStore
}

func NewEnrollmentHandler(s store.EnrollmentStore) *EnrollmentHandler {
	return &EnrollmentHandler{store: s}
}

func (h *EnrollmentHandler) Enroll(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		CourseID uint `json:"course_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	enrollment := model.Enrollment{UserID: userID, CourseID: body.CourseID}
	if err := h.store.Enroll(&enrollment); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			http.Error(w, "already enrolled", http.StatusConflict)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(enrollment)
}

func (h *EnrollmentHandler) Unenroll(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if id <= 0 || err != nil {
		http.Error(w, "invalid enrollment id", http.StatusBadRequest)
		return
	}
	if err := h.store.Unenroll(uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *EnrollmentHandler) FindAll(w http.ResponseWriter, _ *http.Request) {
	enrollments, err := h.store.FindAll()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(enrollments)
}

func (h *EnrollmentHandler) FindByUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("user_id"))
	if id <= 0 || err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	courses, err := h.store.FindByUser(uint(id))
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(courses)
}

func (h *EnrollmentHandler) FindByCourse(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("course_id"))
	if id <= 0 || err != nil {
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}
	users, err := h.store.FindByCourse(uint(id))
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(users)
}
