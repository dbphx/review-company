package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/review-company/backend/internal/db"
	"github.com/review-company/backend/internal/model"
	"github.com/review-company/backend/internal/service"
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
	companyService  service.CompanyService
	dataModeService service.DataModeService
}

func NewCompanyHandler(companyService service.CompanyService, dataModeService service.DataModeService) *CompanyHandler {
	return &CompanyHandler{companyService: companyService, dataModeService: dataModeService}
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
	order := strings.ToLower(strings.TrimSpace(c.Query("order", "")))
	if order == "asc" || order == "desc" {
		seedVersion := ""
		if h.dataModeService != nil {
			mode, err := h.dataModeService.GetMode()
			if err == nil && mode != "all" {
				seedVersion = mode
			}
		}

		rows, err := h.companyService.GetTopCompaniesByRating(limit, order, seedVersion)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{
			"data":  rows,
			"total": len(rows),
			"page":  1,
			"limit": limit,
		})
	}

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
	_, totalCompanies, err := h.companyService.GetTopCompanies(1, 1)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	seedVersion := ""
	if h.dataModeService != nil {
		mode, err := h.dataModeService.GetMode()
		if err == nil && mode != "all" {
			seedVersion = mode
		}
	}

	reviewStatsQuery := db.DB.Model(&model.Review{}).
		Select("COUNT(id) as total_reviews, COALESCE(AVG(rating), 0) as avg_rating").
		Where("status = ?", model.StatusApproved)
	if seedVersion == "v2" {
		reviewStatsQuery = reviewStatsQuery.Where("seed_version IN ?", []string{"v2", "live"})
	} else if seedVersion != "" {
		reviewStatsQuery = reviewStatsQuery.Where("seed_version = ?", seedVersion)
	}

	var stats struct {
		TotalReviews int64
		AvgRating    float64
	}
	if err := reviewStatsQuery.Scan(&stats).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"total_companies": totalCompanies,
		"total_reviews":   stats.TotalReviews,
		"avg_rating":      stats.AvgRating,
	})
}

func (h *CompanyHandler) GetCompanyByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id không hợp lệ"})
	}

	company, err := h.companyService.GetCompanyByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "không tìm thấy công ty"})
	}

	seedVersion := ""
	if h.dataModeService != nil {
		mode, err := h.dataModeService.GetMode()
		if err == nil && mode != "all" {
			seedVersion = mode
		}
	}

	statsQuery := db.DB.Model(&model.Review{}).
		Select("COALESCE(AVG(rating), 0) as avg_rating, COUNT(id) as total").
		Where("company_id = ? AND status = ?", id, model.StatusApproved)
	if seedVersion != "" {
		statsQuery = statsQuery.Where("seed_version = ?", seedVersion)
	}

	var stats struct {
		AvgRating float64
		Total     int64
	}
	if err := statsQuery.Scan(&stats).Error; err == nil {
		company.AvgRating = stats.AvgRating
		company.TotalReviews = int(stats.Total)
	}

	return c.JSON(company)
}

func (h *CompanyHandler) CreateCompany(c *fiber.Ctx) error {
	var req upsertCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "dữ liệu gửi lên không hợp lệ"})
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "tên công ty là bắt buộc"})
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
		if err == service.ErrDuplicateCompany {
			return c.Status(409).JSON(fiber.Map{"error": "công ty đã tồn tại"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(company)
}

func (h *CompanyHandler) UpdateCompany(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id không hợp lệ"})
	}

	var req upsertCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "dữ liệu gửi lên không hợp lệ"})
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "tên công ty là bắt buộc"})
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
		return c.Status(404).JSON(fiber.Map{"error": "không tìm thấy công ty"})
	}

	return c.JSON(company)
}

func (h *CompanyHandler) DeleteCompany(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id không hợp lệ"})
	}

	err = h.companyService.DeleteCompany(id)
	if err != nil {
		if err == service.ErrCompanyHasReviews {
			return c.Status(409).JSON(fiber.Map{"error": "không thể xóa công ty đã có review"})
		}
		return c.Status(404).JSON(fiber.Map{"error": "không tìm thấy công ty"})
	}

	return c.JSON(fiber.Map{"success": true})
}
