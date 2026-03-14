package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CommentStatus string

const (
	CommentStatusApproved CommentStatus = "APPROVED"
	CommentStatusHidden   CommentStatus = "HIDDEN"
	CommentStatusDeleted  CommentStatus = "DELETED"
)

type Comment struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	ReviewID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"review_id"`
	ParentCommentID *uuid.UUID     `gorm:"type:uuid;index" json:"parent_comment_id"`
	AuthorName      string         `gorm:"type:varchar(100);default:'Ẩn danh'" json:"author_name"`
	Content         string         `gorm:"type:text;not null" json:"content"`
	Status          CommentStatus  `gorm:"type:varchar(20);default:'APPROVED'" json:"status"`
	IPAddress       string         `gorm:"type:varchar(45)" json:"-"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	LikeCount       int64          `gorm:"column:like_count;->;-:migration" json:"like_count"`
	DislikeCount    int64          `gorm:"column:dislike_count;->;-:migration" json:"dislike_count"`

	// Relationships
	Review        Review    `gorm:"foreignKey:ReviewID" json:"-"`
	ParentComment *Comment  `gorm:"foreignKey:ParentCommentID" json:"-"`
	Replies       []Comment `gorm:"foreignKey:ParentCommentID" json:"replies,omitempty"`
}

func (c *Comment) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.AuthorName == "" {
		c.AuthorName = "Ẩn danh"
	}
	if c.Status == "" {
		c.Status = CommentStatusApproved
	}
	return
}
