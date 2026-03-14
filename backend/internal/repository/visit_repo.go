package repository

import (
	"time"

	"github.com/review-company/backend/internal/model"
	"gorm.io/gorm"
)

type DailyVisitCount struct {
	Day    time.Time `json:"day"`
	Visits int64     `json:"visits"`
}

type MonthlyVisitCount struct {
	Month  string `json:"month"`
	Visits int64  `json:"visits"`
}

type VisitRepository interface {
	Create(path, sessionID, ipAddress string) error
	GetDailyVisitCounts(days int) ([]DailyVisitCount, error)
	GetMonthlyVisitCounts(months int) ([]MonthlyVisitCount, error)
}

type visitRepository struct {
	db *gorm.DB
}

func NewVisitRepository(db *gorm.DB) VisitRepository {
	return &visitRepository{db: db}
}

func (r *visitRepository) Create(path, sessionID, ipAddress string) error {
	visit := model.Visit{Path: path, SessionID: sessionID, IPAddress: ipAddress}
	return r.db.Create(&visit).Error
}

func (r *visitRepository) GetDailyVisitCounts(days int) ([]DailyVisitCount, error) {
	if days <= 0 {
		days = 7
	}

	rows := make([]DailyVisitCount, 0)
	query := `
	WITH date_series AS (
	  SELECT generate_series(
	    CURRENT_DATE - (($1::int) - 1),
	    CURRENT_DATE,
	    interval '1 day'
	  )::date AS day
	)
	SELECT ds.day AS day,
	       COALESCE(COUNT(DISTINCT (
	         to_char(date_trunc('hour', v.created_at), 'YYYY-MM-DD HH24') || '|' ||
	         COALESCE(NULLIF(v.ip_address, ''), v.session_id)
	       )), 0) AS visits
	FROM date_series ds
	LEFT JOIN visits v ON DATE(v.created_at) = ds.day
	GROUP BY ds.day
	ORDER BY ds.day ASC
	`

	err := r.db.Raw(query, days).Scan(&rows).Error
	return rows, err
}

func (r *visitRepository) GetMonthlyVisitCounts(months int) ([]MonthlyVisitCount, error) {
	if months <= 0 {
		months = 6
	}

	rows := make([]MonthlyVisitCount, 0)
	query := `
	WITH month_series AS (
	  SELECT to_char(date_trunc('month', CURRENT_DATE) - (interval '1 month' * s.i), 'YYYY-MM') AS month
	  FROM generate_series(0, ($1::int) - 1) AS s(i)
	)
	SELECT ms.month,
	       COALESCE(COUNT(DISTINCT (
	         to_char(date_trunc('hour', v.created_at), 'YYYY-MM-DD HH24') || '|' ||
	         COALESCE(NULLIF(v.ip_address, ''), v.session_id)
	       )), 0) AS visits
	FROM month_series ms
	LEFT JOIN visits v ON to_char(date_trunc('month', v.created_at), 'YYYY-MM') = ms.month
	GROUP BY ms.month
	ORDER BY ms.month ASC
	`

	err := r.db.Raw(query, months).Scan(&rows).Error
	return rows, err
}
