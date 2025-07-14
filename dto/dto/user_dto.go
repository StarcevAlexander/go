package dto

type UserResponse struct {
	ID     string `json:"id"`
	Login  string `json:"login"`
	Name   string `json:"name"`
	Filial string `json:"filial"`
	Role   string `json:"role"`
}
