package seed

import (
	"os"
	"strings"
	"testing"

	"duolabs/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestSampleCourses_AreWellFormed(t *testing.T) {
	legacy := map[string]bool{}
	for _, q := range legacyQuestions {
		legacy[q] = true
	}
	for _, course := range sampleCourses() {
		seen := map[string]bool{}
		lessons := 0
		for _, unit := range course.Units {
			for _, task := range unit.Tasks {
				for _, lesson := range task.Lessons {
					lessons++
					q := lesson.Question
					if strings.TrimSpace(q) == "" || legacy[q] || seen[q] {
						t.Errorf("%s: question %q is empty, legacy or duplicated", course.Title, q)
					}
					seen[q] = true
					if len(lesson.Choices) != 4 {
						t.Errorf("%s: %q has %d choices, want 4", course.Title, q, len(lesson.Choices))
					}
					correct, labels := 0, map[string]bool{}
					for _, c := range lesson.Choices {
						if strings.TrimSpace(c.Label) == "" || labels[c.Label] {
							t.Errorf("%s: %q has an empty or duplicated choice %q", course.Title, q, c.Label)
						}
						labels[c.Label] = true
						if c.Correct {
							correct++
						}
					}
					if correct != 1 {
						t.Errorf("%s: %q has %d correct choices, want 1", course.Title, q, correct)
					}
				}
			}
		}
		if lessons < 5 {
			t.Errorf("%s: only %d lessons", course.Title, lessons)
		}
	}
}

func upgradeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	tmp, err := os.CreateTemp("", "duolabs-seed-*.db")
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
		&model.Choice{}, &model.Answer{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func legacyCourse(title string) model.Course {
	return model.Course{Title: title, Difficulty: model.Beginner, Xp: 500, Units: []model.Unit{{
		Title: "Módulo 1", Tasks: []model.Task{{Title: "Tarefa", Lessons: []model.Lesson{{
			Question: legacyQuestions[0],
			Choices:  []model.Choice{{Label: "const x = 10", Correct: true}, {Label: "var x = 10"}},
		}}}},
	}}}
}

func countLessons(t *testing.T, db *gorm.DB, title string) int64 {
	t.Helper()
	var n int64
	err := db.Table("lessons").
		Joins("JOIN tasks ON tasks.id = lessons.task_id AND tasks.deleted_at IS NULL").
		Joins("JOIN units ON units.id = tasks.unit_id AND units.deleted_at IS NULL").
		Joins("JOIN courses ON courses.id = units.course_id AND courses.deleted_at IS NULL").
		Where("courses.title = ? AND lessons.deleted_at IS NULL", title).Count(&n).Error
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestUpgradeSampleCourses_ReplacesLegacyContent(t *testing.T) {
	db := upgradeTestDB(t)

	goCourse := legacyCourse("Go Fundamentals")
	scratch := legacyCourse("testes")
	custom := model.Course{Title: "Meu curso", Difficulty: model.Beginner, Units: []model.Unit{{
		Title: "U", Tasks: []model.Task{{Title: "T", Lessons: []model.Lesson{{
			Question: "Pergunta do admin?",
			Choices:  []model.Choice{{Label: "a", Correct: true}, {Label: "b"}},
		}}}},
	}}}
	for _, c := range []*model.Course{&goCourse, &scratch, &custom} {
		if err := db.Create(c).Error; err != nil {
			t.Fatal(err)
		}
	}
	legacyLessonID := goCourse.Units[0].Tasks[0].Lessons[0].ID
	if err := db.Create(&model.Answer{UserID: 5, LessonID: legacyLessonID, ChoiceID: 1, Correct: true}).Error; err != nil {
		t.Fatal(err)
	}

	for run := 0; run < 2; run++ { // second run proves idempotence
		if err := Upgrade(db); err != nil {
			t.Fatal(err)
		}

		var wantGo int
		for _, u := range sampleCourses()[0].Units {
			for _, tk := range u.Tasks {
				wantGo += len(tk.Lessons)
			}
		}
		if got := countLessons(t, db, "Go Fundamentals"); got != int64(wantGo) {
			t.Fatalf("run %d: Go Fundamentals has %d lessons, want %d", run, got, wantGo)
		}
		if got := countLessons(t, db, "testes"); got != 0 {
			t.Fatalf("run %d: non-sample course must lose the legacy lesson and stay empty, has %d", run, got)
		}
		if got := countLessons(t, db, "Meu curso"); got != 1 {
			t.Fatalf("run %d: admin course must be untouched, has %d lessons", run, got)
		}
		var legacyLeft, orphanAnswers, python int64
		db.Model(&model.Lesson{}).Where("question IN ?", legacyQuestions).Count(&legacyLeft)
		db.Model(&model.Answer{}).Where("lesson_id = ?", legacyLessonID).Count(&orphanAnswers)
		db.Model(&model.Course{}).Where("title = ?", "Python para Análise de Dados").Count(&python)
		if legacyLeft != 0 || orphanAnswers != 0 {
			t.Fatalf("run %d: legacy lessons=%d answers=%d, want 0", run, legacyLeft, orphanAnswers)
		}
		if python != 0 {
			t.Fatalf("run %d: a missing sample course must not be recreated", run)
		}
	}
}
