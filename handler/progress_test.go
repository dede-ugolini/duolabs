package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"duolabs/handler"
	"duolabs/model"
	"duolabs/store"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type progressEnv struct {
	db       *gorm.DB
	courses  *handler.CourseHandler
	enroll   *handler.EnrollmentHandler
	progress *handler.ProgressHandler
	course   model.Course
}

func newProgressEnv(t *testing.T) *progressEnv {
	t.Helper()
	tmp, err := os.CreateTemp("", "duolabs-progress-*.db")
	if err != nil {
		t.Fatal(err)
	}
	_ = tmp.Close()
	t.Cleanup(func() { _ = os.Remove(tmp.Name()) })

	db, err := gorm.Open(sqlite.Open(tmp.Name()), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Course{}, &model.Unit{}, &model.Task{}, &model.Lesson{},
		&model.Choice{}, &model.User{}, &model.Enrollment{}, &model.Answer{}); err != nil {
		t.Fatal(err)
	}
	env := &progressEnv{
		db:       db,
		courses:  handler.NewCourseHandler(store.NewSQLiteCourseStore(db)),
		enroll:   handler.NewEnrollmentHandler(store.NewSQLiteEnrollmentStore(db)),
		progress: handler.NewProgressHandler(store.NewSQLiteProgressStore(db)),
	}
	lesson := func(q string) model.Lesson {
		return model.Lesson{Question: q, Choices: []model.Choice{
			{Label: "right", Correct: true}, {Label: "wrong"},
		}}
	}
	course := model.Course{
		Title: "C", Difficulty: model.Beginner, Xp: 200,
		Units: []model.Unit{{Title: "U", Tasks: []model.Task{{Title: "T", Lessons: []model.Lesson{lesson("Q1"), lesson("Q2")}}}}},
	}
	if err := db.Create(&course).Error; err != nil {
		t.Fatal(err)
	}
	env.course = course
	return env
}

func (e *progressEnv) lesson(i int) model.Lesson {
	return e.course.Units[0].Tasks[0].Lessons[i]
}

func (e *progressEnv) choice(lessonIdx int, correct bool) uint {
	for _, c := range e.lesson(lessonIdx).Choices {
		if c.Correct == correct {
			return c.ID
		}
	}
	return 0
}

func (e *progressEnv) enrollUser(t *testing.T, userID uint) {
	t.Helper()
	body, _ := json.Marshal(map[string]uint{"course_id": e.course.ID})
	req := httptest.NewRequest(http.MethodPost, "/enrollments", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+generateToken(t, userID, "u"))
	w := httptest.NewRecorder()
	handler.AuthMiddleware(e.enroll.Enroll)(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("enroll: expected 201, got %d", w.Code)
	}
}

func (e *progressEnv) answer(t *testing.T, userID uint, auth bool, lessonID uint, choiceID uint) (int, bool) {
	t.Helper()
	body, _ := json.Marshal(map[string]uint{"choice_id": choiceID})
	req := httptest.NewRequest(http.MethodPost, "/lessons/"+strconv.Itoa(int(lessonID))+"/answer", bytes.NewReader(body))
	req.SetPathValue("id", strconv.Itoa(int(lessonID)))
	if auth {
		req.Header.Set("Authorization", "Bearer "+generateToken(t, userID, "u"))
	}
	w := httptest.NewRecorder()
	handler.AuthMiddleware(e.progress.Answer)(w, req)
	var resp struct{ Correct bool }
	_ = json.NewDecoder(w.Body).Decode(&resp)
	return w.Code, resp.Correct
}

func (e *progressEnv) progressOf(t *testing.T, userID uint) model.CourseProgress {
	t.Helper()
	id := strconv.Itoa(int(e.course.ID))
	req := httptest.NewRequest(http.MethodGet, "/courses/"+id+"/progress", nil)
	req.SetPathValue("id", id)
	req.Header.Set("Authorization", "Bearer "+generateToken(t, userID, "u"))
	w := httptest.NewRecorder()
	handler.AuthMiddleware(e.progress.CourseProgress)(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("progress: expected 200, got %d", w.Code)
	}
	var p model.CourseProgress
	_ = json.NewDecoder(w.Body).Decode(&p)
	return p
}

