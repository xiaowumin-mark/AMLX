package store

import (
	"context"

	"github.com/xiaowumin-mark/AMLX/model"
	"gorm.io/gorm"
)

type RolePermissionStore interface {
	AddPermission(ctx context.Context, roleID, permID uint) error
	RemovePermission(ctx context.Context, roleID, permID uint) error
	ListPermissionIDsByRole(ctx context.Context, roleID uint) ([]uint, error)
	HasPermission(ctx context.Context, roleID uint, permName string) (bool, error)
}

type rolePermissionStore struct {
	db *gorm.DB
}

func NewRolePermissionStore(db *gorm.DB) RolePermissionStore {
	return &rolePermissionStore{db: db}
}

func (s *rolePermissionStore) AddPermission(ctx context.Context, roleID, permID uint) error {
	return s.db.WithContext(ctx).Create(&model.RolePermissions{RoleId: roleID, PermissionId: permID}).Error
}
func (s *rolePermissionStore) RemovePermission(ctx context.Context, roleID, permID uint) error {
	return s.db.WithContext(ctx).Where("role_id = ? AND permission_id = ?", roleID, permID).Delete(&model.RolePermissions{}).Error
}
func (s *rolePermissionStore) ListPermissionIDsByRole(ctx context.Context, roleID uint) ([]uint, error) {
	var permissionIDs []uint
	return permissionIDs, s.db.WithContext(ctx).Model(&model.RolePermissions{}).Where("role_id = ?", roleID).Pluck("permission_id", &permissionIDs).Error

}
func (s *rolePermissionStore) HasPermission(ctx context.Context, roleID uint, permName string) (bool, error) {
	var permission model.Permissions
	return s.db.WithContext(ctx).Where("name = ?", permName).First(&permission).Error == nil, s.db.WithContext(ctx).Model(&model.RolePermissions{}).Where("role_id = ? AND permission_id = ?", roleID, permission.ID).Error
}
