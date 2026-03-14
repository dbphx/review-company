package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/review-company/backend/internal/service"
)

type AnalyticsHandler struct {
	analyticsService service.AnalyticsService
}

func NewAnalyticsHandler(analyticsService service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsService: analyticsService}
}

func (h *AnalyticsHandler) TrackVisit(c *fiber.Ctx) error {
	type request struct {
		Path string `json:"path"`
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "dữ liệu gửi lên không hợp lệ"})
	}

	path := strings.TrimSpace(req.Path)
	if path == "" {
		path = c.Path()
	}

	sessionID := strings.TrimSpace(c.Get("X-Session-ID"))
	if sessionID == "" {
		sessionID = strings.TrimSpace(c.IP())
	}
	ipAddress := strings.TrimSpace(c.IP())

	if err := h.analyticsService.TrackVisit(path, sessionID, ipAddress); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *AnalyticsHandler) GetDailyVisits(c *fiber.Ctx) error {
	days := c.QueryInt("days", 7)
	data, err := h.analyticsService.GetDailyVisits(days)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": data})
}

func (h *AnalyticsHandler) GetMonthlyVisits(c *fiber.Ctx) error {
	months := c.QueryInt("months", 6)
	data, err := h.analyticsService.GetMonthlyVisits(months)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": data})
}
