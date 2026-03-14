package service

import (
	"github.com/google/uuid"
	"github.com/review-company/backend/internal/model"
	"github.com/review-company/backend/internal/repository"
	"gorm.io/gorm"
)

type ReviewService interface {
	GetCompanyReviews(companyID uuid.UUID, page, limit int) ([]model.Review, int64, error)
	GetRecentReviews(limit int) ([]model.Review, error)
	GetDailyReviewCounts(days int) ([]repository.DailyReviewCount, error)
	GetAllReviews(page, limit int, status, companyQuery, seedVersion, createdDate string) ([]model.Review, int64, error)
	GetReviewByID(reviewID uuid.UUID) (*model.Review, error)
	CreateReview(review *model.Review) error
	DeleteReview(reviewID uuid.UUID) error
	GetComments(reviewID uuid.UUID, page, limit int) ([]model.Comment, int64, error)
	CreateComment(comment *model.Comment) error
	DeleteComment(commentID uuid.UUID) error
	VoteReview(reviewID uuid.UUID, voteType model.VoteType, sessionKey string) error
	VoteComment(commentID uuid.UUID, voteType model.VoteType, sessionKey string) error
	MarkAllAsSeedVersion(seedVersion string) error
}

type reviewService struct {
	reviewRepo      repository.ReviewRepository
	commentRepo     repository.CommentRepository
	voteRepo        repository.VoteRepository
	dataModeService DataModeService
}

func NewReviewService(rRepo repository.ReviewRepository, cRepo repository.CommentRepository, vRepo repository.VoteRepository, dm DataModeService) ReviewService {
	return &reviewService{reviewRepo: rRepo, commentRepo: cRepo, voteRepo: vRepo, dataModeService: dm}
}

func (s *reviewService) GetCompanyReviews(companyID uuid.UUID, page, limit int) ([]model.Review, int64, error) {
	seedVersion := s.getCurrentSeedVersionForRead()
	return s.reviewRepo.FindByCompanyID(companyID, page, limit, string(model.StatusApproved), seedVersion)
}

func (s *reviewService) GetRecentReviews(limit int) ([]model.Review, error) {
	seedVersion := s.getCurrentSeedVersionForRead()
	return s.reviewRepo.FindRecent(limit, string(model.StatusApproved), seedVersion)
}

func (s *reviewService) GetDailyReviewCounts(days int) ([]repository.DailyReviewCount, error) {
	seedVersion := s.getCurrentSeedVersionForRead()
	return s.reviewRepo.GetDailyReviewCounts(days, seedVersion)
}

func (s *reviewService) GetAllReviews(page, limit int, status, companyQuery, seedVersion, createdDate string) ([]model.Review, int64, error) {
	return s.reviewRepo.FindAll(page, limit, status, companyQuery, seedVersion, createdDate)
}

func (s *reviewService) GetReviewByID(reviewID uuid.UUID) (*model.Review, error) {
	return s.reviewRepo.FindByID(reviewID)
}

func (s *reviewService) CreateReview(review *model.Review) error {
	if review.SeedVersion == "" {
		mode, err := s.dataModeService.GetMode()
		if err == nil {
			switch mode {
			case "v1", "v2":
				review.SeedVersion = "v2"
			default:
				review.SeedVersion = "v2"
			}
		}
	}
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

func (s *reviewService) VoteReview(reviewID uuid.UUID, voteType model.VoteType, sessionKey string) error {
	exists, err := s.voteRepo.ReviewExists(reviewID)
	if err != nil {
		return err
	}
	if !exists {
		return gorm.ErrRecordNotFound
	}
	return s.voteRepo.UpsertReviewVote(reviewID, sessionKey, voteType)
}

func (s *reviewService) VoteComment(commentID uuid.UUID, voteType model.VoteType, sessionKey string) error {
	exists, err := s.voteRepo.CommentExists(commentID)
	if err != nil {
		return err
	}
	if !exists {
		return gorm.ErrRecordNotFound
	}
	return s.voteRepo.UpsertCommentVote(commentID, sessionKey, voteType)
}

func (s *reviewService) MarkAllAsSeedVersion(seedVersion string) error {
	return s.reviewRepo.MarkAllAsSeedVersion(seedVersion)
}

func (s *reviewService) getCurrentSeedVersionForRead() string {
	if s.dataModeService == nil {
		return ""
	}
	mode, err := s.dataModeService.GetMode()
	if err != nil {
		return ""
	}
	if mode == "all" {
		return ""
	}
	return mode
}
