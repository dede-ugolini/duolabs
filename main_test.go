package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *handler {
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
	if err := db.AutoMigrate(&Course{}); err != nil {
		t.Fatal(err)
	}
	return &handler{db: db}
}

func TestGetCourses_Empty(t *testing.T) {
	h := setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/courses", nil)
	w := httptest.NewRecorder()

	h.coursesHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var courses []Course
	_ = json.NewDecoder(w.Body).Decode(&courses)
	if len(courses) != 0 {
		t.Fatalf("expected empty array, got %d courses", len(courses))
	}
}

func TestGetCourses_ReturnsAll(t *testing.T) {
	h := setupTestDB(t)

	for _, c := range []Course{
		{Title: "Go Basics", Xp: 100},
		{Title: "Advanced Go", Xp: 250},
	} {
		body, _ := json.Marshal(c)
		req := httptest.NewRequest(http.MethodPost, "/courses", bytes.NewReader(body))
		w := httptest.NewRecorder()
		h.coursesHandler(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("POST failed: %d", w.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/courses", nil)
	w := httptest.NewRecorder()
	h.coursesHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var courses []Course
	json.NewDecoder(w.Body).Decode(&courses)
	if len(courses) != 2 {
		t.Fatalf("expected 2 courses, got %d", len(courses))
	}
}

func TestGetCourses_OrderDesc(t *testing.T) {
	h := setupTestDB(t)

	for _, title := range []string{"First", "Second", "Third"} {
		body, _ := json.Marshal(Course{Title: title, Xp: 10})
		req := httptest.NewRequest(http.MethodPost, "/courses", bytes.NewReader(body))
		w := httptest.NewRecorder()
		h.coursesHandler(w, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/courses", nil)
	w := httptest.NewRecorder()
	h.coursesHandler(w, req)

	var courses []Course
	json.NewDecoder(w.Body).Decode(&courses)
	if len(courses) != 3 {
		t.Fatalf("expected 3 courses, got %d", len(courses))
	}
	if courses[0].Title != "Third" {
		t.Fatalf("expected newest first, got %q", courses[0].Title)
	}
}

func TestPostCourse_Success(t *testing.T) {
	h := setupTestDB(t)

	body, _ := json.Marshal(Course{Title: "Testing 101", Xp: 50})
	req := httptest.NewRequest(http.MethodPost, "/courses", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.coursesHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var course Course
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
	h := setupTestDB(t)

	req := httptest.NewRequest(http.MethodPost, "/courses", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()

	h.coursesHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestPostThenGet(t *testing.T) {
	h := setupTestDB(t)

	body, _ := json.Marshal(Course{Title: "Integration", Xp: 999})
	postReq := httptest.NewRequest(http.MethodPost, "/courses", bytes.NewReader(body))
	postW := httptest.NewRecorder()
	h.coursesHandler(postW, postReq)

	if postW.Code != http.StatusCreated {
		t.Fatalf("POST failed: %d", postW.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/courses", nil)
	getW := httptest.NewRecorder()
	h.coursesHandler(getW, getReq)

	var courses []Course
	json.NewDecoder(getW.Body).Decode(&courses)
	if len(courses) != 1 {
		t.Fatalf("expected 1 course, got %d", len(courses))
	}
	if courses[0].Title != "Integration" {
		t.Fatalf("expected title %q, got %q", "Integration", courses[0].Title)
	}
}
