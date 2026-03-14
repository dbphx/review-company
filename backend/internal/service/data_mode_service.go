package service

import (
	"strings"

	"github.com/review-company/backend/internal/repository"
)

const DataModeSettingKey = "admin_data_mode"

type DataModeService interface {
	GetMode() (string, error)
	SetMode(mode string) (string, error)
}

type dataModeService struct {
	repo repository.SystemSettingRepository
}

func NewDataModeService(repo repository.SystemSettingRepository) DataModeService {
	return &dataModeService{repo: repo}
}

func (s *dataModeService) GetMode() (string, error) {
	v, err := s.repo.Get(DataModeSettingKey)
	if err != nil {
		return "v1", err
	}
	v = normalizeMode(v)
	if v == "" {
		v = "v1"
	}
	return v, nil
}

func (s *dataModeService) SetMode(mode string) (string, error) {
	m := normalizeMode(mode)
	if m == "" {
		m = "v1"
	}
	if err := s.repo.Set(DataModeSettingKey, m); err != nil {
		return "", err
	}
	return m, nil
}

func normalizeMode(input string) string {
	v := strings.ToLower(strings.TrimSpace(input))
	if v == "v1" || v == "v2" || v == "all" {
		return v
	}
	return ""
}
