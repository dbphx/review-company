package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/review/backend/internal/config"
	"github.com/review/backend/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminAuthHandler struct {
	db     *gorm.DB
	config *config.Config
}

func NewAdminAuthHandler(db *gorm.DB, cfg *config.Config) *AdminAuthHandler {
	return &AdminAuthHandler{db: db, config: cfg}
}

type AdminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AdminAuthHandler) Login(c *fiber.Ctx) error {
	var req AdminLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var admin model.AdminUser
	if err := h.db.First(&admin, "email = ?", req.Email).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["admin_id"] = admin.ID
	claims["email"] = admin.Email
	claims["role"] = admin.Role
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// Trong thực tế nên lấy Secret từ Config
	t, err := token.SignedString([]byte(h.config.AdminJWTSecret))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not login"})
	}

	return c.JSON(fiber.Map{
		"token": t,
		"admin": fiber.Map{
			"id":    admin.ID,
			"email": admin.Email,
			"name":  admin.Name,
			"role":  admin.Role,
		},
	})
}
