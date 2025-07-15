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
	users    []models.User // Кэш пользователей в памяти
}

type usersFile struct {
	Users []models.User `json:"users"`
}

// NewJSONUserStorage создает новое JSON хранилище пользователей
func NewJSONUserStorage(filePath string) *JSONUserStorage {
	storage := &JSONUserStorage{
		filePath: filePath,
	}
	// Загружаем пользователей при инициализации
	err := storage.loadUsers()
	if err != nil {
		return nil
	}
	return storage
}

// UserStorage определяет интерфейс для работы с пользователями
type UserStorage interface {
	CreateUser(user models.User) error
	GetUserByLogin(login string) (models.User, error)
	GetUserByID(id string) (models.User, error)
	GetAllUsers() ([]models.User, error)
	SaveAllUsers(users []models.User) error
}

// AuthService предоставляет методы аутентификации
type AuthService struct {
	UserStorage UserStorage
	jwtKey      []byte // Добавьте это поле
}

// NewAuthService создает новый экземпляр AuthService
func NewAuthService(storage UserStorage, jwtKey []byte) *AuthService {
	return &AuthService{
		UserStorage: storage,
		jwtKey:      jwtKey, // Инициализируем поле
	}
}

func (s *JSONUserStorage) loadUsers() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			s.users = []models.User{}
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

func (s *JSONUserStorage) saveUsers() error {
	data, err := json.MarshalIndent(usersFile{Users: s.users}, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

// GetAllUsers возвращает всех пользователей
func (s *JSONUserStorage) GetAllUsers() ([]models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.users, nil
}

// CreateUser создает нового пользователя
func (s *JSONUserStorage) CreateUser(user models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.users {
		if u.Login == user.Login {
			return fmt.Errorf("user with login %s already exists", user.Login)
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

// AuthMiddleware проверяет JWT токен и статус пользователя
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
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return s.jwtKey, nil
		})

		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		userID := claims.Subject
		if userID == "" {
			http.Error(w, "Invalid token claims: missing sub", http.StatusUnauthorized)
			return
		}

		user, err := s.UserStorage.GetUserByID(userID)
		if err != nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		if user.Status != models.StatusActive {
			http.Error(w, "User is not active", http.StatusForbidden)
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
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
		Login: login,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtKey)
}

func (s *AuthService) validateToken(tokenString string) (jwt.MapClaims, error) {
	// Добавляем проверку алгоритма подписи
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// SaveAllUsers сохраняет всех пользователей
func (s *JSONUserStorage) SaveAllUsers(users []models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users = users
	return s.saveUsers()
}
