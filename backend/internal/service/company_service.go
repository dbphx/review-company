package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/review-company/backend/internal/model"
	"github.com/review-company/backend/internal/repository"
)

var ErrCompanyHasReviews = errors.New("cannot delete company with existing reviews")
var ErrDuplicateCompany = errors.New("company already exists")

type CompanyService interface {
	SearchCompanies(query string) ([]model.Company, error)
	GetTopCompanies(page, limit int) ([]model.Company, int64, error)
	GetTopCompaniesByRating(limit int, order, seedVersion string) ([]model.Company, error)
	GetTopCompaniesBySort(limit int, sortBy, seedVersion string) ([]model.Company, error)
	GetCompanyByID(id uuid.UUID) (*model.Company, error)
	CreateCompany(company *model.Company) error
	UpdateCompany(id uuid.UUID, input *model.Company) (*model.Company, error)
	DeleteCompany(id uuid.UUID) error
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

func (s *companyService) GetTopCompaniesByRating(limit int, order, seedVersion string) ([]model.Company, error) {
	return s.repo.FindTopByRating(limit, order, seedVersion)
}

func (s *companyService) GetTopCompaniesBySort(limit int, sortBy, seedVersion string) ([]model.Company, error) {
	return s.repo.FindTopBySort(limit, sortBy, seedVersion)
}

func (s *companyService) GetCompanyByID(id uuid.UUID) (*model.Company, error) {
	return s.repo.FindByID(id)
}

func (s *companyService) CreateCompany(company *model.Company) error {
	err := s.repo.Create(company)
	if err == repository.ErrDuplicateCompany {
		return ErrDuplicateCompany
	}
	return err
}

func (s *companyService) UpdateCompany(id uuid.UUID, input *model.Company) (*model.Company, error) {
	company, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	company.Name = input.Name
	company.LogoURL = input.LogoURL
	company.Website = input.Website
	company.Industry = input.Industry
	company.Size = input.Size
	company.Description = input.Description

	if err := s.repo.Update(company); err != nil {
		return nil, err
	}

	return company, nil
}

func (s *companyService) DeleteCompany(id uuid.UUID) error {
	company, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if company.TotalReviews > 0 {
		return ErrCompanyHasReviews
	}

	return s.repo.Delete(id)
}
