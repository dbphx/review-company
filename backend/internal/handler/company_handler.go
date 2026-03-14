package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/review/backend/internal/model"
	"github.com/review/backend/internal/service"
)

type upsertCompanyRequest struct {
	Name        string `json:"name"`
	LogoURL     string `json:"logo_url"`
	Website     string `json:"website"`
	Industry    string `json:"industry"`
	Size        string `json:"size"`
	Description string `json:"description"`
}

type CompanyHandler struct {
	companyService service.CompanyService
}

func NewCompanyHandler(service service.CompanyService) *CompanyHandler {
	return &CompanyHandler{companyService: service}
}

func (h *CompanyHandler) Search(c *fiber.Ctx) error {
	query := c.Query("q", "")
	companies, err := h.companyService.SearchCompanies(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(companies)
}

func (h *CompanyHandler) GetTopCompanies(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	companies, total, err := h.companyService.GetTopCompanies(page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{
		"data":  companies,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *CompanyHandler) GetCompanyStats(c *fiber.Ctx) error {
	companies, totalCompanies, err := h.companyService.GetTopCompanies(1, 100000)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var totalReviews int64
	var ratingSum float64
	for _, cp := range companies {
		totalReviews += int64(cp.TotalReviews)
		ratingSum += cp.AvgRating
	}

	avgRating := 0.0
	if totalCompanies > 0 {
		avgRating = ratingSum / float64(totalCompanies)
	}

	return c.JSON(fiber.Map{
		"total_companies": totalCompanies,
		"total_reviews":   totalReviews,
		"avg_rating":      avgRating,
	})
}

func (h *CompanyHandler) GetCompanyByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid uuid"})
	}

	company, err := h.companyService.GetCompanyByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "company not found"})
	}
	return c.JSON(company)
}

func (h *CompanyHandler) CreateCompany(c *fiber.Ctx) error {
	var req upsertCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name is required"})
	}

	company := &model.Company{
		Name:        name,
		LogoURL:     strings.TrimSpace(req.LogoURL),
		Website:     strings.TrimSpace(req.Website),
		Industry:    strings.TrimSpace(req.Industry),
		Size:        strings.TrimSpace(req.Size),
		Description: strings.TrimSpace(req.Description),
	}

	if err := h.companyService.CreateCompany(company); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(company)
}

func (h *CompanyHandler) UpdateCompany(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid uuid"})
	}

	var req upsertCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name is required"})
	}

	input := &model.Company{
		Name:        name,
		LogoURL:     strings.TrimSpace(req.LogoURL),
		Website:     strings.TrimSpace(req.Website),
		Industry:    strings.TrimSpace(req.Industry),
		Size:        strings.TrimSpace(req.Size),
		Description: strings.TrimSpace(req.Description),
	}

	company, err := h.companyService.UpdateCompany(id, input)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "company not found"})
	}

	return c.JSON(company)
}

func (h *CompanyHandler) DeleteCompany(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid uuid"})
	}

	err = h.companyService.DeleteCompany(id)
	if err != nil {
		if err == service.ErrCompanyHasReviews {
			return c.Status(409).JSON(fiber.Map{"error": "cannot delete company with existing reviews"})
		}
		return c.Status(404).JSON(fiber.Map{"error": "company not found"})
	}

	return c.JSON(fiber.Map{"success": true})
}
