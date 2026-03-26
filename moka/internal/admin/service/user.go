package service

import (
	"github.com/casbin/casbin/v3"
	"gorm.io/gorm"
)

type UserService struct {
	db       *gorm.DB
	enforcer *casbin.SyncedCachedEnforcer
}

func NewUserService(db *gorm.DB, enforcer *casbin.SyncedCachedEnforcer) *UserService {
	return &UserService{
		db:       db,
		enforcer: enforcer,
	}
}

