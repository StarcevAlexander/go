package main

import (
	"log"
	"myapp/handlers"
	"myapp/internal/auth"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	// Инициализация зависимостей
	userStorage := auth.NewJSONUserStorage("storage/jsons/users.json")
	if userStorage == nil {
		log.Fatal("Failed to initialize user storage")
	}

	authService := auth.NewAuthService(userStorage, []byte("we-will-rock-you"))
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(authService)

	// Настройка роутера
	r := setupRouter(authService, authHandler, userHandler)

	// Настройка сервера
	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Server starting on :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func setupRouter(authService *auth.AuthService, authHandler *handlers.AuthHandler, userHandler *handlers.UserHandler) *chi.Mux {
	r := chi.NewRouter()

	// Базовые middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Настройка CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Публичные маршруты
	r.Route("/api", func(r chi.Router) {
		r.Post("/login", authHandler.Login)
		r.Post("/register", authHandler.Register)

		// Защищенные маршруты
		r.Group(func(r chi.Router) {
			r.Use(authService.AuthMiddleware)

			// Пользовательские endpoints
			r.Get("/users", userHandler.GetAllUsers)
			r.Get("/users/{id}", userHandler.GetUserData)
			r.Put("/users/{id}", userHandler.UpdateUserData)
			r.Get("/profile", userHandler.GetProfile)

			// Модули
			r.Get("/modules", userHandler.GetModules)

			// Загрузка файлов
			r.Get("/download/{filename}", userHandler.DownloadFile)
		})
	})

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return r
}
