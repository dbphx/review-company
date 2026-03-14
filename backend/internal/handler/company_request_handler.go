package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/review-company/backend/internal/db"
	"github.com/review-company/backend/internal/model"
	"github.com/review-company/backend/internal/service"
)

type createCompanyRequestInput struct {
	Name        string `json:"name"`
	LogoURL     string `json:"logo_url"`
	Website     string `json:"website"`
	Industry    string `json:"industry"`
	Size        string `json:"size"`
	Description string `json:"description"`
}

type updateCompanyRequestStatusInput struct {
	Status      string `json:"status"`
	Name        string `json:"name"`
	LogoURL     string `json:"logo_url"`
	Website     string `json:"website"`
	Industry    string `json:"industry"`
	Size        string `json:"size"`
	Description string `json:"description"`
}

type CompanyRequestHandler struct {
	companyService service.CompanyService
}

func NewCompanyRequestHandler(companyService service.CompanyService) *CompanyRequestHandler {
	return &CompanyRequestHandler{companyService: companyService}
}

func (h *CompanyRequestHandler) Create(c *fiber.Ctx) error {
	var input createCompanyRequestInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "dữ liệu gửi lên không hợp lệ"})
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "tên công ty là bắt buộc"})
	}

	req := &model.CompanyRequest{
		Name:        name,
		LogoURL:     strings.TrimSpace(input.LogoURL),
		Website:     strings.TrimSpace(input.Website),
		Industry:    strings.TrimSpace(input.Industry),
		Size:        strings.TrimSpace(input.Size),
		Description: strings.TrimSpace(input.Description),
		RequestedIP: c.IP(),
		Status:      model.CompanyRequestPending,
	}

	if err := db.DB.Create(req).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(req)
}

func (h *CompanyRequestHandler) List(c *fiber.Ctx) error {
	status := strings.ToUpper(strings.TrimSpace(c.Query("status", "PENDING")))
	query := db.DB.Model(&model.CompanyRequest{})
	if status != "" && status != "ALL" {
		query = query.Where("status = ?", status)
	}

	var requests []model.CompanyRequest
	if err := query.Order("created_at desc").Limit(100).Find(&requests).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": requests, "total": len(requests)})
}

func (h *CompanyRequestHandler) UpdateStatus(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id không hợp lệ"})
	}

	var input updateCompanyRequestStatusInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "dữ liệu gửi lên không hợp lệ"})
	}

	status := strings.ToUpper(strings.TrimSpace(input.Status))
	if status != string(model.CompanyRequestApproved) && status != string(model.CompanyRequestRejected) {
		return c.Status(400).JSON(fiber.Map{"error": "status chỉ được là APPROVED hoặc REJECTED"})
	}

	var req model.CompanyRequest
	if err := db.DB.First(&req, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "không tìm thấy yêu cầu"})
	}

	if req.Status == model.CompanyRequestApproved || req.Status == model.CompanyRequestRejected {
		return c.Status(409).JSON(fiber.Map{"error": "yêu cầu đã được xử lý"})
	}

	if status == string(model.CompanyRequestApproved) {
		name := strings.TrimSpace(input.Name)
		if name == "" {
			name = strings.TrimSpace(req.Name)
		}
		if name == "" {
			return c.Status(400).JSON(fiber.Map{"error": "tên công ty là bắt buộc"})
		}

		company := &model.Company{
			Name:        name,
			LogoURL:     strings.TrimSpace(input.LogoURL),
			Website:     strings.TrimSpace(input.Website),
			Industry:    strings.TrimSpace(input.Industry),
			Size:        strings.TrimSpace(input.Size),
			Description: strings.TrimSpace(input.Description),
		}
		if company.LogoURL == "" {
			company.LogoURL = strings.TrimSpace(req.LogoURL)
		}
		if company.Website == "" {
			company.Website = strings.TrimSpace(req.Website)
		}
		if company.Industry == "" {
			company.Industry = strings.TrimSpace(req.Industry)
		}
		if company.Size == "" {
			company.Size = strings.TrimSpace(req.Size)
		}
		if company.Description == "" {
			company.Description = strings.TrimSpace(req.Description)
		}
		if err := h.companyService.CreateCompany(company); err != nil {
			if err == service.ErrDuplicateCompany {
				return c.Status(409).JSON(fiber.Map{"error": "công ty đã tồn tại"})
			}
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
	}

	req.Status = model.CompanyRequestStatus(status)
	if err := db.DB.Save(&req).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(req)
}
