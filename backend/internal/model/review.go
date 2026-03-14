package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReviewStatus string

const (
	StatusApproved ReviewStatus = "APPROVED"
	StatusHidden   ReviewStatus = "HIDDEN"
	StatusDeleted  ReviewStatus = "DELETED"
)

type Review struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"company_id"`
	AuthorName   string         `gorm:"type:varchar(100);default:'Ẩn danh'" json:"author_name"`
	Rating       int            `gorm:"type:int;not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	Title        string         `gorm:"type:varchar(255);not null" json:"title"`
	Content      string         `gorm:"type:text;not null" json:"content"`
	Pros         string         `gorm:"type:text" json:"pros"`
	Cons         string         `gorm:"type:text" json:"cons"`
	SalaryGross  *float64       `gorm:"type:numeric(15,2)" json:"salary_gross"`
	InterviewExp string         `gorm:"type:text" json:"interview_exp"`
	Status       ReviewStatus   `gorm:"type:varchar(20);default:'APPROVED';index" json:"status"`
	IPAddress    string         `gorm:"type:varchar(45)" json:"-"`
	SessionID    string         `gorm:"type:varchar(255)" json:"-"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	CommentCount int            `gorm:"column:comment_count;->;-:migration" json:"comment_count"`

	// Relationships
	Company  Company   `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Comments []Comment `gorm:"foreignKey:ReviewID" json:"comments,omitempty"`
}

func (r *Review) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.AuthorName == "" {
		r.AuthorName = "Ẩn danh"
	}
	if r.Status == "" {
		r.Status = StatusApproved
	}
	return
}
