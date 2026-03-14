package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Company struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name         string    `gorm:"type:varchar(255);index;not null" json:"name"`
	LogoURL      string    `gorm:"type:varchar(255)" json:"logo_url"`
	Website      string    `gorm:"type:varchar(255)" json:"website"`
	Industry     string    `gorm:"type:varchar(255)" json:"industry"`
	Size         string    `gorm:"type:varchar(50)" json:"size"`
	Description  string    `gorm:"type:text" json:"description"`
	AvgRating    float64   `gorm:"type:numeric(3,2);default:0" json:"avg_rating"`
	TotalReviews int       `gorm:"type:int;default:0" json:"total_reviews"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (c *Company) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return
}
