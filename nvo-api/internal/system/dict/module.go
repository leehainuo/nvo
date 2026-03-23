package dict

import (
	"nvo-api/core"
	"nvo-api/internal/system/dict/api"
	"nvo-api/internal/system/dict/domain"
	"nvo-api/internal/system/dict/service"

	"github.com/gin-gonic/gin"
)

type Module struct {
	pocket  *core.Pocket
	handler *api.DictHandler
	service domain.DictService
}

func NewModule(pocket *core.Pocket) *Module {
	dictService := service.NewDictService(pocket.DB)
	dictHandler := api.NewDictHandler(dictService)

	return &Module{
		pocket:  pocket,
		handler: dictHandler,
		service: dictService,
	}
}

func (m *Module) Service() domain.DictService {
	return m.service
}

func (m *Module) Name() string {
	return "dict"
}

func (m *Module) Models() []any {
	return []any{
		&domain.DictType{},
		&domain.DictData{},
	}
}

func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	// 字典类型管理
	dictTypes := r.Group("/dict/types")
	{
		dictTypes.POST("", m.handler.CreateDictType)
		dictTypes.GET("", m.handler.ListDictTypes)
		dictTypes.GET("/:id", m.handler.GetDictTypeByID)
		dictTypes.PUT("/:id", m.handler.UpdateDictType)
		dictTypes.DELETE("/:id", m.handler.DeleteDictType)
	}

	// 字典数据管理
	dictData := r.Group("/dict/data")
	{
		dictData.POST("", m.handler.CreateDictData)
		dictData.GET("", m.handler.ListDictData)
		dictData.GET("/:id", m.handler.GetDictDataByID)
		dictData.PUT("/:id", m.handler.UpdateDictData)
		dictData.DELETE("/:id", m.handler.DeleteDictData)
	}

	// 根据字典类型获取数据（公开接口）
	r.GET("/dict/type/:type", m.handler.GetDictDataByType)
}
