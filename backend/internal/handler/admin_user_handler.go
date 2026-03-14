package handler

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/review-company/backend/internal/db"
	"github.com/review-company/backend/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminUserHandler struct {
	db *gorm.DB
}

func NewAdminUserHandler(db *gorm.DB) *AdminUserHandler {
	return &AdminUserHandler{db: db}
}

type createAdminUserRequest struct {
	Email        string `json:"email"`
	Name         string `json:"name"`
	PasswordHash string `json:"password_hash"`
	Role         string `json:"role"`
}

type adminSessionItem struct {
	SessionID string `json:"session_id"`
	AdminID   string `json:"admin_id"`
	Role      string `json:"role"`
	Email     string `json:"email"`
	IssuedAt  string `json:"issued_at"`
	TTL       int64  `json:"ttl_seconds"`
}

func (h *AdminUserHandler) List(c *fiber.Ctx) error {
	admins := make([]model.AdminUser, 0)
	if err := h.db.Order("created_at asc").Find(&admins).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": admins, "total": len(admins)})
}

func (h *AdminUserHandler) Create(c *fiber.Ctx) error {
	var req createAdminUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "dữ liệu gửi lên không hợp lệ"})
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.Name)
	passwordHash := strings.TrimSpace(req.PasswordHash)
	role := strings.ToUpper(strings.TrimSpace(req.Role))
	if role == "" {
		role = "MOD"
	}

	if email == "" || name == "" || passwordHash == "" {
		return c.Status(400).JSON(fiber.Map{"error": "email, tên và mật khẩu băm là bắt buộc"})
	}
	if role != "ADMIN" && role != "MOD" {
		return c.Status(400).JSON(fiber.Map{"error": "role chỉ được là ADMIN hoặc MOD"})
	}

	claimsAny := c.Locals("admin_claims")
	claims, _ := claimsAny.(jwt.MapClaims)
	requesterRole, _ := claims["role"].(string)
	if requesterRole != "ADMIN" {
		return c.Status(403).JSON(fiber.Map{"error": "chỉ ADMIN mới được tạo tài khoản quản trị"})
	}

	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(passwordHash), 10)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "không thể tạo mật khẩu"})
	}

	admin := model.AdminUser{
		Email:    email,
		Name:     name,
		Password: string(bcryptHash),
		Role:     role,
	}

	if err := h.db.Create(&admin).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return c.Status(409).JSON(fiber.Map{"error": "email admin đã tồn tại"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"id": admin.ID, "email": admin.Email, "name": admin.Name, "role": admin.Role})
}

func (h *AdminUserHandler) Delete(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id admin không hợp lệ"})
	}

	claimsAny := c.Locals("admin_claims")
	claims, _ := claimsAny.(jwt.MapClaims)
	requesterRole, _ := claims["role"].(string)
	if requesterRole != "ADMIN" {
		return c.Status(403).JSON(fiber.Map{"error": "chỉ ADMIN mới được xóa tài khoản quản trị"})
	}
	if claims != nil {
		if adminIDRaw, ok := claims["admin_id"].(string); ok && adminIDRaw == id.String() {
			return c.Status(400).JSON(fiber.Map{"error": "không thể tự xóa tài khoản đang đăng nhập"})
		}
	}

	var total int64
	if err := h.db.Model(&model.AdminUser{}).Count(&total).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if total <= 1 {
		return c.Status(400).JSON(fiber.Map{"error": "phải giữ lại ít nhất 1 tài khoản admin"})
	}

	res := h.db.Delete(&model.AdminUser{}, "id = ?", id)
	if res.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": res.Error.Error()})
	}
	if res.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "không tìm thấy tài khoản admin"})
	}

	_, _ = revokeAllSessionsByAdminID(id)

	return c.JSON(fiber.Map{"success": true})
}

func (h *AdminUserHandler) RevokeModSessions(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id admin không hợp lệ"})
	}

	var target model.AdminUser
	if err := h.db.First(&target, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "không tìm thấy tài khoản quản trị"})
	}
	if target.Role != "MOD" {
		return c.Status(400).JSON(fiber.Map{"error": "chỉ được thu hồi phiên của tài khoản MOD"})
	}

	revoked, err := revokeAllSessionsByAdminID(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "không thể thu hồi phiên MOD"})
	}

	return c.JSON(fiber.Map{"success": true, "revoked_sessions": revoked})
}

func (h *AdminUserHandler) ListActiveSessions(c *fiber.Ctx) error {
	roleFilter := strings.ToUpper(strings.TrimSpace(c.Query("role", "")))
	emailFilter := strings.ToLower(strings.TrimSpace(c.Query("email", "")))

	keys, err := db.RDB.Keys(context.Background(), "admin_session:*").Result()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "không thể tải danh sách phiên"})
	}

	out := make([]adminSessionItem, 0, len(keys))
	for _, key := range keys {
		raw, err := db.RDB.Get(context.Background(), key).Result()
		if err != nil {
			continue
		}
		data := map[string]string{}
		if err := json.Unmarshal([]byte(raw), &data); err != nil {
			continue
		}

		item := adminSessionItem{
			SessionID: strings.TrimPrefix(key, "admin_session:"),
			AdminID:   data["admin_id"],
			Role:      data["role"],
			Email:     data["email"],
			IssuedAt:  data["issued_at"],
		}

		if roleFilter != "" && item.Role != roleFilter {
			continue
		}
		if emailFilter != "" && !strings.Contains(strings.ToLower(item.Email), emailFilter) {
			continue
		}

		ttl, err := db.RDB.TTL(context.Background(), key).Result()
		if err == nil {
			item.TTL = int64(ttl.Seconds())
		}

		out = append(out, item)
	}

	return c.JSON(fiber.Map{"data": out, "total": len(out)})
}

func (h *AdminUserHandler) RevokeSessionByID(c *fiber.Ctx) error {
	sessionID := strings.TrimSpace(c.Params("sessionId"))
	if sessionID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "session id không hợp lệ"})
	}

	key := "admin_session:" + sessionID
	raw, err := db.RDB.Get(context.Background(), key).Result()
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "không tìm thấy phiên đăng nhập"})
	}

	data := map[string]string{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "dữ liệu phiên không hợp lệ"})
	}
	if data["role"] != "MOD" {
		return c.Status(400).JSON(fiber.Map{"error": "chỉ được thu hồi phiên của MOD"})
	}

	_ = db.RDB.Del(context.Background(), key).Err()
	if adminID := data["admin_id"]; adminID != "" {
		_ = db.RDB.SRem(context.Background(), "admin_user_sessions:"+adminID, sessionID).Err()
	}

	return c.JSON(fiber.Map{"success": true})
}

func revokeAllSessionsByAdminID(adminID uuid.UUID) (int64, error) {
	userSessionsKey := "admin_user_sessions:" + adminID.String()
	ids, err := db.RDB.SMembers(context.Background(), userSessionsKey).Result()
	if err != nil {
		return 0, err
	}
	var revoked int64
	for _, sessionID := range ids {
		if err := db.RDB.Del(context.Background(), "admin_session:"+sessionID).Err(); err == nil {
			revoked++
		}
	}
	_ = db.RDB.Del(context.Background(), userSessionsKey).Err()
	return revoked, nil
}
