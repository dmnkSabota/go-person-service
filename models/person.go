package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Person struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ExternalID  uuid.UUID `json:"external_id" gorm:"type:uuid;unique;not null"`
	Name        string    `json:"name" gorm:"not null"`
	Email       string    `json:"email" gorm:"not null"`
	DateOfBirth time.Time `json:"date_of_birth" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SavePersonRequest struct {
	ExternalID  uuid.UUID `json:"external_id" binding:"required"`
	Name        string    `json:"name" binding:"required"`
	Email       string    `json:"email" binding:"required,email"`
	DateOfBirth time.Time `json:"date_of_birth" binding:"required"`
}

type PersonResponse struct {
	ExternalID  uuid.UUID `json:"external_id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	DateOfBirth time.Time `json:"date_of_birth"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (p *Person) BeforeCreate(*gorm.DB) error {
	if p.ExternalID == uuid.Nil {
		p.ExternalID = uuid.New()
	}
	return nil
}

func (p *Person) ToResponse() PersonResponse {
	return PersonResponse{
		ExternalID:  p.ExternalID,
		Name:        p.Name,
		Email:       p.Email,
		DateOfBirth: p.DateOfBirth,
	}
}

func FromSaveRequest(req SavePersonRequest) Person {
	return Person{
		ExternalID:  req.ExternalID,
		Name:        req.Name,
		Email:       req.Email,
		DateOfBirth: req.DateOfBirth,
	}
}
