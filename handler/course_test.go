package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"duolabs/handler"
	"duolabs/model"
	"duolabs/store"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestStore(t *testing.T) store.CourseStore {
	t.Helper()
	tmp, err := os.CreateTemp("", "duolabs-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	_ = tmp.Close()
	t.Cleanup(func() { _ = os.Remove(tmp.Name()) })

	db, err := gorm.Open(sqlite.Open(tmp.Name()), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Course{}, &model.Unit{}, &model.Task{}, &model.Lesson{}, &model.Choice{}); err != nil {
		t.Fatal(err)
	}
	return store.NewSQLiteCourseStore(db)
}

func TestGetByID_Found(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))
	body, _ := json.Marshal(model.Course{
		Title:      "Hello World",
		Difficulty: model.Beginner,
		Xp:         100,
	})
	h.Create(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/courses", bytes.NewReader(body)))

	req := httptest.NewRequest(http.MethodGet, "/courses/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.GetByID(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("Expected %d, got %d", http.StatusFound, w.Code)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))
	req := httptest.NewRequest(http.MethodGet, "/courses/9999", nil)
	req.SetPathValue("id", "9999")
	w := httptest.NewRecorder()
	h.GetByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetByID_BadID(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))
	req := httptest.NewRequest(http.MethodGet, "/courses/abc", nil)
	req.SetPathValue("id", "abc")
	w := httptest.NewRecorder()
	h.GetByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDelete_Found(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))

	body, _ := json.Marshal(model.Course{Title: "Delete Me", Difficulty: model.Beginner, Xp: 10})
	postReq := httptest.NewRequest(http.MethodPost, "/courses", bytes.NewReader(body))
	postW := httptest.NewRecorder()
	h.Create(postW, postReq)

	var created model.Course
	_ = json.NewDecoder(postW.Body).Decode(&created)

	delReq := httptest.NewRequest(http.MethodDelete, "/courses/"+strconv.Itoa(int(created.ID)), nil)
	delReq.SetPathValue("id", strconv.Itoa(int(created.ID)))
	delW := httptest.NewRecorder()
	h.Delete(delW, delReq)

	if delW.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", delW.Code)
	}
}

func TestDelete_NotFound(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))
	req := httptest.NewRequest(http.MethodDelete, "/courses/9999", nil)
	req.SetPathValue("id", "9999")
	w := httptest.NewRecorder()
	h.Delete(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestDelete_BadID(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))
	req := httptest.NewRequest(http.MethodDelete, "/courses/abc", nil)
	req.SetPathValue("id", "abc")
	w := httptest.NewRecorder()
	h.Delete(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetCourses_Empty(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))

	req := httptest.NewRequest(http.MethodGet, "/courses", nil)
	w := httptest.NewRecorder()
	h.GetAll(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var courses []model.Course
	_ = json.NewDecoder(w.Body).Decode(&courses)
	if len(courses) != 0 {
		t.Fatalf("expected empty array, got %d courses", len(courses))
	}
}

