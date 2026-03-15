package service

import (
	"errors"
	"nvo-api/internal/system/menu/domain"
	"nvo-api/internal/system/menu/repository"
	"gorm.io/gorm"
)

type MenuService struct {
	db   *gorm.DB
	repo *repository.MenuRepository
}

func NewMenuService(db *gorm.DB) domain.MenuService {
	return &MenuService{
		db:   db,
		repo: repository.NewMenuRepository(db),
	}
}

func (s *MenuService) Create(req *domain.CreateMenuRequest) (*domain.Menu, error) {
	menu := &domain.Menu{
		ParentID:  req.ParentID,
		Name:      req.Name,
		Path:      req.Path,
		Component: req.Component,
		Icon:      req.Icon,
		Sort:      req.Sort,
		Type:      req.Type,
		Visible:   req.Visible,
		Status:    req.Status,
		Perms:     req.Perms,
	}

	if err := s.repo.Create(menu); err != nil {
		return nil, err
	}
	return menu, nil
}

func (s *MenuService) GetByID(id uint) (*domain.Menu, error) {
	menu, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("菜单不存在")
		}
		return nil, err
	}
	return menu, nil
}

func (s *MenuService) Update(id uint, req *domain.UpdateMenuRequest) error {
	menu, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("菜单不存在")
		}
		return err
	}

	if req.ParentID != nil {
		menu.ParentID = *req.ParentID
	}
	if req.Name != "" {
		menu.Name = req.Name
	}
	if req.Path != "" {
		menu.Path = req.Path
	}
	if req.Component != "" {
		menu.Component = req.Component
	}
	if req.Icon != "" {
		menu.Icon = req.Icon
	}
	if req.Sort != nil {
		menu.Sort = *req.Sort
	}
	if req.Type != nil {
		menu.Type = *req.Type
	}
	if req.Visible != nil {
		menu.Visible = *req.Visible
	}
	if req.Status != nil {
		menu.Status = *req.Status
	}
	if req.Perms != "" {
		menu.Perms = req.Perms
	}

	return s.repo.Update(menu)
}

func (s *MenuService) Delete(id uint) error {
	if _, err := s.repo.GetByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("菜单不存在")
		}
		return err
	}

	children, _ := s.repo.GetByParentID(id)
	if len(children) > 0 {
		return errors.New("存在子菜单，无法删除")
	}

	return s.repo.Delete(id)
}

func (s *MenuService) GetTree() ([]*domain.MenuTree, error) {
	menus, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	return s.buildTree(menus, 0), nil
}

func (s *MenuService) GetList() ([]*domain.Menu, error) {
	return s.repo.GetAll()
}

func (s *MenuService) buildTree(menus []*domain.Menu, parentID uint) []*domain.MenuTree {
	var tree []*domain.MenuTree
	for _, menu := range menus {
		if menu.ParentID == parentID {
			node := &domain.MenuTree{
				Menu:     menu,
				Children: s.buildTree(menus, menu.ID),
			}
			tree = append(tree, node)
		}
	}
	return tree
}
