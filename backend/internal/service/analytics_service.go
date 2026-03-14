package service

import (
	"github.com/review-company/backend/internal/repository"
)

type AnalyticsService interface {
	TrackVisit(path, sessionID, ipAddress string) error
	GetDailyVisits(days int) ([]repository.DailyVisitCount, error)
	GetMonthlyVisits(months int) ([]repository.MonthlyVisitCount, error)
}

type analyticsService struct {
	visitRepo repository.VisitRepository
}

func NewAnalyticsService(visitRepo repository.VisitRepository) AnalyticsService {
	return &analyticsService{visitRepo: visitRepo}
}

func (s *analyticsService) TrackVisit(path, sessionID, ipAddress string) error {
	return s.visitRepo.Create(path, sessionID, ipAddress)
}

func (s *analyticsService) GetDailyVisits(days int) ([]repository.DailyVisitCount, error) {
	return s.visitRepo.GetDailyVisitCounts(days)
}

func (s *analyticsService) GetMonthlyVisits(months int) ([]repository.MonthlyVisitCount, error) {
	return s.visitRepo.GetMonthlyVisitCounts(months)
}
