package storage

import (
	"encoding/json"
	"myapp/internal/models"
	"os"
	"sync"
)

// JSONUserStorage реализует UserStorage для хранения в JSON файле
type JSONUserStorage struct {
	filePath string
	mu       sync.Mutex
}

// NewUserStorage создает новое хранилище пользователей
func NewUserStorage(filePath string) UserStorage {
	return &JSONUserStorage{
		filePath: filePath,
	}
}

// CreateUser добавляет нового пользователя
func (s *JSONUserStorage) CreateUser(user models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	users, err := s.readUsers()
	if err != nil {
		return err
	}

	users[user.Login] = user
	return s.writeUsers(users)
}

// GetUserByLogin возвращает пользователя по логину
func (s *JSONUserStorage) GetUserByLogin(login string) (models.User, error) {
	users, err := s.readUsers()
	if err != nil {
		return models.User{}, err
	}

	user, exists := users[login]
	if !exists {
		return models.User{}, os.ErrNotExist
	}

	return user, nil
}

// GetUserByID возвращает пользователя по ID
func (s *JSONUserStorage) GetUserByID(id string) (models.User, error) {
	users, err := s.readUsers()
	if err != nil {
		return models.User{}, err
	}

	for _, user := range users {
		if user.ID == id {
			return user, nil
		}
	}

	return models.User{}, os.ErrNotExist
}

// UpdateUser обновляет данные пользователя
func (s *JSONUserStorage) UpdateUser(user models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	users, err := s.readUsers()
	if err != nil {
		return err
	}

	// Проверяем существование пользователя
	if _, exists := users[user.Login]; !exists {
		return os.ErrNotExist
	}

	users[user.Login] = user
	return s.writeUsers(users)
}

// DeleteUser удаляет пользователя
func (s *JSONUserStorage) DeleteUser(login string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	users, err := s.readUsers()
	if err != nil {
		return err
	}

	delete(users, login)
	return s.writeUsers(users)
}

// GetAllUsers возвращает всех пользователей
func (s *JSONUserStorage) GetAllUsers() (map[string]models.User, error) {
	return s.readUsers()
}

// readUsers читает пользователей из файла
func (s *JSONUserStorage) readUsers() (map[string]models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]models.User), nil
		}
		return nil, err
	}

	var users map[string]models.User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// writeUsers записывает пользователей в файл
func (s *JSONUserStorage) writeUsers(users map[string]models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}
