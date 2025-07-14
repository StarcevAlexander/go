package storage

import (
	"myapp/internal/models"
)

type UserStorage interface {
	CreateUser(user models.User) error
	GetUserByID(id string) (models.User, error)
	GetUserByLogin(login string) (models.User, error)
	UpdateUser(user models.User) error
	DeleteUser(id string) error
	GetAllUsers() (map[string]models.User, error)
}
