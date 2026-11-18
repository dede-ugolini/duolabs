package handler

import (
	"errors"
	"net/http"

	"duolabs/store"

	"gorm.io/gorm"
)

type ProgressHandler struct {
	store store.ProgressStore
}

func NewProgressHandler(s store.ProgressStore) *ProgressHandler {
	return &ProgressHandler{store: s}
}

// Answer grades the choice the student picked for a lesson. The response only
// says whether it was right, never which choice is the correct one.
func (h *ProgressHandler) Answer(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}
	lessonID, ok := pathID(w, r, "id", "invalid lesson id")
	if !ok {
		return
	}
	var body struct {
		ChoiceID uint `json:"choice_id"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if body.ChoiceID == 0 {
		http.Error(w, "choice_id is required", http.StatusBadRequest)
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
		internalError(w, err)
	default:
		writeJSON(w, http.StatusOK, map[string]bool{"correct": correct})
	}
}

func (h *ProgressHandler) CourseProgress(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}
	courseID, ok := pathID(w, r, "id", "invalid course id")
	if !ok {
		return
	}
	progress, err := h.store.CourseProgress(userID, uint(courseID))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, progress)
}