func TestAnswer_RequiresAuth(t *testing.T) {
	e := newProgressEnv(t)
	if code, _ := e.answer(t, 0, false, e.lesson(0).ID, e.choice(0, true)); code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", code)
	}
}

func TestAnswer_RequiresEnrollment(t *testing.T) {
	e := newProgressEnv(t)
	if code, _ := e.answer(t, 7, true, e.lesson(0).ID, e.choice(0, true)); code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", code)
	}
}

func TestAnswer_UnknownLessonAndBadChoice(t *testing.T) {
	e := newProgressEnv(t)
	e.enrollUser(t, 7)
	if code, _ := e.answer(t, 7, true, 9999, 1); code != http.StatusNotFound {
		t.Fatalf("unknown lesson: expected 404, got %d", code)
	}
	if code, _ := e.answer(t, 7, true, e.lesson(0).ID, e.choice(1, true)); code != http.StatusBadRequest {
		t.Fatalf("choice of another lesson: expected 400, got %d", code)
	}
	if code, _ := e.answer(t, 7, true, e.lesson(0).ID, 0); code != http.StatusBadRequest {
		t.Fatalf("missing choice: expected 400, got %d", code)
	}
}

func TestAnswer_WrongThenRightCompletesCourse(t *testing.T) {
	e := newProgressEnv(t)
	e.enrollUser(t, 7)

	if code, ok := e.answer(t, 7, true, e.lesson(0).ID, e.choice(0, false)); code != http.StatusOK || ok {
		t.Fatalf("wrong answer: expected 200/false, got %d/%v", code, ok)
	}
	if p := e.progressOf(t, 7); p.CompletedLessons != 0 || p.Completed {
		t.Fatalf("wrong answer must not complete anything: %+v", p)
	}

	if code, ok := e.answer(t, 7, true, e.lesson(0).ID, e.choice(0, true)); code != http.StatusOK || !ok {
		t.Fatalf("right answer: expected 200/true, got %d/%v", code, ok)
	}
	p := e.progressOf(t, 7)
	if p.TotalLessons != 2 || p.CompletedLessons != 1 || p.Completed || p.XpEarned != 0 {
		t.Fatalf("after 1 of 2 lessons: %+v", p)
	}

	if _, ok := e.answer(t, 7, true, e.lesson(1).ID, e.choice(1, true)); !ok {
		t.Fatal("second lesson right answer should be correct")
	}
	p = e.progressOf(t, 7)
	if !p.Completed || p.CompletedLessons != 2 || p.XpEarned != 200 || len(p.CompletedLessonIDs) != 2 {
		t.Fatalf("after all lessons: %+v", p)
	}
}

func TestAnswer_CorrectIsSticky(t *testing.T) {
	e := newProgressEnv(t)
	e.enrollUser(t, 7)
	e.answer(t, 7, true, e.lesson(0).ID, e.choice(0, true))
	e.answer(t, 7, true, e.lesson(0).ID, e.choice(0, false))
	if p := e.progressOf(t, 7); p.CompletedLessons != 1 {
		t.Fatalf("a later wrong answer must not undo completion: %+v", p)
	}
}

func TestProgress_IsPerUser(t *testing.T) {
	e := newProgressEnv(t)
	e.enrollUser(t, 7)
	e.enrollUser(t, 8)
	e.answer(t, 7, true, e.lesson(0).ID, e.choice(0, true))
	if p := e.progressOf(t, 8); p.CompletedLessons != 0 {
		t.Fatalf("user 8 must not inherit user 7 progress: %+v", p)
	}
}

