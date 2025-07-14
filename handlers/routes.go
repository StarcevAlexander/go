package handlers

import (
	"encoding/json"
	"myapp/internal/auth"
	"myapp/internal/storage"
	"net/http"
)

type Handlers struct {
	authService *auth.AuthService
	userStorage auth.UserStorage
	dataStorage *storage.DataStorage
}

func SetupRoutes(authService *auth.AuthService, userStorage auth.UserStorage, dataStorage *storage.DataStorage) *http.ServeMux {
	h := &Handlers{
		authService: authService,
		userStorage: userStorage,
		dataStorage: dataStorage,
	}

	router := http.NewServeMux()

	// Публичные маршруты
	router.HandleFunc("POST /register", h.RegisterHandler)
	router.HandleFunc("POST /login", h.LoginHandler)

	// Защищенные маршруты
	protected := http.NewServeMux()
	protected.HandleFunc("GET /users", h.GetAllUsers)
	protected.HandleFunc("GET /profile", h.ProfileHandler)
	protected.HandleFunc("POST /data", h.AddDataHandler)
	protected.HandleFunc("GET /data", h.GetDataHandler)

	// Применяем middleware
	router.Handle("/", authService.AuthMiddleware(
		authService.RoleMiddleware("admin")(protected),
	))

	return router
}

func (h *Handlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Реализация регистрации
}

func (h *Handlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Реализация входа
}

func (h *Handlers) ProfileHandler(w http.ResponseWriter, r *http.Request) {
	// Получение профиля пользователя
}

func (h *Handlers) AddDataHandler(w http.ResponseWriter, r *http.Request) {
	// Добавление данных
}

func (h *Handlers) GetDataHandler(w http.ResponseWriter, r *http.Request) {
	// Получение данных
}

func (h *Handlers) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userStorage.GetAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Очищаем пароли перед отправкой
	for k, user := range users {
		user.Password = ""
		users[k] = user
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