func TestGetCourses_ReturnsAll(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))

	for _, c := range []model.Course{
		{Title: "Go Basics", Difficulty: model.Beginner, Xp: 100},
		{Title: "Advanced Go", Difficulty: model.Advanced, Xp: 250},
	} {
		body, _ := json.Marshal(c)
		req := httptest.NewRequest(http.MethodPost, "/courses", bytes.NewReader(body))
		w := httptest.NewRecorder()
		h.Create(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("POST failed: %d", w.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/courses", nil)
	w := httptest.NewRecorder()
	h.GetAll(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var courses []model.Course
	_ = json.NewDecoder(w.Body).Decode(&courses)
	if len(courses) != 2 {
		t.Fatalf("expected 2 courses, got %d", len(courses))
	}
}

func TestGetCourses_OrderDesc(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))

	for _, title := range []string{"First", "Second", "Third"} {
		body, _ := json.Marshal(model.Course{Title: title, Difficulty: model.Intermediate, Xp: 10})
		req := httptest.NewRequest(http.MethodPost, "/courses", bytes.NewReader(body))
		w := httptest.NewRecorder()
		h.Create(w, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/courses", nil)
	w := httptest.NewRecorder()
	h.GetAll(w, req)

	var courses []model.Course
	_ = json.NewDecoder(w.Body).Decode(&courses)
	if len(courses) != 3 {
		t.Fatalf("expected 3 courses, got %d", len(courses))
	}
	if courses[0].Title != "Third" {
		t.Fatalf("expected newest first, got %q", courses[0].Title)
	}
}

func TestPostCourse_Success(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))

	body, _ := json.Marshal(model.Course{Title: "Testing 101", Difficulty: model.Beginner, Xp: 50})
	req := httptest.NewRequest(http.MethodPost, "/courses", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var course model.Course
	_ = json.NewDecoder(w.Body).Decode(&course)
	if course.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if course.Title != "Testing 101" {
		t.Fatalf("expected title %q, got %q", "Testing 101", course.Title)
	}
	if course.Xp != 50 {
		t.Fatalf("expected xp 50, got %d", course.Xp)
	}
}

func TestPostCourse_InvalidJSON(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))

	req := httptest.NewRequest(http.MethodPost, "/courses", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()
	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestPostThenGet(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))

	body, _ := json.Marshal(model.Course{Title: "Integration", Difficulty: model.Advanced, Xp: 999})
	postReq := httptest.NewRequest(http.MethodPost, "/courses", bytes.NewReader(body))
	postW := httptest.NewRecorder()
	h.Create(postW, postReq)

	if postW.Code != http.StatusCreated {
		t.Fatalf("POST failed: %d", postW.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/courses", nil)
	getW := httptest.NewRecorder()
	h.GetAll(getW, getReq)

	var courses []model.Course
	json.NewDecoder(getW.Body).Decode(&courses)
	if len(courses) != 1 {
		t.Fatalf("expected 1 course, got %d", len(courses))
	}
	if courses[0].Title != "Integration" {
		t.Fatalf("expected title %q, got %q", "Integration", courses[0].Title)
	}
}

func TestUpdate_Found(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))
	postBody, _ := json.Marshal(model.Course{Title: "Update Me", Difficulty: model.Beginner, Xp: 67})
	postReq := httptest.NewRequest(http.MethodPost, "/courses", bytes.NewReader(postBody))
	h.Create(httptest.NewRecorder(), postReq)

	putBody, _ := json.Marshal(model.Course{Title: "Update Found", Difficulty: model.Beginner, Xp: 67})
	putReq := httptest.NewRequest(http.MethodPut, "/courses/1", bytes.NewReader(putBody))
	putReq.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.Update(w, putReq)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected %d, got %d", http.StatusOK, w.Code)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))
	body, _ := json.Marshal(model.Course{Title: "Update Me", Difficulty: model.Beginner, Xp: 67})
	req := httptest.NewRequest(http.MethodPut, "/courses/1", bytes.NewReader(body))
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.Update(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestUpdate_InvalidJSON(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))
	postBody, _ := json.Marshal(model.Course{Title: "Update Me", Difficulty: model.Beginner, Xp: 67})
	postReq := httptest.NewRequest(http.MethodPost, "/courses", bytes.NewReader(postBody))
	h.Create(httptest.NewRecorder(), postReq)

	putReq := httptest.NewRequest(http.MethodPut, "/courses/1", bytes.NewReader([]byte("not json")))
	putReq.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.Update(w, putReq)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdate_BadID(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))
	body, _ := json.Marshal(model.Course{Title: "Update Me", Difficulty: model.Beginner, Xp: 67})
	req := httptest.NewRequest(http.MethodPut, "/courses/abc", bytes.NewReader(body))
	req.SetPathValue("id", "abc")
	w := httptest.NewRecorder()
	h.Update(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateCourse_Nested(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))

	course := model.Course{
		Title:       "Go Fundamentals",
		Description: "Learn Go from scratch",
		Category:    "programming language",
		Difficulty:  model.Beginner,
		Xp:          500,
		Units: []model.Unit{
			{
				Title: "Basics",
				Tasks: []model.Task{
					{
						Title: "Hello World",
						Lessons: []model.Lesson{
							{
								Question:        "What prints hello?",
								CorrectChoiceID: 1,
								Choices: []model.Choice{
									{Label: "fmt.Println"},
									{Label: "fmt.Print"},
									{Label: "println"},
								},
							},
						},
					},
				},
			},
		},
	}

	body, _ := json.Marshal(course)
	req := httptest.NewRequest(http.MethodPost, "/courses", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var created model.Course
	_ = json.NewDecoder(w.Body).Decode(&created)
	if created.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if created.Title != "Go Fundamentals" {
		t.Fatalf("expected title %q, got %q", "Go Fundamentals", created.Title)
	}
	if created.Category != "programming language" {
		t.Fatalf("expected category %q, got %q", "programming language", created.Category)
	}
	if created.Difficulty != model.Beginner {
		t.Fatalf("expected difficulty %q, got %q", model.Beginner, created.Difficulty)
	}
}

func TestCreateCourse_InvalidDifficulty(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))

	body, _ := json.Marshal(model.Course{Title: "Bad", Difficulty: "expert", Xp: 10})
	req := httptest.NewRequest(http.MethodPost, "/courses", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetCourse_Nested(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))

	course := model.Course{
		Title:       "Nested Course",
		Description: "Full tree",
		Category:    "back end",
		Difficulty:  model.Intermediate,
		Xp:          300,
		Units: []model.Unit{
			{
				Title: "Unit 1",
				Tasks: []model.Task{
					{
						Title: "Task 1",
						Lessons: []model.Lesson{
							{
								Question:        "Q1?",
								CorrectChoiceID: 1,
								Choices: []model.Choice{
									{Label: "A"},
									{Label: "B"},
								},
							},
						},
					},
				},
			},
		},
	}

	body, _ := json.Marshal(course)
	postReq := httptest.NewRequest(http.MethodPost, "/courses", bytes.NewReader(body))
	postW := httptest.NewRecorder()
	h.Create(postW, postReq)

	var created model.Course
	_ = json.NewDecoder(postW.Body).Decode(&created)

	getReq := httptest.NewRequest(http.MethodGet, "/courses/"+strconv.Itoa(int(created.ID)), nil)
	getReq.SetPathValue("id", strconv.Itoa(int(created.ID)))
	getW := httptest.NewRecorder()
	h.GetByID(getW, getReq)

	if getW.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", getW.Code)
	}

	var got model.Course
	_ = json.NewDecoder(getW.Body).Decode(&got)
	if len(got.Units) != 1 {
		t.Fatalf("expected 1 unit, got %d", len(got.Units))
	}
	if got.Units[0].Title != "Unit 1" {
		t.Fatalf("expected unit title %q, got %q", "Unit 1", got.Units[0].Title)
	}
	if len(got.Units[0].Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(got.Units[0].Tasks))
	}
	if len(got.Units[0].Tasks[0].Lessons) != 1 {
		t.Fatalf("expected 1 lesson, got %d", len(got.Units[0].Tasks[0].Lessons))
	}
	if len(got.Units[0].Tasks[0].Lessons[0].Choices) != 2 {
		t.Fatalf("expected 2 choices, got %d", len(got.Units[0].Tasks[0].Lessons[0].Choices))
	}
}
