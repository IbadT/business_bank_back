// internal/models/roles.go
package models

// Роли пользователей
const (
	RoleUser  = "user"  // Обычный пользователь
	RoleAdmin = "admin" // Администратор
)

// IsValidRole проверяет, является ли роль валидной
func IsValidRole(role string) bool {
	return role == RoleUser || role == RoleAdmin
}

// HasRole проверяет, имеет ли пользователь указанную роль
func (u *User) HasRole(role string) bool {
	return u.Role == role
}

// IsAdmin проверяет, является ли пользователь администратором
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}
