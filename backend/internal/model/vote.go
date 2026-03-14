package model

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VoteType string

const (
	VoteLike    VoteType = "LIKE"
	VoteDislike VoteType = "DISLIKE"
)

func ParseVoteType(input string) (VoteType, bool) {
	s := strings.TrimSpace(strings.ToLower(input))
	switch s {
	case "like":
		return VoteLike, true
	case "dislike":
		return VoteDislike, true
	default:
		return "", false
	}
}

type ReviewVote struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	ReviewID   uuid.UUID      `gorm:"type:uuid;not null;index:idx_review_votes_unique,unique" json:"review_id"`
	SessionKey string         `gorm:"type:varchar(255);not null;index:idx_review_votes_unique,unique" json:"-"`
	VoteType   VoteType       `gorm:"type:varchar(20);not null" json:"vote_type"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (v *ReviewVote) BeforeCreate(tx *gorm.DB) (err error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

type CommentVote struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	CommentID  uuid.UUID      `gorm:"type:uuid;not null;index:idx_comment_votes_unique,unique" json:"comment_id"`
	SessionKey string         `gorm:"type:varchar(255);not null;index:idx_comment_votes_unique,unique" json:"-"`
	VoteType   VoteType       `gorm:"type:varchar(20);not null" json:"vote_type"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (v *CommentVote) BeforeCreate(tx *gorm.DB) (err error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}
