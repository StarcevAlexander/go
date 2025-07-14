package handlers

import (
	"encoding/json"
	"go-auth-project/dto/dto"
	"go-auth-project/internal/auth"
	"go-auth-project/internal/models"
	"net/http"
)

type UserHandler struct {
	authService *auth.AuthService
}

func NewUserHandler(authService *auth.AuthService) *UserHandler {
	return &UserHandler{authService: authService}
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	// Получаем пользователя из контекста
	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем всех пользователей в виде map
	allUsersMap, err := h.authService.UserStorage.GetAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Преобразуем map в slice
	allUsers := convertMapToSlice(allUsersMap)

	var filteredUsers []models.User

	switch user.Role {
	case models.RoleOwner:
		// Owner видит всех
		filteredUsers = allUsers

	case models.RoleAdmin:
		// Admin видит только свой филиал
		for _, u := range allUsers {
			if u.Filial == user.Filial {
				filteredUsers = append(filteredUsers, u)
			}
		}

	case models.RoleHelper:
		// Helper видит только пользователей из своего филиала с ролью "user"
		for _, u := range allUsers {
			if u.Filial == user.Filial && u.Role == models.RoleUser {
				filteredUsers = append(filteredUsers, u)
			}
		}

	case models.RoleUser:
		// User не имеет доступа
		http.Error(w, "Forbidden: users cannot access this resource", http.StatusForbidden)
		return

	default:
		http.Error(w, "Forbidden: unknown role", http.StatusForbidden)
		return
	}

	// Преобразуем в DTO
	var dtos []dto.UserResponse
	for _, u := range filteredUsers {
		dtos = append(dtos, dto.UserResponse{
			ID:     u.ID,
			Login:  u.Login,
			Name:   u.Name,
			Filial: u.Filial,
			Role:   string(u.Role),
		})
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(dtos)
	if err != nil {
		return
	}
}

func convertMapToSlice(usersMap map[string]models.User) []models.User {
	users := make([]models.User, 0, len(usersMap))
	for _, user := range usersMap {
		users = append(users, user)
	}
	return users
}
