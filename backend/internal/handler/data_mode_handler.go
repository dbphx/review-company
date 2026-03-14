package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/review-company/backend/internal/service"
)

type DataModeHandler struct {
	service service.DataModeService
}

func NewDataModeHandler(s service.DataModeService) *DataModeHandler {
	return &DataModeHandler{service: s}
}

func (h *DataModeHandler) Get(c *fiber.Ctx) error {
	mode, err := h.service.GetMode()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"mode": mode})
}

func (h *DataModeHandler) Set(c *fiber.Ctx) error {
	type request struct {
		Mode string `json:"mode"`
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "dữ liệu gửi lên không hợp lệ"})
	}

	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode != "v1" && mode != "v2" && mode != "all" {
		return c.Status(400).JSON(fiber.Map{"error": "mode chỉ được là v1, v2 hoặc all"})
	}

	stored, err := h.service.SetMode(mode)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"mode": stored})
}
