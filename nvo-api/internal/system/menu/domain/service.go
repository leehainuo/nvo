package domain

type MenuService interface {
	Create(req *CreateMenuRequest) (*Menu, error)
	GetByID(id uint) (*Menu, error)
	Update(id uint, req *UpdateMenuRequest) error
	Delete(id uint) error
	GetTree() ([]*MenuTree, error)
	GetList() ([]*Menu, error)
}
