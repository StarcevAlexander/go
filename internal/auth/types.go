package auth

import "github.com/golang-jwt/jwt/v5"

// Claims - структура для хранения claims JWT-токена
// Claims - структура для хранения JWT claims
type Claims struct {
	jwt.RegisteredClaims        // Встроенная структура с стандартными claims
	Login                string `json:"login"`
	Role                 string `json:"role,omitempty"`
}

// Другие типы, связанные с аутентификацией...
