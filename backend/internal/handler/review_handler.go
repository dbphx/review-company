package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/review/backend/internal/model"
	"github.com/review/backend/internal/service"
)

type ReviewHandler struct {
	reviewService service.ReviewService
}

func NewReviewHandler(service service.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviewService: service}
}

func (h *ReviewHandler) GetCompanyReviews(c *fiber.Ctx) error {
	companyIDParam := c.Params("id")
	companyID, err := uuid.Parse(companyIDParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid company id"})
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

	reviews, total, err := h.reviewService.GetAllReviews(page, limit, status)
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
		return c.Status(400).JSON(fiber.Map{"error": "invalid review id"})
	}

	review, err := h.reviewService.GetReviewByID(reviewID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "review not found"})
	}

	return c.JSON(review)
}

func (h *ReviewHandler) CreateReview(c *fiber.Ctx) error {
	var review model.Review
	if err := c.BodyParser(&review); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	companyIDParam := c.Params("id")
	companyID, err := uuid.Parse(companyIDParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid company id"})
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
		return c.Status(400).JSON(fiber.Map{"error": "invalid review id"})
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
		return c.Status(400).JSON(fiber.Map{"error": "invalid review id"})
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
		return c.Status(400).JSON(fiber.Map{"error": "invalid review id"})
	}

	var comment model.Comment
	if err := c.BodyParser(&comment); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
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
		return c.Status(400).JSON(fiber.Map{"error": "invalid comment id"})
	}

	if err := h.reviewService.DeleteComment(commentID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "comment deleted"})
}
