package service

import (
	"context"
	"errors"
	"strings"

	"github.com/xiaowumin-mark/AMLX/model"
	"github.com/xiaowumin-mark/AMLX/store"
	"gorm.io/gorm"
)

var (
	ErrRoleNotFound       = errors.New("role not found")
	ErrRoleExists         = errors.New("role already exists")
	ErrPermissionNotFound = errors.New("permission not found")
	ErrPermissionExists   = errors.New("permission already exists")
)

type PermissionService interface {
	CreateRole(ctx context.Context, name, description string) (*model.Roles, error)
	CreatePermission(ctx context.Context, name, description string) (*model.Permissions, error)
	AddPermissionToRole(ctx context.Context, roleID, permID uint) error
	RemovePermissionFromRole(ctx context.Context, roleID, permID uint) error
	ListPermissionsByRole(ctx context.Context, roleID uint) ([]model.Permissions, error)
	HasPermission(ctx context.Context, roleID uint, permName string) (bool, error)
	EnsureAdminRole(ctx context.Context) (uint, error)
}

type permissionService struct {
	roles           store.RoleStore
	permissions     store.PermissionStore
	rolePermissions store.RolePermissionStore
}

func NewPermissionService(roles store.RoleStore, permissions store.PermissionStore, rolePermissions store.RolePermissionStore) PermissionService {
	return &permissionService{
		roles:           roles,
		permissions:     permissions,
		rolePermissions: rolePermissions,
	}
}

func (s *permissionService) CreateRole(ctx context.Context, name, description string) (*model.Roles, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	if name == "" {
		return nil, ErrInvalidInput
	}
	if _, err := s.roles.GetByName(ctx, name); err == nil {
		return nil, ErrRoleExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	role := &model.Roles{
		Name:        name,
		Description: description,
	}
	if err := s.roles.Create(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *permissionService) CreatePermission(ctx context.Context, name, description string) (*model.Permissions, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	if name == "" {
		return nil, ErrInvalidInput
	}
	if _, err := s.permissions.GetByName(ctx, name); err == nil {
		return nil, ErrPermissionExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	permission := &model.Permissions{
		Name:        name,
		Description: description,
	}
	if err := s.permissions.Create(ctx, permission); err != nil {
		return nil, err
	}
	return permission, nil
}

func (s *permissionService) AddPermissionToRole(ctx context.Context, roleID, permID uint) error {
	if roleID == 0 || permID == 0 {
		return ErrInvalidInput
	}
	if _, err := s.roles.GetByID(ctx, roleID); errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrRoleNotFound
	} else if err != nil {
		return err
	}
	if _, err := s.permissions.GetByID(ctx, permID); errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrPermissionNotFound
	} else if err != nil {
		return err
	}
	return s.rolePermissions.AddPermission(ctx, roleID, permID)
}

func (s *permissionService) RemovePermissionFromRole(ctx context.Context, roleID, permID uint) error {
	if roleID == 0 || permID == 0 {
		return ErrInvalidInput
	}
	return s.rolePermissions.RemovePermission(ctx, roleID, permID)
}

func (s *permissionService) ListPermissionsByRole(ctx context.Context, roleID uint) ([]model.Permissions, error) {
	if roleID == 0 {
		return nil, ErrInvalidInput
	}
	ids, err := s.rolePermissions.ListPermissionIDsByRole(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []model.Permissions{}, nil
	}
	return s.permissions.ListByIDs(ctx, ids)
}

func (s *permissionService) HasPermission(ctx context.Context, roleID uint, permName string) (bool, error) {
	permName = strings.TrimSpace(permName)
	if roleID == 0 || permName == "" {
		return false, ErrInvalidInput
	}
	return s.rolePermissions.HasPermission(ctx, roleID, permName)
}

func (s *permissionService) EnsureAdminRole(ctx context.Context) (uint, error) {
	role, err := s.roles.GetByName(ctx, "admin")
	if errors.Is(err, gorm.ErrRecordNotFound) {
		role, err = s.CreateRole(ctx, "admin", "System administrator")
		if err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}

	perm, err := s.permissions.GetByName(ctx, "admin")
	if errors.Is(err, gorm.ErrRecordNotFound) {
		perm, err = s.CreatePermission(ctx, "admin", "Super admin permission")
		if err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}

	_ = s.rolePermissions.AddPermission(ctx, role.ID, perm.ID)
	return role.ID, nil
}
