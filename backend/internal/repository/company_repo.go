package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/uuid"
	"github.com/review/backend/internal/model"
	"gorm.io/gorm"
)

type CompanyRepository interface {
	FindAll(page, limit int) ([]model.Company, int64, error)
	FindByID(id uuid.UUID) (*model.Company, error)
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

func (r *companyRepository) Create(company *model.Company) error {
	err := r.db.Create(company).Error
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
