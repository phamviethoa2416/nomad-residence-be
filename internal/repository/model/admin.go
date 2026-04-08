package model

import (
	"nomad-residence-be/internal/domain/entity"
	"time"
)

type Admin struct {
	ID uint `gorm:"primaryKey"`

	Email        string `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string `gorm:"type:varchar(255);not null"`
	FullName     string `gorm:"type:varchar(255);not null"`
	Phone        *string

	Role   entity.AdminRole `gorm:"type:varchar(20);default:'admin'"`
	Status string           `gorm:"type:varchar(20);default:'active';index"`

	FailedLoginAttempts int
	LockedUntil         *time.Time
	LastLoginAt         *time.Time
	PasswordChangedAt   *time.Time

	BaseModel
}

func AdminFromDomain(e *entity.Admin) *Admin {
	if e == nil {
		return nil
	}
	return &Admin{
		ID:                  e.ID,
		Email:               e.Email,
		PasswordHash:        e.PasswordHash,
		FullName:            e.FullName,
		Phone:               e.Phone,
		Role:                e.Role,
		Status:              e.Status,
		FailedLoginAttempts: e.FailedLoginAttempts,
		LockedUntil:         e.LockedUntil,
		LastLoginAt:         e.LastLoginAt,
		PasswordChangedAt:   e.PasswordChangedAt,
		BaseModel:           baseFromDomain(e.BaseModel),
	}
}

func (m *Admin) ToDomain() *entity.Admin {
	if m == nil {
		return nil
	}
	return &entity.Admin{
		ID:                  m.ID,
		Email:               m.Email,
		PasswordHash:        m.PasswordHash,
		FullName:            m.FullName,
		Phone:               m.Phone,
		Role:                m.Role,
		Status:              m.Status,
		FailedLoginAttempts: m.FailedLoginAttempts,
		LockedUntil:         m.LockedUntil,
		LastLoginAt:         m.LastLoginAt,
		PasswordChangedAt:   m.PasswordChangedAt,
		BaseModel:           m.BaseModel.toDomainBase(),
	}
}

func AdminsToDomain(models []Admin) []entity.Admin {
	result := make([]entity.Admin, len(models))
	for i := range models {
		result[i] = *models[i].ToDomain()
	}
	return result
}
