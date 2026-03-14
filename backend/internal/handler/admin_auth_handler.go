package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/review-company/backend/internal/config"
	"github.com/review-company/backend/internal/db"
	"github.com/review-company/backend/internal/model"
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
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
}

func (h *AdminAuthHandler) Login(c *fiber.Ctx) error {
	var req AdminLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Dữ liệu gửi lên không hợp lệ"})
	}
	if req.Email == "" || req.PasswordHash == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Email và mật khẩu băm là bắt buộc"})
	}

	var admin model.AdminUser
	if err := h.db.First(&admin, "email = ?", req.Email).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Email hoặc mật khẩu không đúng"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.PasswordHash)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Email hoặc mật khẩu không đúng"})
	}

	sessionID := uuid.NewString()
	sessionKey := "admin_session:" + sessionID
	userSessionsKey := "admin_user_sessions:" + admin.ID.String()
	sessionData := map[string]string{
		"admin_id":  admin.ID.String(),
		"role":      admin.Role,
		"email":     admin.Email,
		"issued_at": time.Now().Format(time.RFC3339),
	}
	encoded, _ := json.Marshal(sessionData)
	if err := db.RDB.Set(context.Background(), sessionKey, string(encoded), time.Hour).Err(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "không thể tạo phiên đăng nhập"})
	}
	if err := db.RDB.SAdd(context.Background(), userSessionsKey, sessionID).Err(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "không thể lưu phiên đăng nhập"})
	}
	_ = db.RDB.Expire(context.Background(), userSessionsKey, 24*time.Hour).Err()

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["admin_id"] = admin.ID.String()
	claims["email"] = admin.Email
	claims["role"] = admin.Role
	claims["session_id"] = sessionID
	claims["exp"] = time.Now().Add(time.Hour).Unix()

	// Trong thực tế nên lấy Secret từ Config
	t, err := token.SignedString([]byte(h.config.AdminJWTSecret))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Không thể đăng nhập"})
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

func HashPasswordForTransport(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}
