package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/review-company/backend/internal/config"
	"github.com/review-company/backend/internal/db"
	"github.com/review-company/backend/internal/handler"
	"github.com/review-company/backend/internal/middleware"
	"github.com/review-company/backend/internal/repository"
	"github.com/review-company/backend/internal/service"
)

func main() {
	cfg := config.LoadConfig()

	db.ConnectPostgres(cfg)
	db.ConnectElasticsearch(cfg)
	db.ConnectRedis(cfg)

	// Dependency Injection
	companyRepo := repository.NewCompanyRepository(db.DB, db.ES)
	reviewRepo := repository.NewReviewRepository(db.DB)
	commentRepo := repository.NewCommentRepository(db.DB)
	voteRepo := repository.NewVoteRepository(db.DB)
	visitRepo := repository.NewVisitRepository(db.DB)
	settingRepo := repository.NewSystemSettingRepository(db.DB)

	companySvc := service.NewCompanyService(companyRepo)
	analyticsSvc := service.NewAnalyticsService(visitRepo)
	dataModeSvc := service.NewDataModeService(settingRepo)
	reviewSvc := service.NewReviewService(reviewRepo, commentRepo, voteRepo, dataModeSvc)

	companyHandler := handler.NewCompanyHandler(companySvc, dataModeSvc)
	companyRequestHandler := handler.NewCompanyRequestHandler(companySvc)
	reviewHandler := handler.NewReviewHandler(reviewSvc, dataModeSvc)
	authHandler := handler.NewAuthHandler(db.DB)
	adminAuthHandler := handler.NewAdminAuthHandler(db.DB, cfg)
	adminUserHandler := handler.NewAdminUserHandler(db.DB)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsSvc)
	dataModeHandler := handler.NewDataModeHandler(dataModeSvc)
	adminMiddleware := middleware.NewAdminAuthMiddleware(cfg)

	app := fiber.New()
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CorsOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Session-ID",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
	}))

	api := app.Group("/api")

	// Auth
	api.Post("/auth/google", authHandler.GoogleLogin)
	api.Post("/admin/login", adminAuthHandler.Login)
	api.Get("/admin/users", adminMiddleware.RequireAdminOrMod(), adminUserHandler.List)
	api.Get("/admin/sessions", adminMiddleware.RequireAdminOrMod(), adminUserHandler.ListActiveSessions)
	api.Get("/admin/data-mode", adminMiddleware.RequireAdminOrMod(), dataModeHandler.Get)
	api.Get("/admin/search", adminMiddleware.RequireAdminOrMod(), companyHandler.Search)
	api.Get("/admin/companies/top", adminMiddleware.RequireAdminOrMod(), companyHandler.GetTopCompanies)
	api.Get("/admin/companies/stats/summary", adminMiddleware.RequireAdminOrMod(), companyHandler.GetCompanyStats)
	api.Get("/admin/reviews", adminMiddleware.RequireAdminOrMod(), reviewHandler.GetAllReviews)
	api.Get("/admin/reviews/:reviewId", adminMiddleware.RequireAdminOrMod(), reviewHandler.GetReviewByID)
	api.Get("/admin/reviews/:reviewId/comments", adminMiddleware.RequireAdminOrMod(), reviewHandler.GetComments)
	api.Get("/admin/reviews/stats/daily", adminMiddleware.RequireAdminOrMod(), reviewHandler.GetDailyReviewCounts)
	api.Get("/data-mode", dataModeHandler.Get)
	api.Post("/admin/data-mode", adminMiddleware.RequireAdmin(), dataModeHandler.Set)
	api.Post("/admin/users", adminMiddleware.RequireAdmin(), adminUserHandler.Create)
	api.Delete("/admin/users/:id", adminMiddleware.RequireAdmin(), adminUserHandler.Delete)
	api.Delete("/admin/users/:id/sessions", adminMiddleware.RequireAdmin(), adminUserHandler.RevokeModSessions)
	api.Delete("/admin/sessions/:sessionId", adminMiddleware.RequireAdmin(), adminUserHandler.RevokeSessionByID)

	// Search
	api.Get("/search", companyHandler.Search)
	api.Post("/analytics/visits", analyticsHandler.TrackVisit)

	// Companies
	api.Get("/companies/top", companyHandler.GetTopCompanies)
	api.Get("/companies/:id", companyHandler.GetCompanyByID)
	api.Get("/companies/stats/summary", companyHandler.GetCompanyStats)
	api.Post("/company-requests", companyRequestHandler.Create)
	api.Get("/admin/company-requests", adminMiddleware.RequireAdminOrMod(), companyRequestHandler.List)
	api.Patch("/admin/company-requests/:id", adminMiddleware.RequireAdminOrMod(), companyRequestHandler.UpdateStatus)
	api.Post("/companies", adminMiddleware.RequireAdminOrMod(), companyHandler.CreateCompany)
	api.Put("/companies/:id", adminMiddleware.RequireAdminOrMod(), companyHandler.UpdateCompany)
	api.Delete("/companies/:id", adminMiddleware.RequireAdminOrMod(), companyHandler.DeleteCompany)

	// Reviews
	api.Get("/reviews/recent", reviewHandler.GetRecentReviews)
	api.Get("/reviews/stats/daily", reviewHandler.GetDailyReviewCounts)
	api.Get("/analytics/visits/daily", adminMiddleware.RequireAdminOrMod(), analyticsHandler.GetDailyVisits)
	api.Get("/analytics/visits/monthly", adminMiddleware.RequireAdminOrMod(), analyticsHandler.GetMonthlyVisits)
	api.Get("/reviews", reviewHandler.GetAllReviews)
	api.Get("/reviews/:reviewId", reviewHandler.GetReviewByID)
	api.Post("/reviews/:reviewId/vote", reviewHandler.VoteReview)
	api.Delete("/reviews/:reviewId", adminMiddleware.RequireAdminOrMod(), reviewHandler.DeleteReview)
	api.Get("/companies/:id/reviews", reviewHandler.GetCompanyReviews)
	api.Post("/companies/:id/reviews", reviewHandler.CreateReview)

	// Comments
	api.Get("/reviews/:reviewId/comments", reviewHandler.GetComments)
	api.Post("/reviews/:reviewId/comments", reviewHandler.CreateComment)
	api.Post("/comments/:commentId/vote", reviewHandler.VoteComment)
	api.Delete("/comments/:commentId", adminMiddleware.RequireAdminOrMod(), reviewHandler.DeleteComment)

	log.Printf("Server is starting on port %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
