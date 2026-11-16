package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"duolabs/store"

	"gorm.io/gorm"
)

type ProgressHandler struct {
	store store.ProgressStore
}

func NewProgressHandler(s store.ProgressStore) *ProgressHandler {
	return &ProgressHandler{store: s}
}

func (h *ProgressHandler) Answer(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	lessonID, err := strconv.Atoi(r.PathValue("id"))
	if lessonID <= 0 || err != nil {
		http.Error(w, "invalid lesson id", http.StatusBadRequest)
		return
	}
	var body struct {
		ChoiceID uint `json:"choice_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ChoiceID == 0 {
		http.Error(w, "invalid JSON: choice_id is required", http.StatusBadRequest)
		return
	}
	correct, err := h.store.SubmitAnswer(userID, uint(lessonID), body.ChoiceID)
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		http.Error(w, "lesson not found", http.StatusNotFound)
	case errors.Is(err, store.ErrNotEnrolled):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, store.ErrInvalidChoice):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case err != nil:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	default:
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"correct": correct})
	}
}

func (h *ProgressHandler) CourseProgress(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	courseID, err := strconv.Atoi(r.PathValue("id"))
	if courseID <= 0 || err != nil {
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}
	progress, err := h.store.CourseProgress(userID, uint(courseID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(progress)
}
