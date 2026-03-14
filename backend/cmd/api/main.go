package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/review/backend/internal/config"
	"github.com/review/backend/internal/db"
	"github.com/review/backend/internal/handler"
	"github.com/review/backend/internal/middleware"
	"github.com/review/backend/internal/repository"
	"github.com/review/backend/internal/service"
)

func main() {
	cfg := config.LoadConfig()

	db.ConnectPostgres(cfg)
	db.ConnectElasticsearch(cfg)

	// Dependency Injection
	companyRepo := repository.NewCompanyRepository(db.DB, db.ES)
	reviewRepo := repository.NewReviewRepository(db.DB)
	commentRepo := repository.NewCommentRepository(db.DB)
	voteRepo := repository.NewVoteRepository(db.DB)

	companySvc := service.NewCompanyService(companyRepo)
	reviewSvc := service.NewReviewService(reviewRepo, commentRepo, voteRepo)

	companyHandler := handler.NewCompanyHandler(companySvc)
	reviewHandler := handler.NewReviewHandler(reviewSvc)
	authHandler := handler.NewAuthHandler(db.DB)
	adminAuthHandler := handler.NewAdminAuthHandler(db.DB, cfg)
	adminMiddleware := middleware.NewAdminAuthMiddleware(cfg)

	app := fiber.New()
	app.Use(logger.New())
	app.Use(cors.New())

	api := app.Group("/api")

	// Auth
	api.Post("/auth/google", authHandler.GoogleLogin)
	api.Post("/admin/login", adminAuthHandler.Login)

	// Search
	api.Get("/search", companyHandler.Search)

	// Companies
	api.Get("/companies/top", companyHandler.GetTopCompanies)
	api.Get("/companies/:id", companyHandler.GetCompanyByID)
	api.Get("/companies/stats/summary", companyHandler.GetCompanyStats)
	api.Post("/companies", adminMiddleware.RequireAdmin(), companyHandler.CreateCompany)
	api.Put("/companies/:id", adminMiddleware.RequireAdmin(), companyHandler.UpdateCompany)
	api.Delete("/companies/:id", adminMiddleware.RequireAdmin(), companyHandler.DeleteCompany)

	// Reviews
	api.Get("/reviews/recent", reviewHandler.GetRecentReviews)
	api.Get("/reviews/stats/daily", reviewHandler.GetDailyReviewCounts)
	api.Get("/reviews", reviewHandler.GetAllReviews)
	api.Get("/reviews/:reviewId", reviewHandler.GetReviewByID)
	api.Post("/reviews/:reviewId/vote", reviewHandler.VoteReview)
	api.Delete("/reviews/:reviewId", adminMiddleware.RequireAdmin(), reviewHandler.DeleteReview)
	api.Get("/companies/:id/reviews", reviewHandler.GetCompanyReviews)
	api.Post("/companies/:id/reviews", reviewHandler.CreateReview)

	// Comments
	api.Get("/reviews/:reviewId/comments", reviewHandler.GetComments)
	api.Post("/reviews/:reviewId/comments", reviewHandler.CreateComment)
	api.Post("/comments/:commentId/vote", reviewHandler.VoteComment)
	api.Delete("/comments/:commentId", adminMiddleware.RequireAdmin(), reviewHandler.DeleteComment)

	log.Printf("Server is starting on port %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
