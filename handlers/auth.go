package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"myapp/internal/auth"
	"myapp/internal/models"
)

type AuthHandler struct {
	authService *auth.AuthService
}

func NewAuthHandler(authService *auth.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Получаем пользователя из контекста (который сделал запрос)
	requester, ok := r.Context().Value("user").(models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Проверяем права доступа
	switch requester.Role {
	case models.RoleOwner:
		// owner может регистрировать всех
	case models.RoleAdmin, models.RoleHelper:
		// admin и helper могут регистрировать только в своем филиале
	default:
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Проверяем, что все поля заполнены
	if user.Login == "" || user.Password == "" || user.Name == "" || user.Filial == "" || user.Role == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// Устанавливаем статус по умолчанию — active ✅
	user.Status = models.StatusActive

	// Проверяем права на создание пользователя с указанной ролью
	switch requester.Role {
	case models.RoleAdmin:
		if user.Role == models.RoleOwner || user.Role == models.RoleAdmin {
			http.Error(w, "Admin can't register owners or admins", http.StatusForbidden)
			return
		}
		// Проверяем, что филиал совпадает с филиалом администратора
		if user.Filial != requester.Filial {
			http.Error(w, "Admin can only register users in their own filial", http.StatusForbidden)
			return
		}
	case models.RoleHelper:
		if user.Role != models.RoleUser {
			http.Error(w, "Helper can only register users", http.StatusForbidden)
			return
		}
		// Проверяем, что филиал совпадает с филиалом помощника
		if user.Filial != requester.Filial {
			http.Error(w, "Helper can only register users in their own filial", http.StatusForbidden)
			return
		}
	}

	// Проверяем, существует ли пользователь с таким login
	_, err := h.authService.UserStorage.GetUserByLogin(user.Login)
	if err == nil {
		http.Error(w, "Login already taken", http.StatusConflict)
		return
	}

	// Генерируем уникальный ID как текущую миллисекунду
	user.ID = strconv.FormatInt(time.Now().UnixMilli(), 10)

	// Регистрируем нового пользователя
	if err := h.authService.Register(user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, user, err := h.authService.Login(creds.Login, creds.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if user.Status != models.StatusActive {
		http.Error(w, "User account is not active", http.StatusForbidden)
		return
	}

	response := struct {
		Token string      `json:"token"`
		User  models.User `json:"user"`
	}{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}
