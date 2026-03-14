package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Auth0ID   string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"auth0_id"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Name      string    `gorm:"type:varchar(255)" json:"name"`
	Picture   string    `gorm:"type:varchar(255)" json:"picture"`
	Provider  string    `gorm:"type:varchar(50);default:'google'" json:"provider"`
	Role      string    `gorm:"type:varchar(20);default:'MODERATOR'" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	AllowedCompanies []Company `gorm:"many2many:user_allowed_companies;" json:"allowed_companies,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}
