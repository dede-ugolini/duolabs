package main

import (
	"log"
	"net/http"

	"duolabs/handler"
	"duolabs/model"
	"duolabs/store"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("duolabs.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(&model.Course{}); err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Enrollment{}); err != nil {
		log.Fatal(err)
	}

	courseStore := store.NewSQLiteCourseStore(db)
	courseHandler := handler.NewCourseHandler(courseStore)

	userStore := store.NewSQLiteUserStore(db)
	userHandler := handler.NewUserHandler(userStore)

	enrollmentStore := store.NewSQLiteEnrollmentStore(db)
	enrollmentHandler := handler.NewEnrollmentHandler(enrollmentStore)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /courses", courseHandler.GetAll)
	mux.HandleFunc("POST /courses", courseHandler.Create)
	mux.HandleFunc("GET /courses/{id}", courseHandler.GetByID)
	mux.HandleFunc("DELETE /courses/{id}", courseHandler.Delete)
	mux.HandleFunc("PUT /courses/{id}", courseHandler.Update)

	mux.HandleFunc("POST /register", userHandler.Register)
	mux.HandleFunc("POST /login", userHandler.Login)

	mux.HandleFunc("POST /enrollments", handler.AuthMiddleware(enrollmentHandler.Enroll))
	mux.HandleFunc("DELETE /enrollments/{id}", handler.AuthMiddleware(enrollmentHandler.Unenroll))
	mux.HandleFunc("GET /enrollments", enrollmentHandler.FindAll)
	mux.HandleFunc("GET /enrollments/user/{user_id}", enrollmentHandler.FindByUser)
	mux.HandleFunc("GET /enrollments/course/{course_id}", enrollmentHandler.FindByCourse)

	log.Println("API running on :8081")
	log.Fatal(http.ListenAndServe(":8081", mux))
}
