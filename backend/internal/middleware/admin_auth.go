package middleware

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/review-company/backend/internal/config"
	"github.com/review-company/backend/internal/db"
)

type AdminAuthMiddleware struct {
	config *config.Config
}

func NewAdminAuthMiddleware(cfg *config.Config) *AdminAuthMiddleware {
	return &AdminAuthMiddleware{config: cfg}
}

func (m *AdminAuthMiddleware) RequireAdmin() fiber.Handler {
	return m.RequireRoles("ADMIN")
}

func (m *AdminAuthMiddleware) RequireAdminOrMod() fiber.Handler {
	return m.RequireRoles("ADMIN", "MOD")
}

func (m *AdminAuthMiddleware) RequireRoles(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "thiếu token quản trị"})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.config.AdminJWTSecret), nil
		})
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "token quản trị không hợp lệ"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "thông tin token không hợp lệ"})
		}

		role, _ := claims["role"].(string)
		sessionID, _ := claims["session_id"].(string)
		adminID, _ := claims["admin_id"].(string)
		if sessionID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "phiên đăng nhập không hợp lệ"})
		}
		sessionRaw, err := db.RDB.Get(context.Background(), "admin_session:"+sessionID).Result()
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "phiên đăng nhập đã hết hạn hoặc bị thu hồi"})
		}
		stored := map[string]string{}
		_ = json.Unmarshal([]byte(sessionRaw), &stored)
		if stored["admin_id"] != adminID || stored["role"] != role {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "phiên đăng nhập không còn hợp lệ"})
		}

		allowed := false
		for _, r := range allowedRoles {
			if role == r {
				allowed = true
				break
			}
		}
		if !allowed {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "bạn không có quyền truy cập chức năng này"})
		}

		c.Locals("admin_claims", claims)
		return c.Next()
	}
}
