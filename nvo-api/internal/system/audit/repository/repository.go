package repository

import (
	"nvo-api/internal/system/audit/domain"
	"time"
	"gorm.io/gorm"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(log *domain.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *AuditRepository) GetByID(id uint) (*domain.AuditLog, error) {
	var log domain.AuditLog
	err := r.db.First(&log, id).Error
	return &log, err
}

func (r *AuditRepository) GetList(req *domain.ListAuditLogRequest) ([]*domain.AuditLog, int64, error) {
	var logs []*domain.AuditLog
	var total int64

	query := r.db.Model(&domain.AuditLog{})

	if req.UserID != nil {
		query = query.Where("user_id = ?", *req.UserID)
	}
	if req.Username != "" {
		query = query.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.Module != "" {
		query = query.Where("module = ?", req.Module)
	}
	if req.StartTime != "" {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	query.Count(&total)

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error

	return logs, total, err
}

func (r *AuditRepository) Delete(id uint) error {
	return r.db.Delete(&domain.AuditLog{}, id).Error
}

func (r *AuditRepository) DeleteOldLogs(before time.Time) error {
	return r.db.Where("created_at < ?", before).Delete(&domain.AuditLog{}).Error
}
