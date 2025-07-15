package models

func IsValidRole(role UserRole) bool {
	switch role {
	case RoleOwner, RoleAdmin, RoleTutor, RoleHelper, RoleUser:
		return true
	default:
		return false
	}
}

// IsValidStatus проверяет, является ли статус допустимым
func IsValidStatus(status UserStatus) bool {
	switch status {
	case StatusActive, StatusFrozen, StatusDeleted:
		return true
	default:
		return false
	}
}
