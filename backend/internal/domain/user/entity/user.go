package entity

import "time"

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type UserStatus string

const (
	StatusActive   UserStatus = "active"
	StatusInactive UserStatus = "inactive"
)

type User struct {
	ID        string     `gorm:"primaryKey"`
	Email     string     `gorm:"size:100;uniqueIndex;not null"`
	Password  string     `gorm:"size:200;not null"` // bcrypt hash
	Name      string     `gorm:"size:100"`
	Role      UserRole   `gorm:"size:20;default:user"`
	Status    UserStatus `gorm:"size:20;default:active"`
	Balance   float64    `gorm:"type:decimal(10,4);default:0"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
}

func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (r *UserRole) String() string {
	return string(*r)
}
