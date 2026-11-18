package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"duolabs/handler"
	"duolabs/model"
	"duolabs/seed"
	"duolabs/store"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const defaultAdminPassword = "admin123"

func main() {
	db, err := openDatabase(envOr("DB_PATH", "duolabs.db"))
	if err != nil {
		log.Fatal(err)
	}

	configureAuth()
	seedDatabase(db)

	staticDir := envOr("STATIC_DIR", "static")
	addr := ":" + envOr("PORT", "8081")
	log.Printf("API running on %s (static files from %s)", addr, staticDir)
	log.Fatal(http.ListenAndServe(addr, handler.CORSMiddleware(newRouter(db, staticDir))))
}

// openDatabase opens the SQLite file and creates or updates every table.
func openDatabase(path string) (*gorm.DB, error) {
	// "record not found" is a normal answer here (e.g. checking whether a
	// name is taken), so it is not logged as an error.
	quiet := logger.New(log.New(os.Stdout, "", log.LstdFlags), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
	})
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{Logger: quiet})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(
		&model.Course{}, &model.Unit{}, &model.Task{}, &model.Lesson{}, &model.Choice{},
		&model.User{}, &model.Enrollment{}, &model.Answer{},
	)
	return db, err
}

// configureAuth reads the JWT signing key from $JWT_SECRET. Without it the
// handler package keeps its development default, which anyone can forge.
func configureAuth() {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		handler.SetJWTSecret(secret)
		return
	}
	log.Println("WARNING: JWT_SECRET is not set; using the insecure development default")
}

// seedDatabase fills a new database with the sample courses, keeps old ones
// up to date and makes sure the admin account exists.
func seedDatabase(db *gorm.DB) {
	if err := seed.Run(db); err != nil {
		log.Printf("seed.Run: %v", err)
	}
	if err := seed.Upgrade(db); err != nil {
		log.Printf("seed.Upgrade: %v", err)
	}

	adminName := envOr("ADMIN_NAME", "admin")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		log.Printf("WARNING: ADMIN_PASSWORD is not set; the admin password is %q", defaultAdminPassword)
		adminPassword = defaultAdminPassword
	}
	if err := seed.EnsureAdmin(store.NewSQLiteUserStore(db), adminName, adminPassword); err != nil {
		log.Fatal(err)
	}
	log.Printf("Admin user %q ready", adminName)
}

// newRouter wires stores, handlers and routes. Writes to courses require an
// admin token; enrolling and answering require any logged-in user.
func newRouter(db *gorm.DB, staticDir string) *http.ServeMux {
	courses := handler.NewCourseHandler(store.NewSQLiteCourseStore(db))
	users := handler.NewUserHandler(store.NewSQLiteUserStore(db))
	enrollments := handler.NewEnrollmentHandler(store.NewSQLiteEnrollmentStore(db))
	progress := handler.NewProgressHandler(store.NewSQLiteProgressStore(db))

	mux := http.NewServeMux()

	mux.HandleFunc("GET /courses", courses.GetAll)
	mux.HandleFunc("POST /courses", handler.AdminMiddleware(courses.Create))
	mux.HandleFunc("GET /courses/{id}", courses.GetByID)
	mux.HandleFunc("DELETE /courses/{id}", handler.AdminMiddleware(courses.Delete))
	mux.HandleFunc("PUT /courses/{id}", handler.AdminMiddleware(courses.Update))

	mux.HandleFunc("GET /courses/{id}/progress", handler.AuthMiddleware(progress.CourseProgress))
	mux.HandleFunc("POST /lessons/{id}/answer", handler.AuthMiddleware(progress.Answer))

	mux.HandleFunc("POST /register", users.Register)
	mux.HandleFunc("POST /login", users.Login)

	mux.HandleFunc("POST /enrollments", handler.AuthMiddleware(enrollments.Enroll))
	mux.HandleFunc("DELETE /enrollments/{id}", handler.AuthMiddleware(enrollments.Unenroll))
	mux.HandleFunc("GET /enrollments", enrollments.FindAll)
	mux.HandleFunc("GET /enrollments/user/{user_id}", enrollments.FindByUser)
	mux.HandleFunc("GET /enrollments/course/{course_id}", enrollments.FindByCourse)

	mux.HandleFunc("GET /", handler.StaticHandler(staticDir))
	return mux
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
