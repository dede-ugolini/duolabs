package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

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
	if !IsAdminRequest(r) {
		hideAnswers(&course)
	}
	_ = json.NewEncoder(w).Encode(course)
}

// hideAnswers removes the correct-answer markers so students cannot read them
// from the course payload.
func hideAnswers(course *model.Course) {
	for u := range course.Units {
		for t := range course.Units[u].Tasks {
			for l := range course.Units[u].Tasks[t].Lessons {
				lesson := &course.Units[u].Tasks[t].Lessons[l]
				lesson.CorrectChoiceID = 0
				for c := range lesson.Choices {
					lesson.Choices[c].Correct = false
				}
			}
		}
	}
}

// validateLessons requires every lesson to have a question, at least two
// choices and exactly one correct choice, so that it can always be graded.
func validateLessons(course *model.Course) string {
	for _, unit := range course.Units {
		for _, task := range unit.Tasks {
			for _, lesson := range task.Lessons {
				if strings.TrimSpace(lesson.Question) == "" {
					return "invalid lesson: question is required"
				}
				if len(lesson.Choices) < 2 {
					return "invalid lesson " + strconv.Quote(lesson.Question) + ": at least 2 choices are required"
				}
				correct := 0
				for _, choice := range lesson.Choices {
					if strings.TrimSpace(choice.Label) == "" {
						return "invalid lesson " + strconv.Quote(lesson.Question) + ": choice label is required"
					}
					if choice.Correct {
						correct++
					}
				}
				if correct != 1 {
					return "invalid lesson " + strconv.Quote(lesson.Question) + ": exactly 1 correct choice is required"
				}
			}
		}
	}
	return ""
}

func (h *CourseHandler) Create(w http.ResponseWriter, r *http.Request) {
	var course model.Course
	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if course.Difficulty != model.Beginner && course.Difficulty != model.Intermediate && course.Difficulty != model.Advanced {
		http.Error(w, "invalid difficulty: must be beginner, intermediate, or advanced", http.StatusBadRequest)
		return
	}
	if msg := validateLessons(&course); msg != "" {
		http.Error(w, msg, http.StatusBadRequest)
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
