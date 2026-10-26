package main

import (
	"log"
	"net/http"

	"duolabs/handler"
	"duolabs/model"
	"duolabs/store"

	"github.com/glebarez/sqlite"
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

	seedDatabase(db)

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
	log.Fatal(http.ListenAndServe(":8081", corsMiddleware(mux)))
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

func seedDatabase(db *gorm.DB) {
	var count int64
	db.Model(&model.Course{}).Count(&count)
	if count > 0 {
		return
	}

	courses := []model.Course{
		{
			Title:       "Go Fundamentals",
			Description: "Aprenda a linguagem Go do zero ao avançado com concorrência e microsserviços.",
			Category:    "Programação",
			Difficulty:  model.Beginner,
			Xp:          500,
			Units: []model.Unit{
				{
					Title: "Módulo 1: Introdução ao Go",
					Tasks: []model.Task{
						{
							Title: "Sintaxe e Variáveis",
							Lessons: []model.Lesson{
								{
									Question: "Como declarar uma variável constante em Go?",
									Choices: []model.Choice{
										{Label: "const x = 10"},
										{Label: "var x = 10"},
										{Label: "let x = 10"},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Title:       "Python para Análise de Dados",
			Description: "Domine Pandas, NumPy e visualização de dados com Python.",
			Category:    "Ciência de Dados",
			Difficulty:  model.Intermediate,
			Xp:          750,
			Units: []model.Unit{
				{
					Title: "Módulo 1: Pandas Básico",
					Tasks: []model.Task{
						{
							Title: "Manipulação de DataFrames",
							Lessons: []model.Lesson{
								{
									Question: "Qual função do Pandas é usada para ler um arquivo CSV?",
									Choices: []model.Choice{
										{Label: "pd.read_csv()"},
										{Label: "pd.open_csv()"},
										{Label: "pd.load_csv()"},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Title:       "TypeScript Avançado",
			Description: "Tipos avançados, generics, utilitários e metaprogramação em TypeScript.",
			Category:    "Desenvolvimento Web",
			Difficulty:  model.Advanced,
			Xp:          1000,
			Units: []model.Unit{
				{
					Title: "Módulo 1: Tipos Condicionais e Mapped Types",
					Tasks: []model.Task{
						{
							Title: "Generics Avançados",
							Lessons: []model.Lesson{
								{
									Question: "Qual operador utilitário torna todas as propriedades de um tipo opcionais?",
									Choices: []model.Choice{
										{Label: "Partial<T>"},
										{Label: "Optional<T>"},
										{Label: "Maybe<T>"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, c := range courses {
		db.Create(&c)
	}
	log.Println("Database seeded with sample courses successfully!")
}
