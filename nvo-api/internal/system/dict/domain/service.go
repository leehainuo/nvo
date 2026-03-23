package domain

// DictService 字典服务接口
type DictService interface {
	// 字典类型管理
	CreateDictType(req *CreateDictTypeRequest) (*DictType, error)
	GetDictTypeByID(id uint) (*DictTypeResponse, error)
	GetDictTypeByType(dictType string) (*DictTypeResponse, error)
	UpdateDictType(id uint, req *UpdateDictTypeRequest) error
	DeleteDictType(id uint) error
	ListDictTypes(req *ListDictTypeRequest) ([]*DictTypeResponse, int64, error)

	// 字典数据管理
	CreateDictData(req *CreateDictDataRequest) (*DictData, error)
	GetDictDataByID(id uint) (*DictDataResponse, error)
	UpdateDictData(id uint, req *UpdateDictDataRequest) error
	DeleteDictData(id uint) error
	ListDictData(req *ListDictDataRequest) ([]*DictDataResponse, int64, error)
	GetDictDataByType(dictType string) ([]*DictDataResponse, error)
}
