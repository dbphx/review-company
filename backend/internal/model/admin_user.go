package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminUser struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"` // Không trả về client
	Name      string    `gorm:"type:varchar(255)" json:"name"`
	Role      string    `gorm:"type:varchar(20);default:'ADMIN'" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (a *AdminUser) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return
}
