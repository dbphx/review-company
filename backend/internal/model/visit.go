package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Visit struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Path      string    `gorm:"type:varchar(255);index;not null" json:"path"`
	SessionID string    `gorm:"type:varchar(255);index;not null" json:"session_id"`
	IPAddress string    `gorm:"type:varchar(100);index" json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
}

func (v *Visit) BeforeCreate(tx *gorm.DB) (err error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}
