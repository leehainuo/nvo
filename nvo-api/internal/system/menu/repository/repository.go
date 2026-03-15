package repository

import (
	"nvo-api/internal/system/menu/domain"
	"gorm.io/gorm"
)

type MenuRepository struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

func (r *MenuRepository) Create(menu *domain.Menu) error {
	return r.db.Create(menu).Error
}

func (r *MenuRepository) GetByID(id uint) (*domain.Menu, error) {
	var menu domain.Menu
	err := r.db.First(&menu, id).Error
	return &menu, err
}

func (r *MenuRepository) Update(menu *domain.Menu) error {
	return r.db.Save(menu).Error
}

func (r *MenuRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Menu{}, id).Error
}

func (r *MenuRepository) GetAll() ([]*domain.Menu, error) {
	var menus []*domain.Menu
	err := r.db.Order("sort ASC, id ASC").Find(&menus).Error
	return menus, err
}

func (r *MenuRepository) GetByParentID(parentID uint) ([]*domain.Menu, error) {
	var menus []*domain.Menu
	err := r.db.Where("parent_id = ?", parentID).Order("sort ASC").Find(&menus).Error
	return menus, err
}
