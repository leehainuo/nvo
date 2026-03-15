package audit

import (
	"nvo-api/core"
	"nvo-api/internal/system/audit/api"
	"nvo-api/internal/system/audit/domain"
	"nvo-api/internal/system/audit/service"
	"github.com/gin-gonic/gin"
)

type Module struct {
	pocket  *core.Pocket
	service domain.AuditService
	handler *api.AuditHandler
}

func NewModule(pocket *core.Pocket) *Module {
	auditService := service.NewAuditService(pocket.DB)
	auditHandler := api.NewAuditHandler(auditService)

	return &Module{
		pocket:  pocket,
		service: auditService,
		handler: auditHandler,
	}
}

func (m *Module) Service() domain.AuditService {
	return m.service
}

func (m *Module) Name() string {
	return "audit"
}

func (m *Module) Models() []any {
	return []any{
		&domain.AuditLog{},
	}
}

func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	audits := r.Group("/audit-logs")
	{
		audits.POST("", m.handler.Create)
		audits.GET("", m.handler.GetList)
		audits.GET("/:id", m.handler.GetByID)
		audits.DELETE("/:id", m.handler.Delete)
		audits.POST("/clean", m.handler.CleanOldLogs)
	}
}
