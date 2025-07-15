package storage

import (
	"encoding/json"
	"myapp/internal/models"
	"os"
	"sync"
)

// JSONUserStorage реализует UserStorage для хранения в JSON
type JSONUserStorage struct {
	filePath string
	mu       sync.Mutex
	users    []models.User // Храним пользователей в slice
}

// Структура для представления JSON файла
type usersFile struct {
	Users []models.User `json:"users"`
}

// Загружает пользователей из файла
func (s *JSONUserStorage) loadUsers() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			s.users = []models.User{} // Инициализируем пустым slice
			return nil
		}
		return err
	}

	var file usersFile
	if err := json.Unmarshal(data, &file); err != nil {
		return err
	}

	s.users = file.Users
	return nil
}

// Сохраняет пользователей в файл
func (s *JSONUserStorage) saveUsers() error {
	data, err := json.MarshalIndent(usersFile{Users: s.users}, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

// CreateUser добавляет нового пользователя
func (s *JSONUserStorage) CreateUser(user models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем, нет ли уже пользователя с таким логином
	for _, u := range s.users {
		if u.Login == user.Login {
			return os.ErrExist
		}
	}

	s.users = append(s.users, user)
	return s.saveUsers()
}

// GetUserByLogin возвращает пользователя по логину
func (s *JSONUserStorage) GetUserByLogin(login string) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, user := range s.users {
		if user.Login == login {
			return user, nil
		}
	}

	return models.User{}, os.ErrNotExist
}

// GetUserByID возвращает пользователя по ID
func (s *JSONUserStorage) GetUserByID(id string) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, user := range s.users {
		if user.ID == id {
			return user, nil
		}
	}

	return models.User{}, os.ErrNotExist
}

// UpdateUser обновляет данные пользователя
func (s *JSONUserStorage) UpdateUser(updatedUser models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, user := range s.users {
		if user.ID == updatedUser.ID {
			s.users[i] = updatedUser
			return s.saveUsers()
		}
	}

	return os.ErrNotExist
}

// DeleteUser удаляет пользователя
func (s *JSONUserStorage) DeleteUser(login string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, user := range s.users {
		if user.Login == login {
			// Удаляем пользователя из slice
			s.users = append(s.users[:i], s.users[i+1:]...)
			return s.saveUsers()
		}
	}

	return os.ErrNotExist
}

// GetAllUsers возвращает всех пользователей
func (s *JSONUserStorage) GetAllUsers() ([]models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.users, nil
}

// SaveAllUsers сохраняет всех пользователей
func (s *JSONUserStorage) SaveAllUsers(users []models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users = users
	return s.saveUsers()
}
