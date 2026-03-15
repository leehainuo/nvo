package repository

import (
	"nvo-api/internal/system/dept/domain"
	"gorm.io/gorm"
)

type DeptRepository struct {
	db *gorm.DB
}

func NewDeptRepository(db *gorm.DB) *DeptRepository {
	return &DeptRepository{db: db}
}

func (r *DeptRepository) Create(dept *domain.Dept) error {
	return r.db.Create(dept).Error
}

func (r *DeptRepository) GetByID(id uint) (*domain.Dept, error) {
	var dept domain.Dept
	err := r.db.First(&dept, id).Error
	return &dept, err
}

func (r *DeptRepository) Update(dept *domain.Dept) error {
	return r.db.Save(dept).Error
}

func (r *DeptRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Dept{}, id).Error
}

func (r *DeptRepository) GetAll() ([]*domain.Dept, error) {
	var depts []*domain.Dept
	err := r.db.Order("sort ASC, id ASC").Find(&depts).Error
	return depts, err
}

func (r *DeptRepository) GetByParentID(parentID uint) ([]*domain.Dept, error) {
	var depts []*domain.Dept
	err := r.db.Where("parent_id = ?", parentID).Order("sort ASC").Find(&depts).Error
	return depts, err
}