func TestProgress_UnknownCourse(t *testing.T) {
	e := newProgressEnv(t)
	req := httptest.NewRequest(http.MethodGet, "/courses/999/progress", nil)
	req.SetPathValue("id", "999")
	req.Header.Set("Authorization", "Bearer "+generateToken(t, 7, "u"))
	w := httptest.NewRecorder()
	handler.AuthMiddleware(e.progress.CourseProgress)(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func getCourse(t *testing.T, e *progressEnv, token string) model.Course {
	t.Helper()
	id := strconv.Itoa(int(e.course.ID))
	req := httptest.NewRequest(http.MethodGet, "/courses/"+id, nil)
	req.SetPathValue("id", id)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	e.courses.GetByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get course: expected 200, got %d", w.Code)
	}
	var c model.Course
	_ = json.NewDecoder(w.Body).Decode(&c)
	return c
}

func countCorrect(c model.Course) int {
	n := 0
	for _, l := range c.Units[0].Tasks[0].Lessons {
		for _, ch := range l.Choices {
			if ch.Correct {
				n++
			}
		}
	}
	return n
}

func TestGetCourse_HidesAnswersFromStudents(t *testing.T) {
	e := newProgressEnv(t)
	if n := countCorrect(getCourse(t, e, "")); n != 0 {
		t.Fatalf("anonymous must not see answers, saw %d", n)
	}
	if n := countCorrect(getCourse(t, e, roleToken(t, model.RoleUser))); n != 0 {
		t.Fatalf("student must not see answers, saw %d", n)
	}
	if n := countCorrect(getCourse(t, e, roleToken(t, model.RoleAdmin))); n != 2 {
		t.Fatalf("admin should see 2 correct choices, saw %d", n)
	}
}

func TestCreateCourse_InvalidLessons(t *testing.T) {
	h := handler.NewCourseHandler(setupTestStore(t))
	cases := map[string][]model.Choice{
		"no correct choice":  {{Label: "a"}, {Label: "b"}},
		"two correct":        {{Label: "a", Correct: true}, {Label: "b", Correct: true}},
		"single choice":      {{Label: "a", Correct: true}},
		"blank choice label": {{Label: "a", Correct: true}, {Label: " "}},
	}
	for name, choices := range cases {
		course := model.Course{Title: "C", Difficulty: model.Beginner, Units: []model.Unit{{Title: "U",
			Tasks: []model.Task{{Title: "T", Lessons: []model.Lesson{{Question: "Q?", Choices: choices}}}}}}}
		body, _ := json.Marshal(course)
		w := httptest.NewRecorder()
		h.Create(w, httptest.NewRequest(http.MethodPost, "/courses", bytes.NewReader(body)))
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: expected 400, got %d", name, w.Code)
		}
	}
}

func TestAnswer_ConcurrentSubmitsDoNotFail(t *testing.T) {
	e := newProgressEnv(t)
	e.enrollUser(t, 7)
	right := e.choice(0, true)
	lessonID := e.lesson(0).ID

	var wg sync.WaitGroup
	codes := make(chan int, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			code, _ := e.answer(t, 7, true, lessonID, right)
			codes <- code
		}()
	}
	wg.Wait()
	close(codes)
	for code := range codes {
		if code != http.StatusOK {
			t.Fatalf("concurrent submit: expected 200, got %d", code)
		}
	}
	if p := e.progressOf(t, 7); p.CompletedLessons != 1 {
		t.Fatalf("expected exactly 1 completed lesson, got %+v", p)
	}
}

func TestGetCourses_ListDoesNotLeakAnswers(t *testing.T) {
	e := newProgressEnv(t)
	w := httptest.NewRecorder()
	e.courses.GetAll(w, httptest.NewRequest(http.MethodGet, "/courses", nil))
	body := w.Body.String()
	if strings.Contains(body, `"correct":true`) || strings.Contains(body, `"correct_choice_id":`) && !strings.Contains(body, `"correct_choice_id":0`) {
		t.Fatalf("course list must not expose answers: %s", body)
	}
}
