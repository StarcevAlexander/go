package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"myapp/internal/models"
)

// JSONUserStorage реализует UserStorage для хранения в JSON
type JSONUserStorage struct {
	filePath string
	mu       sync.Mutex
}

// NewJSONUserStorage создает новое JSON хранилище пользователей
func NewJSONUserStorage(filePath string) *JSONUserStorage {
	return &JSONUserStorage{
		filePath: filePath,
	}
}

// UserStorage определяет интерфейс для работы с пользователями
type UserStorage interface {
	CreateUser(user models.User) error
	GetUserByLogin(login string) (models.User, error)
	GetUserByID(id string) (models.User, error)
	GetAllUsers() (map[string]models.User, error)
}

// AuthService предоставляет методы аутентификации
type AuthService struct {
	UserStorage UserStorage
	jwtSecret   string
}

// NewAuthService создает новый экземпляр AuthService
func NewAuthService(userStorage UserStorage, jwtSecret string) *AuthService {
	return &AuthService{
		UserStorage: userStorage,
		jwtSecret:   jwtSecret,
	}
}

func (s *JSONUserStorage) GetAllUsers() (map[string]models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readUsers()
}

// CreateUser создает нового пользователя
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
	s.mu.Lock()
	defer s.mu.Unlock()

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
	s.mu.Lock()
	defer s.mu.Unlock()

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

func (s *JSONUserStorage) readUsers() (map[string]models.User, error) {
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

func (s *JSONUserStorage) writeUsers(users map[string]models.User) error {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

// Register регистрирует нового пользователя
func (s *AuthService) Register(user models.User) error {
	return s.UserStorage.CreateUser(user)
}

// Login выполняет аутентификацию пользователя и возвращает токен
func (s *AuthService) Login(login, password string) (string, models.User, error) {
	user, err := s.UserStorage.GetUserByLogin(login)
	if err != nil {
		return "", models.User{}, errors.New("user not found")
	}

	// Заменяем проверку хеша на прямое сравнение
	if user.Password != password {
		return "", models.User{}, errors.New("invalid credentials")
	}

	token, err := s.generateToken(user.ID, user.Login)
	if err != nil {
		return "", models.User{}, err
	}

	// Очищаем пароль перед возвратом
	user.Password = ""
	return token, user, nil
}

// AuthMiddleware проверяет JWT токен
func (s *AuthService) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		claims, err := s.validateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		userID, ok := claims["sub"].(string)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		user, err := s.UserStorage.GetUserByID(userID)
		if err != nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RoleMiddleware проверяет роль пользователя
func (s *AuthService) RoleMiddleware(requiredRole models.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value("user").(models.User)
			if !ok || models.UserRole(user.Role) != requiredRole {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// generateToken создает JWT токен
func (s *AuthService) generateToken(userID, login string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   userID,
		"login": login,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})

	// Убедитесь, что ключ преобразуется в []byte
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) validateToken(tokenString string) (jwt.MapClaims, error) {
	// Добавляем проверку алгоритма подписи
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
