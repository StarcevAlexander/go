package models

type User struct {
	ID       string     `json:"id"`
	Login    string     `json:"login"`
	Password string     `json:"password"`
	Name     string     `json:"name"`
	Filial   string     `json:"filial"`
	Role     UserRole   `json:"role"`
	Status   UserStatus `json:"status"`
}

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleUser   UserRole = "user"
	RoleHelper UserRole = "helper"
	RoleOwner  UserRole = "owner"
	RoleTutor  UserRole = "tutor"
)

type UserStatus string

const (
	StatusActive  UserStatus = "active"
	StatusFrozen  UserStatus = "frozen"
	StatusDeleted UserStatus = "deleted"
)
