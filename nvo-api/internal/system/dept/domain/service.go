package domain

type DeptService interface {
	Create(req *CreateDeptRequest) (*Dept, error)
	GetByID(id uint) (*Dept, error)
	Update(id uint, req *UpdateDeptRequest) error
	Delete(id uint) error
	GetTree() ([]*DeptTree, error)
	GetList() ([]*Dept, error)
}
