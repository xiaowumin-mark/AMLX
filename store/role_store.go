package store

import (
	"context"

	"github.com/xiaowumin-mark/AMLX/model"
	"gorm.io/gorm"
)

type RoleStore interface {
	GetByID(ctx context.Context, id uint) (*model.Roles, error)
	GetByName(ctx context.Context, name string) (*model.Roles, error)
	Create(ctx context.Context, role *model.Roles) error
}

type roleStore struct {
	db *gorm.DB
}

func NewRoleStore(db *gorm.DB) RoleStore {
	return &roleStore{db: db}
}

func (s *roleStore) GetByID(ctx context.Context, id uint) (*model.Roles, error) {
	var role model.Roles
	return &role, s.db.WithContext(ctx).First(&role, id).Error
}
func (s *roleStore) GetByName(ctx context.Context, name string) (*model.Roles, error) {
	var role model.Roles
	return &role, s.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
}
func (s *roleStore) Create(ctx context.Context, role *model.Roles) error {
	return s.db.WithContext(ctx).Create(role).Error
}
