package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/review-company/backend/internal/model"
	"gorm.io/gorm"
)

var ErrDuplicateCompany = errors.New("company already exists")

type CompanyRepository interface {
	FindAll(page, limit int) ([]model.Company, int64, error)
	FindByID(id uuid.UUID) (*model.Company, error)
	FindTopByRating(limit int, order, seedVersion string) ([]model.Company, error)
	Create(company *model.Company) error
	Update(company *model.Company) error
	Delete(id uuid.UUID) error
	Search(query string) ([]model.Company, error)
}

type companyRepository struct {
	db *gorm.DB
	es *elasticsearch.Client
}

func NewCompanyRepository(db *gorm.DB, es *elasticsearch.Client) CompanyRepository {
	return &companyRepository{db: db, es: es}
}

func (r *companyRepository) FindAll(page, limit int) ([]model.Company, int64, error) {
	var companies []model.Company
	var total int64

	offset := (page - 1) * limit
	r.db.Model(&model.Company{}).Count(&total)
	err := r.db.Offset(offset).Limit(limit).Order("avg_rating desc, total_reviews desc").Find(&companies).Error

	return companies, total, err
}

func (r *companyRepository) FindByID(id uuid.UUID) (*model.Company, error) {
	var company model.Company
	err := r.db.First(&company, "id = ?", id).Error
	return &company, err
}

func (r *companyRepository) FindTopByRating(limit int, order, seedVersion string) ([]model.Company, error) {
	if limit <= 0 {
		limit = 5
	}
	if order != "asc" {
		order = "desc"
	}

	stats := r.db.Table("reviews").
		Select("company_id, COALESCE(AVG(rating), 0) as avg_rating, COUNT(id) as total_reviews").
		Where("status = ?", model.StatusApproved)
	if seedVersion != "" {
		stats = stats.Where("seed_version = ?", seedVersion)
	}
	stats = stats.Group("company_id")

	rows := make([]model.Company, 0, limit)
	err := r.db.Table("companies c").
		Select("c.id, c.name, c.logo_url, c.website, c.industry, c.size, c.description, c.created_at, c.updated_at, s.avg_rating, s.total_reviews").
		Joins("JOIN (?) s ON s.company_id = c.id", stats).
		Order("s.avg_rating " + order + ", s.total_reviews desc").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

func (r *companyRepository) Create(company *model.Company) error {
	err := r.db.Create(company).Error
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrDuplicateCompany
		}
		var existing model.Company
		lookupErr := r.db.Where("LOWER(name) = LOWER(?)", company.Name).First(&existing).Error
		if lookupErr == nil {
			return ErrDuplicateCompany
		}
	}
	if err == nil {
		r.indexToES(company)
	}
	return err
}

func (r *companyRepository) Update(company *model.Company) error {
	err := r.db.Save(company).Error
	if err == nil {
		r.indexToES(company)
	}
	return err
}

func (r *companyRepository) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&model.Company{}, "id = ?", id).Error; err != nil {
		return err
	}
	r.deleteFromES(id)
	return nil
}

// Search queries Elasticsearch
func (r *companyRepository) Search(query string) ([]model.Company, error) {
	if query == "" {
		return []model.Company{}, nil
	}

	searchBody := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":     query,
				"fields":    []string{"name^3", "industry", "description"},
				"fuzziness": "AUTO",
			},
		},
		"size": 10,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchBody); err != nil {
		return nil, err
	}

	res, err := r.es.Search(
		r.es.Search.WithContext(context.Background()),
		r.es.Search.WithIndex("companies"),
		r.es.Search.WithBody(&buf),
		r.es.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching ES: %s", res.Status())
	}

	var rMap map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&rMap); err != nil {
		return nil, err
	}

	var companies []model.Company
	hits := rMap["hits"].(map[string]interface{})["hits"].([]interface{})
	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"].(map[string]interface{})

		// Map simple fields back (For MVP)
		idStr := source["id"].(string)
		parsedID, _ := uuid.Parse(idStr)

		companies = append(companies, model.Company{
			ID:       parsedID,
			Name:     source["name"].(string),
			LogoURL:  safeString(source, "logo_url"),
			Industry: safeString(source, "industry"),
		})
	}

	return companies, nil
}

func (r *companyRepository) indexToES(company *model.Company) {
	doc := map[string]interface{}{
		"id":          company.ID.String(),
		"name":        company.Name,
		"logo_url":    company.LogoURL,
		"industry":    company.Industry,
		"description": company.Description,
	}

	payload, _ := json.Marshal(doc)
	res, err := r.es.Index(
		"companies",
		bytes.NewReader(payload),
		r.es.Index.WithDocumentID(company.ID.String()),
	)
	if err != nil {
		log.Printf("Failed to index company %s to ES: %v", company.ID, err)
		return
	}
	defer res.Body.Close()
}

func (r *companyRepository) deleteFromES(id uuid.UUID) {
	res, err := r.es.Delete("companies", id.String())
	if err != nil {
		log.Printf("Failed to delete company %s from ES: %v", id, err)
		return
	}
	defer res.Body.Close()
}

func safeString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok && val != nil {
		return val.(string)
	}
	return ""
}
