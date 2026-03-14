package service

import (
	"github.com/google/uuid"
	"github.com/review/backend/internal/model"
	"github.com/review/backend/internal/repository"
)

type ReviewService interface {
	GetCompanyReviews(companyID uuid.UUID, page, limit int) ([]model.Review, int64, error)
	GetRecentReviews(limit int) ([]model.Review, error)
	GetDailyReviewCounts(days int) ([]repository.DailyReviewCount, error)
	GetAllReviews(page, limit int, status string) ([]model.Review, int64, error)
	GetReviewByID(reviewID uuid.UUID) (*model.Review, error)
	CreateReview(review *model.Review) error
	DeleteReview(reviewID uuid.UUID) error
	GetComments(reviewID uuid.UUID, page, limit int) ([]model.Comment, int64, error)
	CreateComment(comment *model.Comment) error
	DeleteComment(commentID uuid.UUID) error
}

type reviewService struct {
	reviewRepo  repository.ReviewRepository
	commentRepo repository.CommentRepository
}

func NewReviewService(rRepo repository.ReviewRepository, cRepo repository.CommentRepository) ReviewService {
	return &reviewService{reviewRepo: rRepo, commentRepo: cRepo}
}

func (s *reviewService) GetCompanyReviews(companyID uuid.UUID, page, limit int) ([]model.Review, int64, error) {
	return s.reviewRepo.FindByCompanyID(companyID, page, limit, string(model.StatusApproved))
}

func (s *reviewService) GetRecentReviews(limit int) ([]model.Review, error) {
	return s.reviewRepo.FindRecent(limit, string(model.StatusApproved))
}

func (s *reviewService) GetDailyReviewCounts(days int) ([]repository.DailyReviewCount, error) {
	return s.reviewRepo.GetDailyReviewCounts(days)
}

func (s *reviewService) GetAllReviews(page, limit int, status string) ([]model.Review, int64, error) {
	return s.reviewRepo.FindAll(page, limit, status)
}

func (s *reviewService) GetReviewByID(reviewID uuid.UUID) (*model.Review, error) {
	return s.reviewRepo.FindByID(reviewID)
}

func (s *reviewService) CreateReview(review *model.Review) error {
	return s.reviewRepo.Create(review)
}

func (s *reviewService) DeleteReview(reviewID uuid.UUID) error {
	review, err := s.reviewRepo.FindByID(reviewID)
	if err != nil {
		return err
	}

	if err := s.commentRepo.DeleteByReviewID(reviewID); err != nil {
		return err
	}
	if err := s.reviewRepo.UpdateStatus(reviewID, model.StatusDeleted); err != nil {
		return err
	}

	return s.reviewRepo.RecalculateCompanyStats(review.CompanyID)
}

func (s *reviewService) GetComments(reviewID uuid.UUID, page, limit int) ([]model.Comment, int64, error) {
	return s.commentRepo.FindByReviewID(reviewID, page, limit)
}

func (s *reviewService) CreateComment(comment *model.Comment) error {
	return s.commentRepo.Create(comment)
}

func (s *reviewService) DeleteComment(commentID uuid.UUID) error {
	return s.commentRepo.DeleteThread(commentID)
}
