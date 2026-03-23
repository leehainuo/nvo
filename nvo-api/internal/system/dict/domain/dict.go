package domain

import "time"

// DictType 字典类型
type DictType struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	DictName  string    `gorm:"size:100;not null;comment:字典名称" json:"dict_name"`
	DictType  string    `gorm:"size:100;not null;uniqueIndex;comment:字典类型" json:"dict_type"`
	Status    int       `gorm:"default:1;comment:状态(1正常 0停用)" json:"status"`
	Remark    string    `gorm:"size:500;comment:备注" json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DictData 字典数据
type DictData struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	DictSort  int       `gorm:"default:0;comment:字典排序" json:"dict_sort"`
	DictLabel string    `gorm:"size:100;not null;comment:字典标签" json:"dict_label"`
	DictValue string    `gorm:"size:100;not null;comment:字典键值" json:"dict_value"`
	DictType  string    `gorm:"size:100;not null;index;comment:字典类型" json:"dict_type"`
	CSSClass  string    `gorm:"size:100;comment:样式属性" json:"css_class"`
	ListClass string    `gorm:"size:100;comment:表格回显样式" json:"list_class"`
	IsDefault int       `gorm:"default:0;comment:是否默认(1是 0否)" json:"is_default"`
	Status    int       `gorm:"default:1;comment:状态(1正常 0停用)" json:"status"`
	Remark    string    `gorm:"size:500;comment:备注" json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (DictType) TableName() string {
	return "sys_dict_type"
}

// TableName 指定表名
func (DictData) TableName() string {
	return "sys_dict_data"
}

// CreateDictTypeRequest 创建字典类型请求
type CreateDictTypeRequest struct {
	DictName string `json:"dict_name" binding:"required"`
	DictType string `json:"dict_type" binding:"required"`
	Status   int    `json:"status"`
	Remark   string `json:"remark"`
}

// UpdateDictTypeRequest 更新字典类型请求
type UpdateDictTypeRequest struct {
	DictName string `json:"dict_name"`
	Status   int    `json:"status"`
	Remark   string `json:"remark"`
}

// ListDictTypeRequest 字典类型列表请求
type ListDictTypeRequest struct {
	DictName string `form:"dict_name"`
	DictType string `form:"dict_type"`
	Status   *int   `form:"status"`
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"page_size" binding:"required,min=1,max=100"`
}

// DictTypeResponse 字典类型响应
type DictTypeResponse struct {
	ID        uint      `json:"id"`
	DictName  string    `json:"dict_name"`
	DictType  string    `json:"dict_type"`
	Status    int       `json:"status"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse 将 DictType 转换为 DictTypeResponse
func (dt *DictType) ToResponse() *DictTypeResponse {
	return &DictTypeResponse{
		ID:        dt.ID,
		DictName:  dt.DictName,
		DictType:  dt.DictType,
		Status:    dt.Status,
		Remark:    dt.Remark,
		CreatedAt: dt.CreatedAt,
		UpdatedAt: dt.UpdatedAt,
	}
}

// ToDictTypeResponseList 批量转换 DictType 列表
func ToDictTypeResponseList(dictTypes []*DictType) []*DictTypeResponse {
	responses := make([]*DictTypeResponse, 0, len(dictTypes))
	for _, dt := range dictTypes {
		responses = append(responses, dt.ToResponse())
	}
	return responses
}

// CreateDictDataRequest 创建字典数据请求
type CreateDictDataRequest struct {
	DictSort  int    `json:"dict_sort"`
	DictLabel string `json:"dict_label" binding:"required"`
	DictValue string `json:"dict_value" binding:"required"`
	DictType  string `json:"dict_type" binding:"required"`
	CSSClass  string `json:"css_class"`
	ListClass string `json:"list_class"`
	IsDefault int    `json:"is_default"`
	Status    int    `json:"status"`
	Remark    string `json:"remark"`
}

// UpdateDictDataRequest 更新字典数据请求
type UpdateDictDataRequest struct {
	DictSort  int    `json:"dict_sort"`
	DictLabel string `json:"dict_label"`
	DictValue string `json:"dict_value"`
	CSSClass  string `json:"css_class"`
	ListClass string `json:"list_class"`
	IsDefault int    `json:"is_default"`
	Status    int    `json:"status"`
	Remark    string `json:"remark"`
}

// ListDictDataRequest 字典数据列表请求
type ListDictDataRequest struct {
	DictType  string `form:"dict_type"`
	DictLabel string `form:"dict_label"`
	Status    *int   `form:"status"`
	Page      int    `form:"page" binding:"required,min=1"`
	PageSize  int    `form:"page_size" binding:"required,min=1,max=100"`
}

// DictDataResponse 字典数据响应
type DictDataResponse struct {
	ID        uint      `json:"id"`
	DictSort  int       `json:"dict_sort"`
	DictLabel string    `json:"dict_label"`
	DictValue string    `json:"dict_value"`
	DictType  string    `json:"dict_type"`
	CSSClass  string    `json:"css_class"`
	ListClass string    `json:"list_class"`
	IsDefault int       `json:"is_default"`
	Status    int       `json:"status"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse 将 DictData 转换为 DictDataResponse
func (dd *DictData) ToResponse() *DictDataResponse {
	return &DictDataResponse{
		ID:        dd.ID,
		DictSort:  dd.DictSort,
		DictLabel: dd.DictLabel,
		DictValue: dd.DictValue,
		DictType:  dd.DictType,
		CSSClass:  dd.CSSClass,
		ListClass: dd.ListClass,
		IsDefault: dd.IsDefault,
		Status:    dd.Status,
		Remark:    dd.Remark,
		CreatedAt: dd.CreatedAt,
		UpdatedAt: dd.UpdatedAt,
	}
}

// ToDictDataResponseList 批量转换 DictData 列表
func ToDictDataResponseList(dictDataList []*DictData) []*DictDataResponse {
	responses := make([]*DictDataResponse, 0, len(dictDataList))
	for _, dd := range dictDataList {
		responses = append(responses, dd.ToResponse())
	}
	return responses
}
