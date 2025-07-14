package main

import (
	"encoding/json"
	"go-auth-project/internal/models"
	"go-auth-project/internal/storage"
	"log"
	"net/http"
	"strconv"

	"go-auth-project/handlers"
	"go-auth-project/internal/auth"
)

func main() {
	// Инициализация хранилища
	userStorage := auth.NewJSONUserStorage("storage/users.json")

	// Инициализация сервиса аутентификации
	authService := auth.NewAuthService(userStorage, "we-will-rock-you")

	// Создание обработчиков
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(authService)

	// Настройка маршрутов
	mux := http.NewServeMux()

	// Публичные маршруты
	mux.HandleFunc("POST /login", authHandler.Login)

	// Защищенные маршруты
	protected := http.NewServeMux()
	protected.HandleFunc("/profile", profileHandler)
	protected.HandleFunc("/users", userHandler.GetAllUsers)
	protected.HandleFunc("/register", authHandler.Register)

	// Применяем middleware ТОЛЬКО к защищенным маршрутам
	authChain := authService.AuthMiddleware(auth.WithRoleMiddleware(protected))

	mux.Handle("/api/", http.StripPrefix("/api", authChain))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем пользователя из контекста
	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	dataStorage := storage.NewDataStorage("storage/data.json")
	data, err := dataStorage.LoadData()
	if err != nil {
		http.Error(w, "Failed to load data", http.StatusInternalServerError)
		return
	}

	// Преобразуем "users" из интерфейса в []models.UserData
	var users []models.UserData
	if userDataList, ok := data["users"].([]interface{}); ok {
		for _, u := range userDataList {
			userMap := u.(map[string]interface{})
			var userData models.UserData
			err := json.Unmarshal(toJSON(userMap), &userData)
			if err != nil {
				return
			}
			users = append(users, userData)
		}
	}

	// Конвертируем user.ID (string) в int
	userIDInt, err := strconv.Atoi(user.ID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	// Ищем пользователя по ID
	var foundUser *models.UserData
	for _, u := range users {
		if u.ID == userIDInt {
			foundUser = &u
			break
		}
	}

	// Отправляем ответ
	if foundUser != nil {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(foundUser)
		if err != nil {
			return
		}
	} else {
		http.Error(w, "User not found", http.StatusNotFound)
	}
}

func toJSON(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}
