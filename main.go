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

	http.Handle("/courses", courseHandler)

	log.Println("API running on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
