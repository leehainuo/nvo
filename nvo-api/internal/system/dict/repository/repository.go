package repository

import (
	"nvo-api/internal/system/dict/domain"

	"gorm.io/gorm"
)

type DictRepository struct {
	db *gorm.DB
}

func NewDictRepository(db *gorm.DB) *DictRepository {
	return &DictRepository{db: db}
}

// DictType 相关

func (r *DictRepository) CreateDictType(dictType *domain.DictType) error {
	return r.db.Create(dictType).Error
}

func (r *DictRepository) GetDictTypeByID(id uint) (*domain.DictType, error) {
	var dictType domain.DictType
	err := r.db.First(&dictType, id).Error
	return &dictType, err
}

func (r *DictRepository) GetDictTypeByType(dictType string) (*domain.DictType, error) {
	var dt domain.DictType
	err := r.db.Where("dict_type = ?", dictType).First(&dt).Error
	return &dt, err
}

func (r *DictRepository) UpdateDictType(dictType *domain.DictType) error {
	return r.db.Save(dictType).Error
}

func (r *DictRepository) DeleteDictType(id uint) error {
	return r.db.Delete(&domain.DictType{}, id).Error
}

func (r *DictRepository) ListDictTypes(req *domain.ListDictTypeRequest) ([]*domain.DictType, int64, error) {
	var dictTypes []*domain.DictType
	var total int64

	query := r.db.Model(&domain.DictType{})

	if req.DictName != "" {
		query = query.Where("dict_name LIKE ?", "%"+req.DictName+"%")
	}
	if req.DictType != "" {
		query = query.Where("dict_type LIKE ?", "%"+req.DictType+"%")
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.PageSize
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&dictTypes).Error

	return dictTypes, total, err
}

// DictData 相关

func (r *DictRepository) CreateDictData(dictData *domain.DictData) error {
	return r.db.Create(dictData).Error
}

func (r *DictRepository) GetDictDataByID(id uint) (*domain.DictData, error) {
	var dictData domain.DictData
	err := r.db.First(&dictData, id).Error
	return &dictData, err
}

func (r *DictRepository) UpdateDictData(dictData *domain.DictData) error {
	return r.db.Save(dictData).Error
}

func (r *DictRepository) DeleteDictData(id uint) error {
	return r.db.Delete(&domain.DictData{}, id).Error
}

func (r *DictRepository) ListDictData(req *domain.ListDictDataRequest) ([]*domain.DictData, int64, error) {
	var dictDataList []*domain.DictData
	var total int64

	query := r.db.Model(&domain.DictData{})

	if req.DictType != "" {
		query = query.Where("dict_type = ?", req.DictType)
	}
	if req.DictLabel != "" {
		query = query.Where("dict_label LIKE ?", "%"+req.DictLabel+"%")
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.PageSize
	err := query.Order("dict_sort ASC, created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&dictDataList).Error

	return dictDataList, total, err
}

func (r *DictRepository) GetDictDataByType(dictType string) ([]*domain.DictData, error) {
	var dictDataList []*domain.DictData
	err := r.db.Where("dict_type = ? AND status = 1", dictType).
		Order("dict_sort ASC").
		Find(&dictDataList).Error
	return dictDataList, err
}

func (r *DictRepository) CheckDictTypeExists(dictType string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.DictType{}).Where("dict_type = ?", dictType).Count(&count).Error
	return count > 0, err
}

func (r *DictRepository) CountDictDataByType(dictType string) (int64, error) {
	var count int64
	err := r.db.Model(&domain.DictData{}).Where("dict_type = ?", dictType).Count(&count).Error
	return count, err
}
