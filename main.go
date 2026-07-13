package main

import (
	"encoding/json"
	"log"
	"net/http"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Course struct {
	gorm.Model
	Title string `json:"title"`
	Xp    int16  `json:"xp"`
}

type handler struct {
	db *gorm.DB
}

func (h *handler) coursesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getAllCourse(w, r)
	case http.MethodPost:
		h.createCourse(w, r)
	}
}

func (h *handler) getAllCourse(w http.ResponseWriter, r *http.Request) {
	var courses []Course
	h.db.Order("created_at desc").Find(&courses)
	json.NewEncoder(w).Encode(courses)
}

func (h *handler) createCourse(w http.ResponseWriter, r *http.Request) {
	var course Course
	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := h.db.Create(&course).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(course)
}

func main() {
	db, err := gorm.Open(sqlite.Open("duolabs.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(&Course{}); err != nil {
		log.Fatal(err)
	}

	h := handler{db: db}

	http.HandleFunc("/courses", h.coursesHandler)

	log.Println("API running on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
