package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"duolabs/handler"
	"duolabs/model"
	"duolabs/seed"
	"duolabs/store"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open(envOr("DB_PATH", "duolabs.db")), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(&model.Course{}); err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Unit{}); err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Task{}); err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Lesson{}); err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Choice{}); err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Enrollment{}); err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Answer{}); err != nil {
		log.Fatal(err)
	}

	seed.Run(db)
	if err := seed.Upgrade(db); err != nil {
		log.Printf("seed.Upgrade: %v", err)
	}

	courseStore := store.NewSQLiteCourseStore(db)
	courseHandler := handler.NewCourseHandler(courseStore)

	userStore := store.NewSQLiteUserStore(db)
	userHandler := handler.NewUserHandler(userStore)

	adminName := envOr("ADMIN_NAME", "admin")
	if err := handler.EnsureAdmin(userStore, adminName, envOr("ADMIN_PASSWORD", "admin123")); err != nil {
		log.Fatal(err)
	}
	log.Printf("Admin user %q ready", adminName)

	enrollmentStore := store.NewSQLiteEnrollmentStore(db)
	enrollmentHandler := handler.NewEnrollmentHandler(enrollmentStore)

	progressStore := store.NewSQLiteProgressStore(db)
	progressHandler := handler.NewProgressHandler(progressStore)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /courses", courseHandler.GetAll)
	mux.HandleFunc("POST /courses", handler.AdminMiddleware(courseHandler.Create))
	mux.HandleFunc("GET /courses/{id}", courseHandler.GetByID)
	mux.HandleFunc("DELETE /courses/{id}", handler.AdminMiddleware(courseHandler.Delete))
	mux.HandleFunc("PUT /courses/{id}", handler.AdminMiddleware(courseHandler.Update))

	mux.HandleFunc("GET /courses/{id}/progress", handler.AuthMiddleware(progressHandler.CourseProgress))
	mux.HandleFunc("POST /lessons/{id}/answer", handler.AuthMiddleware(progressHandler.Answer))

	mux.HandleFunc("POST /register", userHandler.Register)
	mux.HandleFunc("POST /login", userHandler.Login)

	mux.HandleFunc("POST /enrollments", handler.AuthMiddleware(enrollmentHandler.Enroll))
	mux.HandleFunc("DELETE /enrollments/{id}", handler.AuthMiddleware(enrollmentHandler.Unenroll))
	mux.HandleFunc("GET /enrollments", enrollmentHandler.FindAll)
	mux.HandleFunc("GET /enrollments/user/{user_id}", enrollmentHandler.FindByUser)
	mux.HandleFunc("GET /enrollments/course/{course_id}", enrollmentHandler.FindByCourse)

	staticDir := envOr("STATIC_DIR", "static")
	mux.HandleFunc("GET /", staticHandler(staticDir))

	addr := ":" + envOr("PORT", "8081")
	log.Printf("API running on %s (static files from %s)", addr, staticDir)
	log.Fatal(http.ListenAndServe(addr, corsMiddleware(mux)))
}

// staticHandler serves the built frontend and falls back to index.html so that
// client-side routing keeps working on deep links.
func staticHandler(dir string) http.HandlerFunc {
	fileServer := http.FileServer(http.Dir(dir))
	return func(w http.ResponseWriter, r *http.Request) {
		name := filepath.Join(dir, filepath.Clean("/"+r.URL.Path))
		if info, err := os.Stat(name); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
