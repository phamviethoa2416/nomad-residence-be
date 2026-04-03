package entity

import "time"

type AdminRole string

const (
	AdminRoleAdmin      AdminRole = "admin"
	AdminRoleSuperAdmin AdminRole = "superadmin"
)

type Admin struct {
	ID uint `gorm:"primaryKey"`

	Email string `gorm:"type:varchar(255);uniqueIndex;not null"`

	PasswordHash string `gorm:"type:varchar(255);not null" json:"-"`

	FullName string `gorm:"type:varchar(255);not null"`
	Phone    *string

	Role   AdminRole `gorm:"type:varchar(20);default:'admin'"`
	Status string    `gorm:"type:varchar(20);default:'active';index"`

	FailedLoginAttempts int
	LockedUntil         *time.Time
	LastLoginAt         *time.Time
	PasswordChangedAt   *time.Time

	BaseModel
}
