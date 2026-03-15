package domain

type AuditService interface {
	Create(req *CreateAuditLogRequest) error
	GetByID(id uint) (*AuditLog, error)
	GetList(req *ListAuditLogRequest) ([]*AuditLog, int64, error)
	Delete(id uint) error
	CleanOldLogs(days int) error
}
