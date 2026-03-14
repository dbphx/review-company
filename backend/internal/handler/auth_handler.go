package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/review/backend/internal/model"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

type GoogleTokenRequest struct {
	AccessToken string `json:"access_token"`
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

func (h *AuthHandler) GoogleLogin(c *fiber.Ctx) error {
	var req GoogleTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Lấy thông tin user từ Google
	userInfoReq, err := http.NewRequestWithContext(context.Background(), "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create request to Google"})
	}
	userInfoReq.Header.Set("Authorization", "Bearer "+req.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(userInfoReq)
	if err != nil || resp.StatusCode != http.StatusOK {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid Google token"})
	}
	defer resp.Body.Close()

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to decode Google user info"})
	}

	// Tìm hoặc tạo User trong DB
	var user model.User
	err = h.db.Preload("AllowedCompanies").First(&user, "email = ?", userInfo.Email).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			user = model.User{
				Email:    userInfo.Email,
				Auth0ID:  userInfo.ID,
				Name:     userInfo.Name,
				Picture:  userInfo.Picture,
				Provider: "google",
				Role:     "MODERATOR", // Mặc định là MODERATOR (hoặc GUEST)
			}
			if err := h.db.Create(&user).Error; err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to create user"})
			}
		} else {
			return c.Status(500).JSON(fiber.Map{"error": "Database error"})
		}
	} else {
		// Update thông tin mới nhất
		user.Name = userInfo.Name
		user.Picture = userInfo.Picture
		h.db.Save(&user)
	}

	// Trả về mock token (Nên dùng JWT ở production)
	return c.JSON(fiber.Map{
		"token": "mock-jwt-token-for-" + user.ID.String(),
		"user":  user,
	})
}
