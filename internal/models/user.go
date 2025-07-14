package models

type User struct {
	ID       string   `json:"id"`
	Login    string   `json:"login"`
	Password string   `json:"password"`
	Name     string   `json:"name"`
	Filial   string   `json:"filial"`
	Role     UserRole `json:"role"`
}

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleUser   UserRole = "user"
	RoleHelper UserRole = "helper"
	RoleOwner  UserRole = "owner"
)
