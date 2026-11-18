package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"duolabs/model"
	"duolabs/store"

	"gorm.io/gorm"
)

const errInvalidDifficulty = "invalid difficulty: must be beginner, intermediate, or advanced"

type CourseHandler struct {
	Store store.CourseStore
}

func NewCourseHandler(s store.CourseStore) *CourseHandler {
	return &CourseHandler{Store: s}
}

func (h *CourseHandler) GetAll(w http.ResponseWriter, _ *http.Request) {
	courses, err := h.Store.FindAll()
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, courses)
}

func (h *CourseHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id", "invalid course id")
	if !ok {
		return
	}
	course, err := h.Store.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		internalError(w, err)
		return
	}
	if !IsAdminRequest(r) {
		hideAnswers(&course)
	}
	writeJSON(w, http.StatusOK, course)
}

func (h *CourseHandler) Create(w http.ResponseWriter, r *http.Request) {
	var course model.Course
	if !decodeJSON(w, r, &course) {
		return
	}
	if !course.Difficulty.Valid() {
		http.Error(w, errInvalidDifficulty, http.StatusBadRequest)
		return
	}
	if msg := validateLessons(&course); msg != "" {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	if err := h.Store.Create(&course); err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, course)
}

// Update changes the course's own fields (title, description, category,
// difficulty, xp). Fields left out of the body keep their current value.
func (h *CourseHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id", "invalid course id")
	if !ok {
		return
	}
	var course model.Course
	if !decodeJSON(w, r, &course) {
		return
	}
	if course.Difficulty != "" && !course.Difficulty.Valid() {
		http.Error(w, errInvalidDifficulty, http.StatusBadRequest)
		return
	}
	if err := h.Store.Update(&course, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		internalError(w, err)
		return
	}
	updated, err := h.Store.FindByID(id)
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *CourseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id", "invalid course id")
	if !ok {
		return
	}
	if err := h.Store.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		internalError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// hideAnswers removes the correct-answer markers so students cannot read them
// from the course payload.
func hideAnswers(course *model.Course) {
	for u := range course.Units {
		for t := range course.Units[u].Tasks {
			for l := range course.Units[u].Tasks[t].Lessons {
				choices := course.Units[u].Tasks[t].Lessons[l].Choices
				for c := range choices {
					choices[c].Correct = false
				}
			}
		}
	}
}

// validateLessons checks every lesson of the course and returns the first
// problem found, or "" when all of them can be graded.
func validateLessons(course *model.Course) string {
	for _, unit := range course.Units {
		for _, task := range unit.Tasks {
			for _, lesson := range task.Lessons {
				if msg := validateLesson(lesson); msg != "" {
					return msg
				}
			}
		}
	}
	return ""
}

// validateLesson requires a question, at least two choices with text and
// exactly one correct choice.
func validateLesson(lesson model.Lesson) string {
	if strings.TrimSpace(lesson.Question) == "" {
		return "invalid lesson: question is required"
	}
	prefix := "invalid lesson " + strconv.Quote(lesson.Question) + ": "
	if len(lesson.Choices) < 2 {
		return prefix + "at least 2 choices are required"
	}
	correct := 0
	for _, choice := range lesson.Choices {
		if strings.TrimSpace(choice.Label) == "" {
			return prefix + "choice label is required"
		}
		if choice.Correct {
			correct++
		}
	}
	if correct != 1 {
		return prefix + "exactly 1 correct choice is required"
	}
	return ""
}
