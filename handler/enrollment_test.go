package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"duolabs/handler"
	"duolabs/model"
	"duolabs/store"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var testJWTSecret = []byte("duolabs-secret")

func setupTestEnrollmentStore(t *testing.T) store.EnrollmentStore {
	t.Helper()
	tmp, err := os.CreateTemp("", "duolabs-enrollment-*.db")
	if err != nil {
		t.Fatal(err)
	}
	_ = tmp.Close()
	t.Cleanup(func() { os.Remove(tmp.Name()) })

	db, err := gorm.Open(sqlite.Open(tmp.Name()), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Course{}, &model.User{}, &model.Enrollment{}); err != nil {
		t.Fatal(err)
	}
	return store.NewSQLiteEnrollmentStore(db)
}

func setupTestEnrollmentDB(t *testing.T) (store.EnrollmentStore, *gorm.DB) {
	t.Helper()
	tmp, err := os.CreateTemp("", "duolabs-enrollment-*.db")
	if err != nil {
		t.Fatal(err)
	}
	_ = tmp.Close()
	t.Cleanup(func() { os.Remove(tmp.Name()) })

	db, err := gorm.Open(sqlite.Open(tmp.Name()), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Course{}, &model.User{}, &model.Enrollment{}); err != nil {
		t.Fatal(err)
	}
	return store.NewSQLiteEnrollmentStore(db), db
}

func generateToken(t *testing.T, userID uint, name string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"name":    name,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(testJWTSecret)
	if err != nil {
		t.Fatal(err)
	}
	return tokenString
}

func TestEnroll_Success(t *testing.T) {
	s, db := setupTestEnrollmentDB(t)
	user := model.User{Name: "alice", Password: "hash"}
	db.Create(&user)
	course := model.Course{Title: "Go 101", Xp: 100}
	db.Create(&course)

	h := handler.NewEnrollmentHandler(s)
	body, _ := json.Marshal(map[string]uint{"course_id": course.ID})
	req := httptest.NewRequest(http.MethodPost, "/enrollments", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+generateToken(t, user.ID, user.Name))
	w := httptest.NewRecorder()
	handler.AuthMiddleware(h.Enroll)(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var enrollment model.Enrollment
	json.NewDecoder(w.Body).Decode(&enrollment)
	if enrollment.UserID != user.ID {
		t.Fatalf("expected user_id %d, got %d", user.ID, enrollment.UserID)
	}
	if enrollment.CourseID != course.ID {
		t.Fatalf("expected course_id %d, got %d", course.ID, enrollment.CourseID)
	}
}

func TestEnroll_Duplicate(t *testing.T) {
	s, db := setupTestEnrollmentDB(t)
	user := model.User{Name: "bob", Password: "hash"}
	db.Create(&user)
	course := model.Course{Title: "Go 202", Xp: 200}
	db.Create(&course)

	h := handler.NewEnrollmentHandler(s)
	body, _ := json.Marshal(map[string]uint{"course_id": course.ID})

	req := httptest.NewRequest(http.MethodPost, "/enrollments", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+generateToken(t, user.ID, user.Name))
	w := httptest.NewRecorder()
	handler.AuthMiddleware(h.Enroll)(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("first enroll failed: %d", w.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/enrollments", bytes.NewReader(body))
	req2.Header.Set("Authorization", "Bearer "+generateToken(t, user.ID, user.Name))
	w2 := httptest.NewRecorder()
	handler.AuthMiddleware(h.Enroll)(w2, req2)
	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w2.Code)
	}
}

func TestEnroll_Unauthenticated(t *testing.T) {
	s := setupTestEnrollmentStore(t)
	h := handler.NewEnrollmentHandler(s)

	body, _ := json.Marshal(map[string]uint{"course_id": 1})
	req := httptest.NewRequest(http.MethodPost, "/enrollments", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.AuthMiddleware(h.Enroll)(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestUnenroll_Success(t *testing.T) {
	s, db := setupTestEnrollmentDB(t)
	user := model.User{Name: "charlie", Password: "hash"}
	db.Create(&user)
	course := model.Course{Title: "Go 303", Xp: 300}
	db.Create(&course)

	enrollment := model.Enrollment{UserID: user.ID, CourseID: course.ID}
	db.Create(&enrollment)

	h := handler.NewEnrollmentHandler(s)
	req := httptest.NewRequest(http.MethodDelete, "/enrollments/"+itoa(enrollment.ID), nil)
	req.SetPathValue("id", itoa(enrollment.ID))
	w := httptest.NewRecorder()
	h.Unenroll(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestUnenroll_NotFound(t *testing.T) {
	s := setupTestEnrollmentStore(t)
	h := handler.NewEnrollmentHandler(s)

	req := httptest.NewRequest(http.MethodDelete, "/enrollments/99999", nil)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()
	h.Unenroll(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetEnrollmentsByUser(t *testing.T) {
	s, db := setupTestEnrollmentDB(t)
	user := model.User{Name: "dave", Password: "hash"}
	db.Create(&user)
	c1 := model.Course{Title: "Course A", Xp: 10}
	db.Create(&c1)
	c2 := model.Course{Title: "Course B", Xp: 20}
	db.Create(&c2)

	db.Create(&model.Enrollment{UserID: user.ID, CourseID: c1.ID})
	db.Create(&model.Enrollment{UserID: user.ID, CourseID: c2.ID})

	h := handler.NewEnrollmentHandler(s)
	req := httptest.NewRequest(http.MethodGet, "/enrollments/user/"+itoa(user.ID), nil)
	req.SetPathValue("user_id", itoa(user.ID))
	w := httptest.NewRecorder()
	h.FindByUser(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var courses []model.Course
	json.NewDecoder(w.Body).Decode(&courses)
	if len(courses) != 2 {
		t.Fatalf("expected 2 courses, got %d", len(courses))
	}
}

func TestGetEnrollmentsByCourse(t *testing.T) {
	s, db := setupTestEnrollmentDB(t)
	u1 := model.User{Name: "eve", Password: "hash"}
	db.Create(&u1)
	u2 := model.User{Name: "frank", Password: "hash"}
	db.Create(&u2)
	course := model.Course{Title: "Popular Course", Xp: 50}
	db.Create(&course)

	db.Create(&model.Enrollment{UserID: u1.ID, CourseID: course.ID})
	db.Create(&model.Enrollment{UserID: u2.ID, CourseID: course.ID})

	h := handler.NewEnrollmentHandler(s)
	req := httptest.NewRequest(http.MethodGet, "/enrollments/course/"+itoa(course.ID), nil)
	req.SetPathValue("course_id", itoa(course.ID))
	w := httptest.NewRecorder()
	h.FindByCourse(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var users []model.User
	json.NewDecoder(w.Body).Decode(&users)
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
}

func TestGetAllEnrollments(t *testing.T) {
	s, db := setupTestEnrollmentDB(t)
	user := model.User{Name: "grace", Password: "hash"}
	db.Create(&user)
	course := model.Course{Title: "Solo Course", Xp: 75}
	db.Create(&course)
	db.Create(&model.Enrollment{UserID: user.ID, CourseID: course.ID})

	h := handler.NewEnrollmentHandler(s)
	req := httptest.NewRequest(http.MethodGet, "/enrollments", nil)
	w := httptest.NewRecorder()
	h.FindAll(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var enrollments []model.Enrollment
	json.NewDecoder(w.Body).Decode(&enrollments)
	if len(enrollments) != 1 {
		t.Fatalf("expected 1 enrollment, got %d", len(enrollments))
	}
}

func itoa(n uint) string {
	return strconv.Itoa(int(n))
}
