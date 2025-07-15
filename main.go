package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"myapp/handlers"
	"myapp/internal/auth"
	"net/http"
)

func main() {
	// Инициализация хранилища
	userStorage := auth.NewJSONUserStorage("storage/users.json")

	// Инициализация сервиса аутентификации
	authService := auth.NewAuthService(userStorage, []byte("we-will-rock-you"))

	// Создание обработчиков
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(authService)

	// Создаем маршрутизатор chi
	r := chi.NewRouter()

	// Промежуточные обработчики (middleware)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Публичные маршруты (без авторизации)
	r.Post("/login", authHandler.Login)
	r.Post("/register", authHandler.Register)

	// Защищённые маршруты (требуют авторизации)
	r.Group(func(r chi.Router) {
		r.Use(authService.AuthMiddleware) // middleware для авторизации
		//r.Use(auth.WithRoleMiddleware)    // middleware для проверки роли

		// CRUD пользователей
		r.Get("/users", userHandler.GetAllUsers)
		//r.Put("/users/{id}", userHandler.UpdateUser) // PUT /users/123
		r.Get("/profile", userHandler.GetProfile)
		r.Get("/modules", userHandler.GetModules)
	})

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
