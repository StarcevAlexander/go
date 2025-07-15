package handlers

//
//import (
//	"encoding/json"
//	"fmt"
//	"github.com/go-chi/chi/v5"
//	"myapp/dto/dto"
//	"myapp/internal/models"
//	"net/http"
//	"strings"
//)
//
//func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
//	// Получаем текущего пользователя из контекста
//	currentUser, ok := r.Context().Value("user").(models.User)
//	if !ok {
//		http.Error(w, "Unauthorized: invalid user context", http.StatusUnauthorized)
//		return
//	}
//
//	// Получаем ID пользователя из URL
//	userID := chi.URLParam(r, "id")
//	if userID == "" {
//		http.Error(w, "User ID is required", http.StatusBadRequest)
//		return
//	}
//
//	// Декодируем тело запроса
//	var updateData dto.UserUpdateRequest
//	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
//		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
//		return
//	}
//
//	// Получаем всех пользователей
//	allUsers, err := h.authService.UserStorage.GetAllUsers()
//	if err != nil {
//		http.Error(w, "Failed to get users: "+err.Error(), http.StatusInternalServerError)
//		return
//	}
//
//	// Проверяем существование пользователя
//	userToUpdate, exists := allUsers[userID]
//	if !exists {
//		http.Error(w, fmt.Sprintf("User with ID %s not found", userID), http.StatusNotFound)
//		return
//	}
//
//	// Проверяем права доступа
//	if err := checkUpdatePermissions(currentUser, userToUpdate); err != nil {
//		http.Error(w, "Permission denied: "+err.Error(), http.StatusForbidden)
//		return
//	}
//
//	// Валидация и обновление полей
//	if err := validateAndUpdateUser(&userToUpdate, updateData, currentUser); err != nil {
//		http.Error(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//
//	// Сохраняем изменения
//	allUsers[userID] = userToUpdate
//	if err := h.authService.UserStorage.SaveAllUsers(allUsers); err != nil {
//		http.Error(w, "Failed to save user data: "+err.Error(), http.StatusInternalServerError)
//		return
//	}
//
//	// Возвращаем обновленного пользователя
//	sendUpdatedUserResponse(w, userToUpdate)
//}
//
//// Вспомогательные функции
//
//func checkUpdatePermissions(currentUser, targetUser models.User) error {
//	switch currentUser.Role {
//	case models.RoleOwner:
//		return nil
//	case models.RoleAdmin:
//		if targetUser.Filial != currentUser.Filial {
//			return fmt.Errorf("can only update users from your filial")
//		}
//		return nil
//	case models.RoleHelper:
//		if targetUser.Filial != currentUser.Filial || targetUser.Role != models.RoleUser {
//			return fmt.Errorf("can only update regular users from your filial")
//		}
//		return nil
//	default:
//		return fmt.Errorf("insufficient permissions")
//	}
//}
//
//func validateAndUpdateUser(user *models.User, updateData dto.UserUpdateRequest, currentUser models.User) error {
//	// Обновление имени
//	if updateData.Name != "" {
//		user.Name = strings.TrimSpace(updateData.Name)
//		if user.Name == "" {
//			return fmt.Errorf("name cannot be empty")
//		}
//	}
//
//	// Обновление филиала
//	if updateData.Filial != "" {
//		user.Filial = strings.TrimSpace(updateData.Filial)
//		if user.Filial == "" {
//			return fmt.Errorf("filial cannot be empty")
//		}
//	}
//
//	// Обновление пароля
//	if updateData.Password != "" {
//		if len(updateData.Password) < 6 {
//			return fmt.Errorf("password must be at least 6 characters")
//		}
//		user.Password = updateData.Password
//	}
//
//	// Обновление статуса
//	if updateData.Status != "" {
//		status := models.UserStatus(updateData.Status)
//		if !models.IsValidStatus(status) {
//			return fmt.Errorf("invalid status value")
//		}
//		user.Status = status
//	}
//
//	// Обновление роли (если разрешено)
//	if updateData.Role != "" {
//		if currentUser.Role != models.RoleOwner && currentUser.Role != models.RoleAdmin {
//			return fmt.Errorf("only owner and admin can change roles")
//		}
//
//		role := models.UserRole(updateData.Role)
//		if !models.IsValidRole(role) {
//			return fmt.Errorf("invalid role value")
//		}
//
//		if currentUser.Role == models.RoleAdmin &&
//			(role == models.RoleOwner || role == models.RoleAdmin) {
//			return fmt.Errorf("admin cannot assign owner or admin roles")
//		}
//
//		user.Role = role
//	}
//
//	return nil
//}
//
//func sendUpdatedUserResponse(w http.ResponseWriter, user models.User) {
//	response := dto.UserResponse{
//		ID:     user.ID,
//		Login:  user.Login,
//		Name:   user.Name,
//		Filial: user.Filial,
//		Role:   user.Role,
//		Status: user.Status,
//	}
//
//	w.Header().Set("Content-Type", "application/json")
//	if err := json.NewEncoder(w).Encode(response); err != nil {
//		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
//	}
//}
