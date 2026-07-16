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

type CourseHandler struct {
	Store store.CourseStore
}

func NewCourseHandler(s store.CourseStore) *CourseHandler {
	return &CourseHandler{Store: s}
}

func (h *CourseHandler) GetAll(w http.ResponseWriter, _ *http.Request) {
	courses, err := h.Store.FindAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(courses)
}

func (h *CourseHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if id == 0 || err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	course, err := h.Store.FindByID(id)
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusFound)
	_ = json.NewEncoder(w).Encode(course)
}

func (h *CourseHandler) Create(w http.ResponseWriter, r *http.Request) {
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

func (h *CourseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if id <= 0 || err != nil {
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}
	err = h.Store.Delete(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CourseHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if id <= 0 || err != nil {
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}
	var course model.Course
	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := h.Store.Update(&course, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}
	_ = json.NewEncoder(w).Encode(course)
}
