package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CompanyRequestStatus string

const (
	CompanyRequestPending  CompanyRequestStatus = "PENDING"
	CompanyRequestApproved CompanyRequestStatus = "APPROVED"
	CompanyRequestRejected CompanyRequestStatus = "REJECTED"
)

type CompanyRequest struct {
	ID          uuid.UUID            `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string               `gorm:"type:varchar(255);not null;index" json:"name"`
	LogoURL     string               `gorm:"type:varchar(500)" json:"logo_url"`
	Size        string               `gorm:"type:varchar(100)" json:"size"`
	Status      CompanyRequestStatus `gorm:"type:varchar(20);not null;default:'PENDING';index" json:"status"`
	RequestedIP string               `gorm:"type:varchar(45)" json:"-"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

func (r *CompanyRequest) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.Status == "" {
		r.Status = CompanyRequestPending
	}
	return
}
