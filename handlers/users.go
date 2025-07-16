package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"myapp/dto/dto"
	"myapp/internal/auth"
	"myapp/internal/models"
	"myapp/internal/storage"
	"myapp/internal/utils"
	"net/http"
	"os"
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

	// Получаем всех пользователей (теперь это slice, а не map)
	allUsers, err := h.authService.UserStorage.GetAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
			Password: "********", // Маскируем пароль в ответе
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

func (h *UserHandler) GetUserData(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем текущего пользователя (кто делает запрос)
	currentUser, ok := r.Context().Value("user").(models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Получаем ID запрашиваемого пользователя из URL
	userID := chi.URLParam(r, "id")

	// 3. Проверяем права доступа к данным этого пользователя
	allUsers, err := h.authService.UserStorage.GetAllUsers()
	if err != nil {
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	// Находим запрашиваемого пользователя
	var targetUser *models.User
	for _, u := range allUsers {
		if u.ID == userID {
			targetUser = &u
			break
		}
	}

	if targetUser == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Проверяем права доступа
	switch currentUser.Role {
	case models.RoleOwner:
		// Owner может видеть любого
	case models.RoleAdmin:
		// Admin только своего филиала (и удалённых)
		if targetUser.Filial != currentUser.Filial {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	case models.RoleHelper:
		// Helper только своего филиала, роль User и не удалённых
		if targetUser.Filial != currentUser.Filial ||
			targetUser.Role != models.RoleUser ||
			targetUser.Status == models.StatusDeleted {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	case models.RoleUser:
		// User не имеет доступа
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	default:
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// 4. Определяем файл с данными в зависимости от роли целевого пользователя
	var dataFile string
	switch targetUser.Role {
	case models.RoleUser:
		dataFile = "storage/jsons/user-data.json"
	case models.RoleAdmin, models.RoleOwner:
		dataFile = "storage/jsons/admin-data.json"
	case models.RoleTutor:
		dataFile = "storage/jsons/tutor-data.json"
	case models.RoleHelper:
		dataFile = "storage/jsons/helper-data.json"
	default:
		http.Error(w, "Unknown user role", http.StatusInternalServerError)
		return
	}

	// 5. Загружаем данные из соответствующего файла
	dataStorage := storage.NewDataStorage(dataFile)
	data, err := dataStorage.LoadData()
	if err != nil {
		http.Error(w, "Failed to load user data", http.StatusInternalServerError)
		return
	}

	// 6. Ищем данные конкретного пользователя
	var usersData []models.UserData
	if userDataList, ok := data["users"].([]interface{}); ok {
		for _, u := range userDataList {
			userMap := u.(map[string]interface{})
			var userData models.UserData
			err := json.Unmarshal(utils.ToJSON(userMap), &userData)
			if err != nil {
				http.Error(w, "Failed to parse user data", http.StatusInternalServerError)
				return
			}
			usersData = append(usersData, userData)
		}
	}

	// Конвертируем ID из string в int для сравнения
	targetUserID, err := strconv.Atoi(targetUser.ID)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusInternalServerError)
		return
	}

	// Находим данные запрашиваемого пользователя
	var foundUserData *models.UserData
	for _, ud := range usersData {
		if ud.ID == targetUserID {
			foundUserData = &ud
			break
		}
	}

	if foundUserData == nil {
		http.Error(w, "User data not found", http.StatusNotFound)
		return
	}

	// 7. Возвращаем найденные данные
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(foundUserData); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *UserHandler) UpdateUserData(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем текущего пользователя
	currentUser, ok := r.Context().Value("user").(models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Получаем ID обновляемого пользователя
	userID := chi.URLParam(r, "id")

	// 3. Получаем данные для обновления
	var updateData struct {
		Links   []models.Link       `json:"links,omitempty"`
		Modules []models.ModuleInfo `json:"modules,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 4. Получаем список всех пользователей, чтобы найти целевого
	allUsers, err := h.authService.UserStorage.GetAllUsers()
	if err != nil {
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	var targetUser *models.User
	for _, u := range allUsers {
		if u.ID == userID {
			targetUser = &u
			break
		}
	}
	if targetUser == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// 5. Проверяем права доступа
	switch currentUser.Role {
	case models.RoleOwner:
		// Owner может обновлять всех
	case models.RoleAdmin:
		if targetUser.Filial != currentUser.Filial {
			http.Error(w, "Forbidden: can only update users in your filial", http.StatusForbidden)
			return
		}
	case models.RoleHelper:
		if targetUser.Filial != currentUser.Filial ||
			targetUser.Role != models.RoleUser ||
			targetUser.Status == models.StatusDeleted {
			http.Error(w, "Forbidden: can only update not deleted users in your filial", http.StatusForbidden)
			return
		}
	default:
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// 6. Определяем файл с данными
	var dataFile string
	switch targetUser.Role {
	case models.RoleUser:
		dataFile = "storage/jsons/user-data.json"
	case models.RoleAdmin, models.RoleOwner:
		dataFile = "storage/jsons/admin-data.json"
	case models.RoleTutor:
		dataFile = "storage/jsons/tutor-data.json"
	case models.RoleHelper:
		dataFile = "storage/jsons/helper-data.json"
	default:
		http.Error(w, "Unknown user role", http.StatusInternalServerError)
		return
	}

	// 7. Загружаем данные из файла
	dataStorage := storage.NewDataStorage(dataFile)
	data, err := dataStorage.LoadData()
	if err != nil {
		http.Error(w, "Failed to load user data", http.StatusInternalServerError)
		return
	}

	// 8. Обновляем userData
	var updatedUsers []models.UserData
	if userDataList, ok := data["users"].([]interface{}); ok {
		for _, u := range userDataList {
			userMap := u.(map[string]interface{})
			var userData models.UserData
			if err := json.Unmarshal(utils.ToJSON(userMap), &userData); err != nil {
				http.Error(w, "Failed to parse user data", http.StatusInternalServerError)
				return
			}

			// Конвертируем ID для сравнения
			targetUserID, _ := strconv.Atoi(targetUser.ID)
			if userData.ID == targetUserID {
				// Обновляем только те поля, которые есть в UserData
				userData.Links = updateData.Links
				userData.Modules = updateData.Modules
			}
			updatedUsers = append(updatedUsers, userData)
		}
	}

	// 9. Сохраняем обновлённые данные
	updatedData := map[string]interface{}{
		"users": updatedUsers,
	}
	if err := dataStorage.SaveData(updatedData); err != nil {
		http.Error(w, "Failed to save updated data", http.StatusInternalServerError)
		return
	}

	// 10. Обновляем основные данные пользователя в хранилище
	result := h.authService.UserStorage.UpdateUserData(*targetUser)
	if result != nil {
		if e, ok := result.(error); ok {
			http.Error(w, "Failed to update user: "+e.Error(), http.StatusInternalServerError)
			return
		} else {
			http.Error(w, "Unexpected error during update", http.StatusInternalServerError)
			return
		}
	}

	// 11. Ответ
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"message": "User data updated successfully",
		"user_id": userID,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) GetUserList(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем текущего пользователя из контекста
	currentUser, ok := r.Context().Value("user").(models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Проверяем, что пользователь — RoleOwner
	if currentUser.Role != models.RoleOwner {
		http.Error(w, "Forbidden: only owner can download this file", http.StatusForbidden)
		return
	}

	// 3. Путь к файлу
	filePath := "storage/jsons/users.json"

	// 4. Проверяем существование файла
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to access file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Настраиваем заголовки для скачивания
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=users.json")

	// 6. Отправляем содержимое файла
	http.ServeFile(w, r, filePath)
}
