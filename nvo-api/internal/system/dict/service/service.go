package service

import (
	"errors"
	"nvo-api/core/log"
	"nvo-api/internal/system/dict/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"

	dictDomain "nvo-api/internal/system/dict/domain"
)

type DictService struct {
	repo *repository.DictRepository
}

func NewDictService(db *gorm.DB) *DictService {
	return &DictService{
		repo: repository.NewDictRepository(db),
	}
}

// CreateDictType 创建字典类型
func (s *DictService) CreateDictType(req *dictDomain.CreateDictTypeRequest) (*dictDomain.DictType, error) {
	// 检查字典类型是否已存在
	exists, err := s.repo.CheckDictTypeExists(req.DictType)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("字典类型已存在")
	}

	dictType := &dictDomain.DictType{
		DictName: req.DictName,
		DictType: req.DictType,
		Status:   req.Status,
		Remark:   req.Remark,
	}

	if err := s.repo.CreateDictType(dictType); err != nil {
		log.Error("create dict type failed", zap.Error(err))
		return nil, err
	}

	log.Info("dict type created", zap.String("dict_type", dictType.DictType))
	return dictType, nil
}

// GetDictTypeByID 根据ID获取字典类型
func (s *DictService) GetDictTypeByID(id uint) (*dictDomain.DictTypeResponse, error) {
	dictType, err := s.repo.GetDictTypeByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("字典类型不存在")
		}
		return nil, err
	}

	return dictType.ToResponse(), nil
}

// GetDictTypeByType 根据类型获取字典类型
func (s *DictService) GetDictTypeByType(dictType string) (*dictDomain.DictTypeResponse, error) {
	dt, err := s.repo.GetDictTypeByType(dictType)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("字典类型不存在")
		}
		return nil, err
	}

	return dt.ToResponse(), nil
}

// UpdateDictType 更新字典类型
func (s *DictService) UpdateDictType(id uint, req *dictDomain.UpdateDictTypeRequest) error {
	dictType, err := s.repo.GetDictTypeByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("字典类型不存在")
		}
		return err
	}

	if req.DictName != "" {
		dictType.DictName = req.DictName
	}
	dictType.Status = req.Status
	if req.Remark != "" {
		dictType.Remark = req.Remark
	}

	if err := s.repo.UpdateDictType(dictType); err != nil {
		log.Error("update dict type failed", zap.Error(err))
		return err
	}

	log.Info("dict type updated", zap.Uint("id", id))
	return nil
}

// DeleteDictType 删除字典类型
func (s *DictService) DeleteDictType(id uint) error {
	// 检查是否存在
	dictType, err := s.repo.GetDictTypeByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("字典类型不存在")
		}
		return err
	}

	// 检查是否有关联的字典数据
	count, err := s.repo.CountDictDataByType(dictType.DictType)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该字典类型下存在字典数据，无法删除")
	}

	if err := s.repo.DeleteDictType(id); err != nil {
		log.Error("delete dict type failed", zap.Error(err))
		return err
	}

	log.Info("dict type deleted", zap.Uint("id", id))
	return nil
}

// ListDictTypes 获取字典类型列表
func (s *DictService) ListDictTypes(req *dictDomain.ListDictTypeRequest) ([]*dictDomain.DictTypeResponse, int64, error) {
	dictTypes, total, err := s.repo.ListDictTypes(req)
	if err != nil {
		return nil, 0, err
	}

	return dictDomain.ToDictTypeResponseList(dictTypes), total, nil
}

// CreateDictData 创建字典数据
func (s *DictService) CreateDictData(req *dictDomain.CreateDictDataRequest) (*dictDomain.DictData, error) {
	// 检查字典类型是否存在
	exists, err := s.repo.CheckDictTypeExists(req.DictType)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("字典类型不存在")
	}

	dictData := &dictDomain.DictData{
		DictSort:  req.DictSort,
		DictLabel: req.DictLabel,
		DictValue: req.DictValue,
		DictType:  req.DictType,
		CSSClass:  req.CSSClass,
		ListClass: req.ListClass,
		IsDefault: req.IsDefault,
		Status:    req.Status,
		Remark:    req.Remark,
	}

	if err := s.repo.CreateDictData(dictData); err != nil {
		log.Error("create dict data failed", zap.Error(err))
		return nil, err
	}

	log.Info("dict data created", zap.String("dict_type", dictData.DictType), zap.String("dict_label", dictData.DictLabel))
	return dictData, nil
}

// GetDictDataByID 根据ID获取字典数据
func (s *DictService) GetDictDataByID(id uint) (*dictDomain.DictDataResponse, error) {
	dictData, err := s.repo.GetDictDataByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("字典数据不存在")
		}
		return nil, err
	}

	return dictData.ToResponse(), nil
}

// UpdateDictData 更新字典数据
func (s *DictService) UpdateDictData(id uint, req *dictDomain.UpdateDictDataRequest) error {
	dictData, err := s.repo.GetDictDataByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("字典数据不存在")
		}
		return err
	}

	dictData.DictSort = req.DictSort
	if req.DictLabel != "" {
		dictData.DictLabel = req.DictLabel
	}
	if req.DictValue != "" {
		dictData.DictValue = req.DictValue
	}
	if req.CSSClass != "" {
		dictData.CSSClass = req.CSSClass
	}
	if req.ListClass != "" {
		dictData.ListClass = req.ListClass
	}
	dictData.IsDefault = req.IsDefault
	dictData.Status = req.Status
	if req.Remark != "" {
		dictData.Remark = req.Remark
	}

	if err := s.repo.UpdateDictData(dictData); err != nil {
		log.Error("update dict data failed", zap.Error(err))
		return err
	}

	log.Info("dict data updated", zap.Uint("id", id))
	return nil
}

// DeleteDictData 删除字典数据
func (s *DictService) DeleteDictData(id uint) error {
	_, err := s.repo.GetDictDataByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("字典数据不存在")
		}
		return err
	}

	if err := s.repo.DeleteDictData(id); err != nil {
		log.Error("delete dict data failed", zap.Error(err))
		return err
	}

	log.Info("dict data deleted", zap.Uint("id", id))
	return nil
}

// ListDictData 获取字典数据列表
func (s *DictService) ListDictData(req *dictDomain.ListDictDataRequest) ([]*dictDomain.DictDataResponse, int64, error) {
	dictDataList, total, err := s.repo.ListDictData(req)
	if err != nil {
		return nil, 0, err
	}

	return dictDomain.ToDictDataResponseList(dictDataList), total, nil
}

// GetDictDataByType 根据字典类型获取字典数据
func (s *DictService) GetDictDataByType(dictType string) ([]*dictDomain.DictDataResponse, error) {
	dictDataList, err := s.repo.GetDictDataByType(dictType)
	if err != nil {
		return nil, err
	}

	return dictDomain.ToDictDataResponseList(dictDataList), nil
}
