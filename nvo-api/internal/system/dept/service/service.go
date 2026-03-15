package service

import (
	"errors"
	"nvo-api/internal/system/dept/domain"
	"nvo-api/internal/system/dept/repository"
	"gorm.io/gorm"
)

type DeptService struct {
	db   *gorm.DB
	repo *repository.DeptRepository
}

func NewDeptService(db *gorm.DB) domain.DeptService {
	return &DeptService{
		db:   db,
		repo: repository.NewDeptRepository(db),
	}
}

func (s *DeptService) Create(req *domain.CreateDeptRequest) (*domain.Dept, error) {
	dept := &domain.Dept{
		ParentID: req.ParentID,
		Name:     req.Name,
		Sort:     req.Sort,
		Leader:   req.Leader,
		Phone:    req.Phone,
		Email:    req.Email,
		Status:   req.Status,
	}

	if err := s.repo.Create(dept); err != nil {
		return nil, err
	}
	return dept, nil
}

func (s *DeptService) GetByID(id uint) (*domain.Dept, error) {
	dept, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("部门不存在")
		}
		return nil, err
	}
	return dept, nil
}

func (s *DeptService) Update(id uint, req *domain.UpdateDeptRequest) error {
	dept, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("部门不存在")
		}
		return err
	}

	if req.ParentID != nil {
		dept.ParentID = *req.ParentID
	}
	if req.Name != "" {
		dept.Name = req.Name
	}
	if req.Sort != nil {
		dept.Sort = *req.Sort
	}
	if req.Leader != "" {
		dept.Leader = req.Leader
	}
	if req.Phone != "" {
		dept.Phone = req.Phone
	}
	if req.Email != "" {
		dept.Email = req.Email
	}
	if req.Status != nil {
		dept.Status = *req.Status
	}

	return s.repo.Update(dept)
}

func (s *DeptService) Delete(id uint) error {
	if _, err := s.repo.GetByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("部门不存在")
		}
		return err
	}

	children, _ := s.repo.GetByParentID(id)
	if len(children) > 0 {
		return errors.New("存在子部门，无法删除")
	}

	return s.repo.Delete(id)
}

func (s *DeptService) GetTree() ([]*domain.DeptTree, error) {
	depts, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	return s.buildTree(depts, 0), nil
}

func (s *DeptService) GetList() ([]*domain.Dept, error) {
	return s.repo.GetAll()
}

func (s *DeptService) buildTree(depts []*domain.Dept, parentID uint) []*domain.DeptTree {
	var tree []*domain.DeptTree
	for _, dept := range depts {
		if dept.ParentID == parentID {
			node := &domain.DeptTree{
				Dept:     dept,
				Children: s.buildTree(depts, dept.ID),
			}
			tree = append(tree, node)
		}
	}
	return tree
}
