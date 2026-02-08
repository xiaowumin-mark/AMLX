package store

import (
	"context"

	"github.com/xiaowumin-mark/AMLX/model"
	"gorm.io/gorm"
)

type PermissionStore interface {
	GetByID(ctx context.Context, id uint) (*model.Permissions, error)
	GetByName(ctx context.Context, name string) (*model.Permissions, error)
	ListByIDs(ctx context.Context, ids []uint) ([]model.Permissions, error)
	Create(ctx context.Context, permission *model.Permissions) error
}

type permissionStore struct {
	db *gorm.DB
}

func NewPermissionStore(db *gorm.DB) PermissionStore {
	return &permissionStore{db: db}
}

func (s *permissionStore) GetByID(ctx context.Context, id uint) (*model.Permissions, error) {
	var permission model.Permissions
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&permission).Error
	return &permission, err
}
func (s *permissionStore) GetByName(ctx context.Context, name string) (*model.Permissions, error) {
	var permission model.Permissions
	err := s.db.WithContext(ctx).Where("name = ?", name).First(&permission).Error
	return &permission, err
}
func (s *permissionStore) ListByIDs(ctx context.Context, ids []uint) ([]model.Permissions, error) {
	var permissions []model.Permissions
	err := s.db.WithContext(ctx).Where("id IN ?", ids).Find(&permissions).Error
	return permissions, err
}

func (s *permissionStore) Create(ctx context.Context, permission *model.Permissions) error {
	return s.db.WithContext(ctx).Create(permission).Error
}
