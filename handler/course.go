package handler

import (
	"encoding/json"
	"net/http"

	"duolabs/model"
	"duolabs/store"
)

type CourseHandler struct {
	Store store.CourseStore
}

func NewCourseHandler(s store.CourseStore) *CourseHandler {
	return &CourseHandler{Store: s}
}

func (h *CourseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getAll(w, r)
	case http.MethodPost:
		h.create(w, r)
	}
}

func (h *CourseHandler) getAll(w http.ResponseWriter, _ *http.Request) {
	courses, err := h.Store.FindAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(courses)
}

func (h *CourseHandler) create(w http.ResponseWriter, r *http.Request) {
	var course model.Course
	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := h.Store.Create(&course); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(course)
}
