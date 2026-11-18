package handler

import (
	"errors"
	"net/http"

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
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}
	var body struct {
		CourseID uint `json:"course_id"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	enrollment := model.Enrollment{UserID: userID, CourseID: body.CourseID}
	if err := h.store.Enroll(&enrollment); err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			http.Error(w, "course not found", http.StatusNotFound)
		case errors.Is(err, gorm.ErrDuplicatedKey):
			http.Error(w, "already enrolled", http.StatusConflict)
		default:
			internalError(w, err)
		}
		return
	}
	writeJSON(w, http.StatusCreated, enrollment)
}

// Unenroll cancels an enrollment. Only its owner (or an admin) may do it.
func (h *EnrollmentHandler) Unenroll(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id", "invalid enrollment id")
	if !ok {
		return
	}
	enrollment, err := h.store.FindByID(uint(id))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		internalError(w, err)
		return
	}
	if enrollment.UserID != userID && RoleFromContext(r) != model.RoleAdmin {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err := h.store.Unenroll(enrollment.ID); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		internalError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *EnrollmentHandler) FindAll(w http.ResponseWriter, _ *http.Request) {
	enrollments, err := h.store.FindAll()
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, enrollments)
}

func (h *EnrollmentHandler) FindByUser(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "user_id", "invalid user id")
	if !ok {
		return
	}
	courses, err := h.store.FindByUser(uint(id))
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, courses)
}

func (h *EnrollmentHandler) FindByCourse(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "course_id", "invalid course id")
	if !ok {
		return
	}
	users, err := h.store.FindByCourse(uint(id))
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, users)
}
