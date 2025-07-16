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

	// Basic CORS
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		//AllowedOrigins: []string{"http://localhost:63343"}, // Allow your frontend origin
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		//AllowCredentials: true,
		AllowedOrigins:   []string{"*"}, // Разрешаем все origins
		AllowCredentials: false,         // Должно быть false при использовании "*"
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

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
		r.Get("/users", userHandler.GetAllUsers)
		r.Get("/users/{id}", userHandler.GetUserData)
		r.Put("/users/{id}", userHandler.UpdateUserData)
		r.Get("/profile", userHandler.GetProfile)
		r.Get("/modules", userHandler.GetModules)

		//для ручного бэкапа
		r.Get("/download/{filename}", userHandler.DownloadFile)
	})

	log.Println("Server starting on :8080eee")
	log.Fatal(http.ListenAndServe(":8080", r))
}
