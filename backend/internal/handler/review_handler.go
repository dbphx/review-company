package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/review-company/backend/internal/db"
	"github.com/review-company/backend/internal/model"
	"github.com/review-company/backend/internal/service"
	"gorm.io/gorm"
)

type voteRequest struct {
	Vote     string `json:"vote"`
	VoteType string `json:"vote_type"`
}

type ReviewHandler struct {
	reviewService   service.ReviewService
	dataModeService service.DataModeService
}

func NewReviewHandler(reviewService service.ReviewService, dataModeService service.DataModeService) *ReviewHandler {
	return &ReviewHandler{reviewService: reviewService, dataModeService: dataModeService}
}

func (h *ReviewHandler) GetCompanyReviews(c *fiber.Ctx) error {
	companyIDParam := c.Params("id")
	companyID, err := uuid.Parse(companyIDParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id công ty không hợp lệ"})
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	reviews, total, err := h.reviewService.GetCompanyReviews(companyID, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"data":  reviews,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *ReviewHandler) GetRecentReviews(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 5)
	reviews, err := h.reviewService.GetRecentReviews(limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": reviews, "limit": limit})
}

func (h *ReviewHandler) GetDailyReviewCounts(c *fiber.Ctx) error {
	days := c.QueryInt("days", 7)
	rows, err := h.reviewService.GetDailyReviewCounts(days)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": rows, "days": days})
}

func (h *ReviewHandler) GetAllReviews(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	status := c.Query("status", "")
	companyQuery := c.Query("company", "")
	createdDate := c.Query("created_date", "")
	seedVersion := c.Query("seed_version", "")
	if seedVersion == "" && h.dataModeService != nil {
		mode, err := h.dataModeService.GetMode()
		if err == nil {
			if mode == "v1" || mode == "v2" {
				seedVersion = mode
			}
		}
	}

	reviews, total, err := h.reviewService.GetAllReviews(page, limit, status, companyQuery, seedVersion, createdDate)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"data":  reviews,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *ReviewHandler) GetReviewByID(c *fiber.Ctx) error {
	reviewIDParam := c.Params("reviewId")
	reviewID, err := uuid.Parse(reviewIDParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id review không hợp lệ"})
	}

	review, err := h.reviewService.GetReviewByID(reviewID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "không tìm thấy review"})
	}

	return c.JSON(review)
}

func (h *ReviewHandler) CreateReview(c *fiber.Ctx) error {
	var review model.Review
	if err := c.BodyParser(&review); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "dữ liệu gửi lên không hợp lệ"})
	}

	companyIDParam := c.Params("id")
	companyID, err := uuid.Parse(companyIDParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id công ty không hợp lệ"})
	}
	review.CompanyID = companyID

	// Simple anti-spam
	review.IPAddress = c.IP()

	if err := h.reviewService.CreateReview(&review); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(review)
}

func (h *ReviewHandler) DeleteReview(c *fiber.Ctx) error {
	reviewIDParam := c.Params("reviewId")
	reviewID, err := uuid.Parse(reviewIDParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id review không hợp lệ"})
	}

	if err := h.reviewService.DeleteReview(reviewID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "review deleted"})
}

func (h *ReviewHandler) GetComments(c *fiber.Ctx) error {
	reviewIDParam := c.Params("reviewId")
	reviewID, err := uuid.Parse(reviewIDParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id review không hợp lệ"})
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	comments, total, err := h.reviewService.GetComments(reviewID, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"data":  comments,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *ReviewHandler) CreateComment(c *fiber.Ctx) error {
	reviewIDParam := c.Params("reviewId")
	reviewID, err := uuid.Parse(reviewIDParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id review không hợp lệ"})
	}

	var comment model.Comment
	if err := c.BodyParser(&comment); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "dữ liệu gửi lên không hợp lệ"})
	}

	comment.ReviewID = reviewID
	comment.IPAddress = c.IP()

	if err := h.reviewService.CreateComment(&comment); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(comment)
}

func (h *ReviewHandler) DeleteComment(c *fiber.Ctx) error {
	commentIDParam := c.Params("commentId")
	commentID, err := uuid.Parse(commentIDParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id bình luận không hợp lệ"})
	}

	if err := h.reviewService.DeleteComment(commentID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "comment deleted"})
}

func (h *ReviewHandler) VoteReview(c *fiber.Ctx) error {
	reviewIDParam := c.Params("reviewId")
	reviewID, err := uuid.Parse(reviewIDParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id review không hợp lệ"})
	}

	var req voteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "dữ liệu gửi lên không hợp lệ"})
	}
	if req.Vote == "" {
		req.Vote = req.VoteType
	}

	voteType, ok := model.ParseVoteType(req.Vote)
	if !ok {
		return c.Status(400).JSON(fiber.Map{"error": "giá trị vote chỉ được là like hoặc dislike"})
	}

	sessionKey := c.Get("X-Session-ID")
	if sessionKey == "" {
		sessionKey = c.IP()
	}

	err = h.reviewService.VoteReview(reviewID, voteType, sessionKey)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"error": "không tìm thấy review"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	review, err := h.reviewService.GetReviewByID(reviewID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"review_id":       reviewID,
		"like_count":      review.LikeCount,
		"dislike_count":   review.DislikeCount,
		"selected_action": voteType,
	})
}

func (h *ReviewHandler) VoteComment(c *fiber.Ctx) error {
	commentIDParam := c.Params("commentId")
	commentID, err := uuid.Parse(commentIDParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id bình luận không hợp lệ"})
	}

	var req voteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "dữ liệu gửi lên không hợp lệ"})
	}
	if req.Vote == "" {
		req.Vote = req.VoteType
	}

	voteType, ok := model.ParseVoteType(req.Vote)
	if !ok {
		return c.Status(400).JSON(fiber.Map{"error": "giá trị vote chỉ được là like hoặc dislike"})
	}

	sessionKey := c.Get("X-Session-ID")
	if sessionKey == "" {
		sessionKey = c.IP()
	}

	err = h.reviewService.VoteComment(commentID, voteType, sessionKey)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"error": "không tìm thấy bình luận"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var comment model.Comment
	if err := db.DB.
		Select("comments.*, (SELECT COUNT(1) FROM comment_votes cv WHERE cv.comment_id = comments.id AND cv.vote_type = ?) AS like_count, (SELECT COUNT(1) FROM comment_votes cv WHERE cv.comment_id = comments.id AND cv.vote_type = ?) AS dislike_count", model.VoteLike, model.VoteDislike).
		First(&comment, "id = ?", commentID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"comment_id":      commentID,
		"like_count":      comment.LikeCount,
		"dislike_count":   comment.DislikeCount,
		"selected_action": voteType,
	})
}
