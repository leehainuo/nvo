package service

import (
	"errors"
	"nvo-api/internal/system/audit/domain"
	"nvo-api/internal/system/audit/repository"
	"time"
	"gorm.io/gorm"
)

type AuditService struct {
	db   *gorm.DB
	repo *repository.AuditRepository
}

func NewAuditService(db *gorm.DB) domain.AuditService {
	return &AuditService{
		db:   db,
		repo: repository.NewAuditRepository(db),
	}
}

func (s *AuditService) Create(req *domain.CreateAuditLogRequest) error {
	log := &domain.AuditLog{
		UserID:    req.UserID,
		Username:  req.Username,
		Module:    req.Module,
		Action:    req.Action,
		Method:    req.Method,
		Path:      req.Path,
		IP:        req.IP,
		UserAgent: req.UserAgent,
		ReqBody:   req.ReqBody,
		RespBody:  req.RespBody,
		Status:    req.Status,
		ErrorMsg:  req.ErrorMsg,
		Duration:  req.Duration,
	}

	return s.repo.Create(log)
}

func (s *AuditService) GetByID(id uint) (*domain.AuditLog, error) {
	log, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("审计日志不存在")
		}
		return nil, err
	}
	return log, nil
}

func (s *AuditService) GetList(req *domain.ListAuditLogRequest) ([]*domain.AuditLog, int64, error) {
	return s.repo.GetList(req)
}

func (s *AuditService) Delete(id uint) error {
	if _, err := s.repo.GetByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("审计日志不存在")
		}
		return err
	}

	return s.repo.Delete(id)
}

func (s *AuditService) CleanOldLogs(days int) error {
	if days <= 0 {
		return errors.New("天数必须大于0")
	}

	before := time.Now().AddDate(0, 0, -days)
	return s.repo.DeleteOldLogs(before)
}
