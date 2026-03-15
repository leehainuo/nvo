package repository

import (
	"nvo-api/internal/system/role/domain"

	"gorm.io/gorm"
)

// RoleRepository 角色数据访问层
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository 创建角色仓库
func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Create 创建角色
func (r *RoleRepository) Create(role *domain.Role) error {
	return r.db.Create(role).Error
}

// GetByID 根据 ID 获取角色
func (r *RoleRepository) GetByID(id uint) (*domain.Role, error) {
	var role domain.Role
	err := r.db.First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetByCode 根据编码获取角色
func (r *RoleRepository) GetByCode(code string) (*domain.Role, error) {
	var role domain.Role
	err := r.db.Where("code = ?", code).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// Update 更新角色
func (r *RoleRepository) Update(role *domain.Role) error {
	return r.db.Save(role).Error
}

// Delete 删除角色（软删除）
func (r *RoleRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Role{}, id).Error
}

// List 获取角色列表
func (r *RoleRepository) List(req *domain.ListRoleRequest) ([]*domain.Role, int64, error) {
	var roles []*domain.Role
	var total int64

	query := r.db.Model(&domain.Role{})

	// 条件过滤
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Code != "" {
		query = query.Where("code LIKE ?", "%"+req.Code+"%")
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("sort ASC, id DESC").Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// ExistsByCode 检查角色编码是否存在
func (r *RoleRepository) ExistsByCode(code string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Role{}).Where("code = ?", code).Count(&count).Error
	return count > 0, err
}

// GetAll 获取所有角色
func (r *RoleRepository) GetAll() ([]*domain.Role, error) {
	var roles []*domain.Role
	err := r.db.Where("status = ?", 1).Order("sort ASC").Find(&roles).Error
	return roles, err
}
