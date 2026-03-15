package domain

import (
	"time"

	"gorm.io/gorm"
)

type AuditLog struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UserID    uint           `gorm:"index" json:"user_id"`
	Username  string         `gorm:"size:50" json:"username"`
	Module    string         `gorm:"size:50;index" json:"module"`
	Action    string         `gorm:"size:50" json:"action"`
	Method    string         `gorm:"size:10" json:"method"`
	Path      string         `gorm:"size:200" json:"path"`
	IP        string         `gorm:"size:50" json:"ip"`
	UserAgent string         `gorm:"size:500" json:"user_agent"`
	ReqBody   string         `gorm:"type:text" json:"req_body"`
	RespBody  string         `gorm:"type:text" json:"resp_body"`
	Status    int            `json:"status"`
	ErrorMsg  string         `gorm:"type:text" json:"error_msg"`
	Duration  int64          `json:"duration"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AuditLog) TableName() string {
	return "sys_audit_logs"
}

type CreateAuditLogRequest struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Module    string `json:"module"`
	Action    string `json:"action"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	ReqBody   string `json:"req_body"`
	RespBody  string `json:"resp_body"`
	Status    int    `json:"status"`
	ErrorMsg  string `json:"error_msg"`
	Duration  int64  `json:"duration"`
}

type ListAuditLogRequest struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	UserID    *uint  `form:"user_id"`
	Username  string `form:"username"`
	Module    string `form:"module"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
}
