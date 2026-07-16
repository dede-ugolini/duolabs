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

	courseStore := store.NewSQLiteCourseStore(db)
	courseHandler := handler.NewCourseHandler(courseStore)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /courses", courseHandler.GetAll)
	mux.HandleFunc("POST /courses", courseHandler.Create)
	mux.HandleFunc("GET /courses/{id}", courseHandler.GetByID)
	mux.HandleFunc("DELETE /courses/{id}", courseHandler.Delete)
	mux.HandleFunc("PUT /courses/{id}", courseHandler.Update)

	log.Println("API running on :8081")
	log.Fatal(http.ListenAndServe(":8081", mux))
}
