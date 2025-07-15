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
	// 1. Парсинг входных данных
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 2. Валидация обязательных полей
	if user.Login == "" || user.Password == "" || user.Name == "" || user.Filial == "" || user.Role == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// 3. Если есть авторизованный пользователь - проверяем его права
	if requester, ok := r.Context().Value("user").(models.User); ok {
		// Проверка прав доступа для регистрации
		switch requester.Role {
		case models.RoleAdmin:
			if user.Role == models.RoleOwner || user.Role == models.RoleAdmin {
				http.Error(w, "Admin can't register owners or admins", http.StatusForbidden)
				return
			}
			if user.Filial != requester.Filial {
				http.Error(w, "Admin can only register users in their own filial", http.StatusForbidden)
				return
			}
		case models.RoleHelper:
			if user.Role != models.RoleUser {
				http.Error(w, "Helper can only register users", http.StatusForbidden)
				return
			}
			if user.Filial != requester.Filial {
				http.Error(w, "Helper can only register users in their own filial", http.StatusForbidden)
				return
			}
		default:
			// Для других ролей (кроме owner) запрещаем регистрацию
			if requester.Role != models.RoleOwner {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}
	} else {
		// Если нет авторизованного пользователя - разрешаем регистрацию только обычных пользователей
		if user.Role != models.RoleUser {
			http.Error(w, "Only user registration is allowed without authentication", http.StatusForbidden)
			return
		}
	}

	// 4. Проверка уникальности логина
	if _, err := h.authService.UserStorage.GetUserByLogin(user.Login); err == nil {
		http.Error(w, "Login already taken", http.StatusConflict)
		return
	}

	// 5. Установка значений по умолчанию
	user.ID = strconv.FormatInt(time.Now().UnixMilli(), 10)
	user.Status = models.StatusActive

	// 6. Создание пользователя
	if err := h.authService.Register(user); err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
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
