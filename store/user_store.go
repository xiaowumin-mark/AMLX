package store

import (
	"context"

	"github.com/xiaowumin-mark/AMLX/model"
	"gorm.io/gorm"
)

type UserStore interface {
	GetByID(ctx context.Context, id uint) (*model.Users, error)
	GetByEmail(ctx context.Context, email string) (*model.Users, error)
	Create(ctx context.Context, user *model.Users) error
	Update(ctx context.Context, user *model.Users) error
	SetBan(ctx context.Context, id uint, ban bool) error
	Count(ctx context.Context) (int64, error)
}

type userStore struct {
	db *gorm.DB
}

func NewUserStore(db *gorm.DB) UserStore {
	return &userStore{db: db}
}

func (s *userStore) GetByID(ctx context.Context, id uint) (*model.Users, error) {
	var user model.Users
	return &user, s.db.WithContext(ctx).First(&user, id).Error
}
func (s *userStore) GetByEmail(ctx context.Context, email string) (*model.Users, error) {
	var user model.Users
	return &user, s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
}
func (s *userStore) Create(ctx context.Context, user *model.Users) error {
	return s.db.WithContext(ctx).Create(user).Error
}
func (s *userStore) Update(ctx context.Context, user *model.Users) error {
	return s.db.WithContext(ctx).Updates(user).Error
}
func (s *userStore) SetBan(ctx context.Context, id uint, ban bool) error {
	return s.db.WithContext(ctx).Model(&model.Users{}).Where("id = ?", id).Update("ban", ban).Error
}

func (s *userStore) Count(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&model.Users{}).Count(&count).Error
	return count, err
}
