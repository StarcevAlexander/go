package main

import (
	"log"
	"myapp/handlers"
	"myapp/internal/auth"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Инициализация
	userStorage := auth.NewJSONUserStorage("storage/jsons/users.json")
	authService := auth.NewAuthService(userStorage, []byte("we-will-rock-you"))
	authHandler := handlers.NewAuthHandler(authService)

	// Настройка роутера
	r := chi.NewRouter()

	// Базовые middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Кастомный CORS middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Устанавливаем CORS заголовки
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
			w.Header().Set("Access-Control-Expose-Headers", "Link")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "300")

			// Если это OPTIONS запрос - сразу отвечаем 200
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Публичные маршруты
	r.Post("/login", authHandler.Login)
	r.Post("/register", authHandler.Register)

	// Защищенные маршруты
	r.Group(func(r chi.Router) {
		r.Use(authService.AuthMiddleware)

		r.Get("/download/tutor-data.json", func(w http.ResponseWriter, r *http.Request) {
			// Ваша логика для выдачи файла
			w.Header().Set("Content-Type", "application/json")
			http.ServeFile(w, r, "path/to/tutor-data.json")
		})

		// Другие защищенные маршруты...
	})

	// Запуск сервера
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
