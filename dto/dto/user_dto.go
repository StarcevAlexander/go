package dto

import "myapp/internal/models"

type UserResponse struct {
	ID       string            `json:"id"`
	Login    string            `json:"login"`
	Name     string            `json:"name"`
	Filial   string            `json:"filial"`
	Role     models.UserRole   `json:"role"`
	Password string            `json:"password"`
	Status   models.UserStatus `json:"status"`
}
