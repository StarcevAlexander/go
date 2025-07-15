package handlers

import (
	"encoding/json"
	"myapp/dto/dto"
	"myapp/internal/auth"
	"myapp/internal/models"
	"myapp/internal/storage"
	"myapp/internal/utils"
	"net/http"
	"strconv"
	"time"
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
		// Admin видит только свой филиал и только НЕ удалённых пользователей
		for _, u := range allUsers {
			if u.Filial == user.Filial && u.Status != models.StatusDeleted {
				filteredUsers = append(filteredUsers, u)
			}
		}

	case models.RoleHelper:
		// Helper видит только пользователей из своего филиала с ролью "user" и статусом НЕ deleted
		for _, u := range allUsers {
			if u.Filial == user.Filial && u.Role == models.RoleUser && u.Status != models.StatusDeleted {
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
			ID:       u.ID,
			Login:    u.Login,
			Password: u.Password,
			Name:     u.Name,
			Filial:   u.Filial,
			Role:     u.Role,
			Status:   u.Status,
		})
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(dtos)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) UpdateUser(writer http.ResponseWriter, request *http.Request) {

}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Получаем пользователя из контекста
	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Определяем файл в зависимости от роли пользователя
	var dataFile string
	switch user.Role {
	case models.RoleUser:
		dataFile = "storage/jsons/user-data.json"
	case models.RoleAdmin:
		dataFile = "storage/jsons/admin-data.json"
	case models.RoleOwner:
		dataFile = "storage/jsons/admin-data.json"
	case models.RoleTutor:
		dataFile = "storage/jsons/tutor-data.json"
	case models.RoleHelper:
		dataFile = "storage/jsons/helper-data.json"
	default:
		http.Error(w, "Forbidden: unknown role", http.StatusForbidden)
		return
	}

	// Загружаем данные из нужного файла
	dataStorage := storage.NewDataStorage(dataFile)
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
			err := json.Unmarshal(utils.ToJSON(userMap), &userData)
			if err != nil {
				http.Error(w, "Failed to parse user data", http.StatusInternalServerError)
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
		if err := json.NewEncoder(w).Encode(foundUser); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	} else {
		http.Error(w, "User not found", http.StatusNotFound)
	}
}

func (h *UserHandler) GetModules(w http.ResponseWriter, r *http.Request) {
	// Получаем пользователя из контекста
	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Проверяем, является ли пользователь tutor
	if user.Role != models.RoleTutor {
		http.Error(w, "Forbidden: only tutors can access this resource", http.StatusForbidden)
		return
	}

	// Ищем профиль текущего пользователя в tutor-data.json
	tutorDataStorage := storage.NewDataStorage("storage/tutor-data.json")
	tutorRawData, err := tutorDataStorage.LoadData()
	if err != nil {
		http.Error(w, "Failed to load tutor data", http.StatusInternalServerError)
		return
	}

	var tutors []models.UserData
	if userDataList, ok := tutorRawData["users"].([]interface{}); ok {
		for _, u := range userDataList {
			userMap := u.(map[string]interface{})
			var userData models.UserData
			err := json.Unmarshal(utils.ToJSON(userMap), &userData)
			if err != nil {
				http.Error(w, "Failed to parse tutor data", http.StatusInternalServerError)
				return
			}
			tutors = append(tutors, userData)
		}
	}

	// Конвертируем ID пользователя из string в int
	userIDInt, err := strconv.Atoi(user.ID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	// Находим текущего пользователя среди tutors
	var tutor *models.UserData
	for _, t := range tutors {
		if t.ID == userIDInt {
			tutor = &t
			break
		}
	}

	if tutor == nil {
		http.Error(w, "Tutor not found", http.StatusNotFound)
		return
	}

	// Загружаем все модули из modules-description-links.json
	moduleStorage := storage.NewDataStorage("storage/jsons/modules-description-links.json")
	moduleRawData, err := moduleStorage.LoadData()
	if err != nil {
		http.Error(w, "Failed to load module data", http.StatusInternalServerError)
		return
	}

	var allModules []models.Module
	if moduleList, ok := moduleRawData["learningModules"].([]interface{}); ok {
		for _, m := range moduleList {
			moduleMap := m.(map[string]interface{})
			var module models.Module
			err := json.Unmarshal(utils.ToJSON(moduleMap), &module)
			if err != nil {
				http.Error(w, "Failed to parse module data", http.StatusInternalServerError)
				return
			}
			allModules = append(allModules, module)
		}
	}

	// Создаём мапу для быстрого поиска даты
	now := time.Now().UnixMilli()
	tutorModuleMap := make(map[int]int64)
	for _, info := range tutor.Modules {
		if info.Date > now {
			tutorModuleMap[info.Module] = info.Date
		}
	}

	// Фильтруем модули
	var resultModules []models.Module
	for _, module := range allModules {
		if _, exists := tutorModuleMap[module.ID]; exists {
			resultModules = append(resultModules, module)
		}
	}

	// Отправляем результат
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resultModules); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func convertMapToSlice(usersMap map[string]models.User) []models.User {
	users := make([]models.User, 0, len(usersMap))
	for _, user := range usersMap {
		users = append(users, user)
	}
	return users
}
