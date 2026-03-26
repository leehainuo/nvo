package datascope

import (
	"fmt"

	"gorm.io/gorm"
)

type DataScope string

const (
	All       DataScope = "all"
	Custom    DataScope = "custom"
	Dept      DataScope = "dept"
	DeptChild DataScope = "dept_child"
	Self      DataScope = "self"
)

// Filter 数据权限过滤器
type Filter struct {
	UserID     int64
	DeptID     int64
	RoleScopes []RoleScope
	db         *gorm.DB
}

// RoleScope 角色数据权限信息
type RoleScope struct {
    RoleID      int64
    DataScope   DataScope
    CustomDepts []int64  // 自定义部门列表
}

func NewFilter(db *gorm.DB, userID int64) (*Filter, error) {
	filter := &Filter{
		UserID:     userID,
		db:         db,
		RoleScopes: make([]RoleScope, 0),
	}

	var roles []struct {
		ID        int64
		DataScope DataScope
	}
	
    if err := db.Table("roles").Select("roles.id, roles.data_scope").Joins("INNER JOIN user_roles ON user_roles.role_id = roles.id").Where("user_roles.user_id = ? AND roles.status = 1", userID).Find(&roles).Error; err != nil {
        return nil, fmt.Errorf("failed to get user roles: %w", err)
    }

	for _, role := range roles {
		scope := RoleScope{
			RoleID:    role.ID,
			DataScope: DataScope(role.DataScope),
		}

		if scope.DataScope == Custom {
			var customDepts []struct {
				DeptID int64
			}
			 if err := db.Table("role_data_scopes").Select("dept_id").Where("role_id = ?", role.ID).Find(&customDepts).Error; err == nil {
                for _, d := range customDepts {
                    scope.CustomDepts = append(scope.CustomDepts, d.DeptID)
                }
			}
		}

		filter.RoleScopes = append(filter.RoleScopes, scope)
	}

	return filter, nil
}
