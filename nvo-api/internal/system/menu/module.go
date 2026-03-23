package menu

import (
	"nvo-api/core"
	"nvo-api/internal/system/menu/api"
	"nvo-api/internal/system/menu/domain"
	"nvo-api/internal/system/menu/service"

	"github.com/gin-gonic/gin"
)

type Module struct {
	pocket  *core.Pocket
	handler *api.MenuHandler
	service domain.MenuService
}

func NewModule(pocket *core.Pocket) *Module {
	menuService := service.NewMenuService(pocket.DB)
	menuHandler := api.NewMenuHandler(menuService)

	return &Module{
		pocket:  pocket,
		handler: menuHandler,
		service: menuService,
	}
}

func (m *Module) Service() domain.MenuService {
	return m.service
}

func (m *Module) Name() string {
	return "menu"
}

func (m *Module) Models() []any {
	return []any{
		&domain.Menu{},
	}
}

func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	menus := r.Group("/menus")
	{
		menus.POST("", m.handler.Create)
		menus.GET("", m.handler.GetList)
		menus.GET("/tree", m.handler.GetTree)
		menus.GET("/:id", m.handler.GetByID)
		menus.PUT("/:id", m.handler.Update)
		menus.DELETE("/:id", m.handler.Delete)
	}
}
