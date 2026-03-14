package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/review-company/backend/internal/model"
	"gorm.io/gorm"
)

type DailyReviewCount struct {
	Day     time.Time `json:"day"`
	Reviews int64     `json:"reviews"`
}

type ReviewRepository interface {
	FindByCompanyID(companyID uuid.UUID, page, limit int, status, seedVersion string) ([]model.Review, int64, error)
	FindByID(id uuid.UUID) (*model.Review, error)
	FindRecent(limit int, status, seedVersion string) ([]model.Review, error)
	Create(review *model.Review) error
	UpdateStatus(id uuid.UUID, status model.ReviewStatus) error
	FindAll(page, limit int, status, companyQuery, seedVersion, createdDate string) ([]model.Review, int64, error)
	MarkAllAsSeedVersion(seedVersion string) error
	RecalculateCompanyStats(companyID uuid.UUID) error
	GetDailyReviewCounts(days int, seedVersion string) ([]DailyReviewCount, error)
}

type reviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) FindByCompanyID(companyID uuid.UUID, page, limit int, status, seedVersion string) ([]model.Review, int64, error) {
	var reviews []model.Review
	var total int64
	offset := (page - 1) * limit

	query := r.db.Model(&model.Review{}).Where("company_id = ?", companyID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query = applySeedVersionFilter(query, seedVersion)

	query.Count(&total)
	err := query.
		Select("reviews.*, (SELECT COUNT(1) FROM comments WHERE comments.review_id = reviews.id AND comments.status = ?) AS comment_count, (SELECT COUNT(1) FROM review_votes rv WHERE rv.review_id = reviews.id AND rv.vote_type = ?) AS like_count, (SELECT COUNT(1) FROM review_votes rv WHERE rv.review_id = reviews.id AND rv.vote_type = ?) AS dislike_count", model.CommentStatusApproved, model.VoteLike, model.VoteDislike).
		Preload("Company").
		Offset(offset).
		Limit(limit).
		Order("created_at desc").
		Find(&reviews).Error
	return reviews, total, err
}

func (r *reviewRepository) FindByID(id uuid.UUID) (*model.Review, error) {
	var review model.Review
	err := r.db.
		Select("reviews.*, (SELECT COUNT(1) FROM review_votes rv WHERE rv.review_id = reviews.id AND rv.vote_type = ?) AS like_count, (SELECT COUNT(1) FROM review_votes rv WHERE rv.review_id = reviews.id AND rv.vote_type = ?) AS dislike_count", model.VoteLike, model.VoteDislike).
		Preload("Company").
		First(&review, "id = ?", id).Error
	return &review, err
}

func (r *reviewRepository) FindRecent(limit int, status, seedVersion string) ([]model.Review, error) {
	var reviews []model.Review
	query := r.db.Model(&model.Review{}).Preload("Company")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query = applySeedVersionFilter(query, seedVersion)
	err := query.
		Select("reviews.*, (SELECT COUNT(1) FROM comments WHERE comments.review_id = reviews.id AND comments.status = ?) AS comment_count, (SELECT COUNT(1) FROM review_votes rv WHERE rv.review_id = reviews.id AND rv.vote_type = ?) AS like_count, (SELECT COUNT(1) FROM review_votes rv WHERE rv.review_id = reviews.id AND rv.vote_type = ?) AS dislike_count", model.CommentStatusApproved, model.VoteLike, model.VoteDislike).
		Order("created_at desc").
		Limit(limit).
		Find(&reviews).Error
	return reviews, err
}

func (r *reviewRepository) Create(review *model.Review) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(review).Error; err != nil {
			return err
		}

		// Recalculate company average rating
		var company model.Company
		if err := tx.First(&company, "id = ?", review.CompanyID).Error; err != nil {
			return err
		}

		var stats struct {
			AvgRating float64
			Total     int
		}
		tx.Model(&model.Review{}).
			Select("COALESCE(AVG(rating), 0) as avg_rating, COUNT(id) as total").
			Where("company_id = ? AND status = ?", review.CompanyID, model.StatusApproved).
			Scan(&stats)

		company.AvgRating = stats.AvgRating
		company.TotalReviews = stats.Total

		return tx.Save(&company).Error
	})
}

