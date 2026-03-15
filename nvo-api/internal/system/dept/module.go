package dept

import (
	"nvo-api/core"
	"nvo-api/internal/system/dept/api"
	"nvo-api/internal/system/dept/domain"
	"nvo-api/internal/system/dept/service"
	"github.com/gin-gonic/gin"
)

type Module struct {
	pocket  *core.Pocket
	service domain.DeptService
	handler *api.DeptHandler
}

func NewModule(pocket *core.Pocket) *Module {
	deptService := service.NewDeptService(pocket.DB)
	deptHandler := api.NewDeptHandler(deptService)

	return &Module{
		pocket:  pocket,
		service: deptService,
		handler: deptHandler,
	}
}

func (m *Module) Service() domain.DeptService {
	return m.service
}

func (m *Module) Name() string {
	return "dept"
}

func (m *Module) Models() []any {
	return []any{
		&domain.Dept{},
	}
}

func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	depts := r.Group("/depts")
	{
		depts.POST("", m.handler.Create)
		depts.GET("", m.handler.GetList)
		depts.GET("/tree", m.handler.GetTree)
		depts.GET("/:id", m.handler.GetByID)
		depts.PUT("/:id", m.handler.Update)
		depts.DELETE("/:id", m.handler.Delete)
	}
}
