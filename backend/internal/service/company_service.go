package service

import (
	"github.com/google/uuid"
	"github.com/review/backend/internal/model"
	"github.com/review/backend/internal/repository"
)

type CompanyService interface {
	SearchCompanies(query string) ([]model.Company, error)
	GetTopCompanies(page, limit int) ([]model.Company, int64, error)
	GetCompanyByID(id uuid.UUID) (*model.Company, error)
}

type companyService struct {
	repo repository.CompanyRepository
}

func NewCompanyService(repo repository.CompanyRepository) CompanyService {
	return &companyService{repo: repo}
}

func (s *companyService) SearchCompanies(query string) ([]model.Company, error) {
	return s.repo.Search(query)
}

func (s *companyService) GetTopCompanies(page, limit int) ([]model.Company, int64, error) {
	return s.repo.FindAll(page, limit)
}

func (s *companyService) GetCompanyByID(id uuid.UUID) (*model.Company, error) {
	return s.repo.FindByID(id)
}
