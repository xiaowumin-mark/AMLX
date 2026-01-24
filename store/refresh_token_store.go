package store

import (
	"context"

	"github.com/xiaowumin-mark/AMLX/model"
	"gorm.io/gorm"
)

type RefreshTokenStore interface {
	Create(ctx context.Context, token *model.RefreshTokens) error             // 创建
	GetValid(ctx context.Context, token string) (*model.RefreshTokens, error) // 获取
	Revoke(ctx context.Context, token string) error                           // 撤销
	RevokeByUser(ctx context.Context, userID uint) error                      // 撤销
}

type refreshTokenStore struct {
	db *gorm.DB
}

func NewRefreshTokenStore(db *gorm.DB) RefreshTokenStore {
	return &refreshTokenStore{db: db}
}

func (s *refreshTokenStore) Create(ctx context.Context, token *model.RefreshTokens) error {
	return s.db.WithContext(ctx).Create(token).Error
}
func (s *refreshTokenStore) GetValid(ctx context.Context, token string) (*model.RefreshTokens, error) {
	var refreshToken model.RefreshTokens
	return &refreshToken, s.db.WithContext(ctx).Where("token = ? AND revoked = ?", token, false).First(&refreshToken).Error
}
func (s *refreshTokenStore) Revoke(ctx context.Context, token string) error {
	return s.db.WithContext(ctx).Model(&model.RefreshTokens{}).Where("token = ?", token).Update("revoked", true).Error
}

func (s *refreshTokenStore) RevokeByUser(ctx context.Context, userID uint) error {
	return s.db.WithContext(ctx).Model(&model.RefreshTokens{}).Where("user_id = ?", userID).Update("revoked", true).Error
}
