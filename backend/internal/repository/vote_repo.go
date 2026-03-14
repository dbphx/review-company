package repository

import (
	"github.com/google/uuid"
	"github.com/review/backend/internal/model"
	"gorm.io/gorm"
)

type VoteRepository interface {
	UpsertReviewVote(reviewID uuid.UUID, sessionKey string, voteType model.VoteType) error
	UpsertCommentVote(commentID uuid.UUID, sessionKey string, voteType model.VoteType) error
	ReviewExists(reviewID uuid.UUID) (bool, error)
	CommentExists(commentID uuid.UUID) (bool, error)
}

type voteRepository struct {
	db *gorm.DB
}

func NewVoteRepository(db *gorm.DB) VoteRepository {
	return &voteRepository{db: db}
}

func (r *voteRepository) UpsertReviewVote(reviewID uuid.UUID, sessionKey string, voteType model.VoteType) error {
	var vote model.ReviewVote
	err := r.db.Where("review_id = ? AND session_key = ?", reviewID, sessionKey).First(&vote).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return r.db.Create(&model.ReviewVote{
				ReviewID:   reviewID,
				SessionKey: sessionKey,
				VoteType:   voteType,
			}).Error
		}
		return err
	}

	return r.db.Model(&model.ReviewVote{}).
		Where("id = ?", vote.ID).
		Updates(map[string]interface{}{"vote_type": voteType}).Error
}

func (r *voteRepository) UpsertCommentVote(commentID uuid.UUID, sessionKey string, voteType model.VoteType) error {
	var vote model.CommentVote
	err := r.db.Where("comment_id = ? AND session_key = ?", commentID, sessionKey).First(&vote).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return r.db.Create(&model.CommentVote{
				CommentID:  commentID,
				SessionKey: sessionKey,
				VoteType:   voteType,
			}).Error
		}
		return err
	}

	return r.db.Model(&model.CommentVote{}).
		Where("id = ?", vote.ID).
		Updates(map[string]interface{}{"vote_type": voteType}).Error
}

func (r *voteRepository) ReviewExists(reviewID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Review{}).
		Where("id = ? AND status <> ?", reviewID, model.StatusDeleted).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *voteRepository) CommentExists(commentID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Comment{}).
		Where("id = ? AND status <> ?", commentID, model.CommentStatusDeleted).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
