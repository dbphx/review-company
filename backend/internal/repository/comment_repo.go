package repository

import (
	"github.com/google/uuid"
	"github.com/review-company/backend/internal/model"
	"gorm.io/gorm"
)

type CommentRepository interface {
	FindByReviewID(reviewID uuid.UUID, page, limit int) ([]model.Comment, int64, error)
	Create(comment *model.Comment) error
	UpdateStatus(id uuid.UUID, status model.CommentStatus) error
	DeleteThread(id uuid.UUID) error
	DeleteByReviewID(reviewID uuid.UUID) error
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) FindByReviewID(reviewID uuid.UUID, page, limit int) ([]model.Comment, int64, error) {
	var comments []model.Comment
	var total int64
	offset := (page - 1) * limit

	query := r.db.Model(&model.Comment{}).Where("review_id = ? AND status = ?", reviewID, model.CommentStatusApproved)
	query.Count(&total)
	err := query.
		Select("comments.*, (SELECT COUNT(1) FROM comment_votes cv WHERE cv.comment_id = comments.id AND cv.vote_type = ?) AS like_count, (SELECT COUNT(1) FROM comment_votes cv WHERE cv.comment_id = comments.id AND cv.vote_type = ?) AS dislike_count", model.VoteLike, model.VoteDislike).
		Offset(offset).
		Limit(limit).
		Order("created_at asc").
		Find(&comments).Error

	return comments, total, err
}

func (r *commentRepository) Create(comment *model.Comment) error {
	return r.db.Create(comment).Error
}

func (r *commentRepository) UpdateStatus(id uuid.UUID, status model.CommentStatus) error {
	return r.db.Model(&model.Comment{}).Where("id = ?", id).Update("status", status).Error
}

func (r *commentRepository) DeleteThread(id uuid.UUID) error {
	query := `
	WITH RECURSIVE comment_tree AS (
		SELECT id FROM comments WHERE id = ?
		UNION ALL
		SELECT c.id
		FROM comments c
		JOIN comment_tree ct ON c.parent_comment_id = ct.id
	)
	UPDATE comments
	SET status = ?, deleted_at = NOW(), updated_at = NOW()
	WHERE id IN (SELECT id FROM comment_tree)
	`
	return r.db.Exec(query, id, model.CommentStatusDeleted).Error
}

func (r *commentRepository) DeleteByReviewID(reviewID uuid.UUID) error {
	return r.db.Model(&model.Comment{}).
		Where("review_id = ?", reviewID).
		Updates(map[string]interface{}{"status": model.CommentStatusDeleted, "deleted_at": gorm.Expr("NOW()")}).Error
}