func (r *reviewRepository) UpdateStatus(id uuid.UUID, status model.ReviewStatus) error {
	return r.db.Model(&model.Review{}).Where("id = ?", id).Update("status", status).Error
}

func (r *reviewRepository) FindAll(page, limit int, status, companyQuery, seedVersion, createdDate string) ([]model.Review, int64, error) {
	var reviews []model.Review
	var total int64
	offset := (page - 1) * limit

	query := r.db.Model(&model.Review{}).
		Joins("LEFT JOIN companies ON companies.id = reviews.company_id").
		Preload("Company")
	if status != "" {
		query = query.Where("status = ?", status)
	} else {
		query = query.Where("status <> ?", model.StatusDeleted)
	}
	if companyQuery != "" {
		query = query.Where("LOWER(companies.name) LIKE LOWER(?)", "%"+companyQuery+"%")
	}
	query = applySeedVersionFilter(query, seedVersion)
	if createdDate != "" {
		query = query.Where("DATE(reviews.created_at) = ?", createdDate)
	}

	query.Count(&total)
	err := query.
		Select("reviews.*, (SELECT COUNT(1) FROM comments WHERE comments.review_id = reviews.id AND comments.status = ?) AS comment_count, (SELECT COUNT(1) FROM review_votes rv WHERE rv.review_id = reviews.id AND rv.vote_type = ?) AS like_count, (SELECT COUNT(1) FROM review_votes rv WHERE rv.review_id = reviews.id AND rv.vote_type = ?) AS dislike_count", model.CommentStatusApproved, model.VoteLike, model.VoteDislike).
		Preload("Company").
		Offset(offset).
		Limit(limit).
		Order("created_at desc").
		Find(&reviews).Error
	return reviews, total, err
}

func (r *reviewRepository) RecalculateCompanyStats(companyID uuid.UUID) error {
	var stats struct {
		AvgRating float64
		Total     int
	}

	r.db.Model(&model.Review{}).
		Select("COALESCE(AVG(rating), 0) as avg_rating, COUNT(id) as total").
		Where("company_id = ? AND status = ?", companyID, model.StatusApproved).
		Scan(&stats)

	return r.db.Model(&model.Company{}).
		Where("id = ?", companyID).
		Updates(map[string]interface{}{"avg_rating": stats.AvgRating, "total_reviews": stats.Total}).Error
}

func (r *reviewRepository) GetDailyReviewCounts(days int, seedVersion string) ([]DailyReviewCount, error) {
	if days <= 0 {
		days = 7
	}

	var rows []DailyReviewCount
	query := `
	WITH date_series AS (
	  SELECT generate_series(
	    CURRENT_DATE - (($1::int) - 1),
	    CURRENT_DATE,
	    interval '1 day'
	  )::date AS day
	)
	SELECT ds.day AS day, COALESCE(COUNT(r.id), 0) AS reviews
	FROM date_series ds
	LEFT JOIN reviews r
	  ON DATE(r.created_at) = ds.day
	 AND r.status = $2
	 AND (
	   $3 = ''
	   OR r.seed_version = $3
	 )
	GROUP BY ds.day
	ORDER BY ds.day ASC
	`

	err := r.db.Raw(query, days, model.StatusApproved, seedVersion).Scan(&rows).Error
	return rows, err
}

func applySeedVersionFilter(query *gorm.DB, seedVersion string) *gorm.DB {
	if seedVersion == "" {
		return query
	}
	return query.Where("seed_version = ?", seedVersion)
}

func (r *reviewRepository) MarkAllAsSeedVersion(seedVersion string) error {
	return r.db.Model(&model.Review{}).
		Where("seed_version IS NULL OR seed_version = '' OR seed_version = ?", "live").
		Update("seed_version", seedVersion).Error
}
