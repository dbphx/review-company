package repository

import (
	"github.com/review-company/backend/internal/model"
	"gorm.io/gorm"
)

type SystemSettingRepository interface {
	Get(key string) (string, error)
	Set(key, value string) error
}

type systemSettingRepository struct {
	db *gorm.DB
}

func NewSystemSettingRepository(db *gorm.DB) SystemSettingRepository {
	return &systemSettingRepository{db: db}
}

func (r *systemSettingRepository) Get(key string) (string, error) {
	var setting model.SystemSetting
	err := r.db.First(&setting, "key = ?", key).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return setting.Value, nil
}

func (r *systemSettingRepository) Set(key, value string) error {
	var setting model.SystemSetting
	err := r.db.First(&setting, "key = ?", key).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return r.db.Create(&model.SystemSetting{Key: key, Value: value}).Error
		}
		return err
	}
	setting.Value = value
	return r.db.Save(&setting).Error
}
