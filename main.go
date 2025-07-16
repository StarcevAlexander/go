package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"log"
	"myapp/handlers"
	"myapp/internal/auth"
	"net/http"
)

func main() {
	// Инициализация хранилища
	userStorage := auth.NewJSONUserStorage("storage/jsons/users.json")

	// Инициализация сервиса аутентификации
	authService := auth.NewAuthService(userStorage, []byte("we-will-rock-you"))

	// Создание обработчиков
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(authService)

	// Создаем маршрутизатор chi
	r := chi.NewRouter()

	// Настройка CORS middleware
	r.Use(cors.Handler(cors.Options{
		// Разрешенные origins (можно указать конкретные домены или "*" для всех)
		AllowedOrigins: []string{"*"}, // Или например: []string{"http://localhost:3000"}

		// Разрешенные HTTP методы
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},

		// Разрешенные заголовки
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},

		// Разрешить передачу cookies (если нужно)
		AllowCredentials: false,

		// Максимальное время кеширования preflight запросов
		MaxAge: 300, // 5 минут
	}))

	// Другие промежуточные обработчики (middleware)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Публичные маршруты (без авторизации)
	r.Post("/login", authHandler.Login)
	r.Post("/register", authHandler.Register)

	// Защищённые маршруты (требуют авторизации)
	r.Group(func(r chi.Router) {
		r.Use(authService.AuthMiddleware)
		r.Get("/users", userHandler.GetAllUsers)
		r.Get("/users/{id}", userHandler.GetUserData)
		r.Put("/users/{id}", userHandler.UpdateUserData)
		r.Get("/profile", userHandler.GetProfile)
		r.Get("/modules", userHandler.GetModules)
		r.Get("/download/{filename}", userHandler.DownloadFile)
	})

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
