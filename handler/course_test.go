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

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestStore(t *testing.T) store.CourseStore {
	t.Helper()
	tmp, err := os.CreateTemp("", "duolabs-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmp.Close()
	t.Cleanup(func() { os.Remove(tmp.Name()) })

	db, err := gorm.Open(sqlite.Open(tmp.Name()), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Course{}); err != nil {
		t.Fatal(err)
	}
	return store.NewSQLiteCourseStore(db)
}

func TestGetByID_Found(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))
	body, _ := json.Marshal(model.Course{
		Title: "Hello World",
		Xp:    100,
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

	body, _ := json.Marshal(model.Course{Title: "Delete Me", Xp: 10})
	postReq := httptest.NewRequest(http.MethodPost, "/courses", bytes.NewReader(body))
	postW := httptest.NewRecorder()
	h.Create(postW, postReq)

	var created model.Course
	json.NewDecoder(postW.Body).Decode(&created)

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
	json.NewDecoder(w.Body).Decode(&courses)
	if len(courses) != 0 {
		t.Fatalf("expected empty array, got %d courses", len(courses))
	}
}

func TestGetCourses_ReturnsAll(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))

	for _, c := range []model.Course{
		{Title: "Go Basics", Xp: 100},
		{Title: "Advanced Go", Xp: 250},
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
	json.NewDecoder(w.Body).Decode(&courses)
	if len(courses) != 2 {
		t.Fatalf("expected 2 courses, got %d", len(courses))
	}
}

func TestGetCourses_OrderDesc(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))

	for _, title := range []string{"First", "Second", "Third"} {
		body, _ := json.Marshal(model.Course{Title: title, Xp: 10})
		req := httptest.NewRequest(http.MethodPost, "/courses", bytes.NewReader(body))
		w := httptest.NewRecorder()
		h.Create(w, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/courses", nil)
	w := httptest.NewRecorder()
	h.GetAll(w, req)

	var courses []model.Course
	json.NewDecoder(w.Body).Decode(&courses)
	if len(courses) != 3 {
		t.Fatalf("expected 3 courses, got %d", len(courses))
	}
	if courses[0].Title != "Third" {
		t.Fatalf("expected newest first, got %q", courses[0].Title)
	}
}

func TestPostCourse_Success(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))

	body, _ := json.Marshal(model.Course{Title: "Testing 101", Xp: 50})
	req := httptest.NewRequest(http.MethodPost, "/courses", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var course model.Course
	json.NewDecoder(w.Body).Decode(&course)
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

	body, _ := json.Marshal(model.Course{Title: "Integration", Xp: 999})
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
