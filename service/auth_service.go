package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/xiaowumin-mark/AMLX/config"
	"github.com/xiaowumin-mark/AMLX/model"
	"github.com/xiaowumin-mark/AMLX/store"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserBanned          = errors.New("user is banned")
	ErrRegistrationClosed  = errors.New("registration disabled")
	ErrRefreshTokenInvalid = errors.New("refresh token invalid")
)

type TokenPair struct {
	AccessToken     string    `json:"access_token"`
	RefreshToken    string    `json:"refresh_token"`
	AccessExpiresAt time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type RegisterRequest struct {
	Name     string
	Email    string
	Password string
	RoleID   uint
}

type LoginRequest struct {
	Email    string
	Password string
}

type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) (*model.Users, *TokenPair, error)
	Login(ctx context.Context, req LoginRequest) (*model.Users, *TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAll(ctx context.Context, userID uint) error
	ParseAccessToken(token string) (*AccessClaims, error)
	ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error
}

type authService struct {
	users         store.UserStore
	refreshTokens store.RefreshTokenStore
	tokens        *JWTManager
	cfg           config.AuthConfig
}

func NewAuthService(cfg config.AuthConfig, users store.UserStore, refreshTokens store.RefreshTokenStore, tokens *JWTManager) AuthService {
	return &authService{
		users:         users,
		refreshTokens: refreshTokens,
		tokens:        tokens,
		cfg:           cfg,
	}
}

func (s *authService) Register(ctx context.Context, req RegisterRequest) (*model.Users, *TokenPair, error) {
	if !s.cfg.AllowRegisterValue() {
		return nil, nil, ErrRegistrationClosed
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return nil, nil, ErrInvalidInput
	}
	if !s.cfg.AllowRegisterRole {
		req.RoleID = s.cfg.DefaultRoleID
	}
	if req.RoleID == 0 {
		req.RoleID = s.cfg.DefaultRoleID
	}

	if _, err := s.users.GetByEmail(ctx, req.Email); err == nil {
		return nil, nil, ErrEmailExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, err
	}

	hashedPassword, err := HashPassword(req.Password, s.cfg.BcryptCost)
	if err != nil {
		return nil, nil, err
	}
	user := &model.Users{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		RoleId:   req.RoleID,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return nil, nil, err
	}
	return user, pair, nil
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (*model.Users, *TokenPair, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	password := strings.TrimSpace(req.Password)
	if email == "" || password == "" {
		return nil, nil, ErrInvalidInput
	}

	user, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, nil, err
	}
	if user.Ban {
		return nil, nil, ErrUserBanned
	}
	if err := ComparePassword(user.Password, password); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return nil, nil, err
	}
	return user, pair, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, ErrRefreshTokenInvalid
	}

	claims, err := s.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}
	userID, err := parseSubject(claims.Subject)
	if err != nil {
		return nil, ErrRefreshTokenInvalid
	}

	tokenHash := hashToken(refreshToken)
	record, err := s.refreshTokens.GetValid(ctx, tokenHash)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRefreshTokenInvalid
	}
	if err != nil {
		return nil, err
	}
	if record.UserId != userID {
		return nil, ErrRefreshTokenInvalid
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.Ban {
		return nil, ErrUserBanned
	}

	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}

	if !s.cfg.RefreshTokenReuse {
		_ = s.refreshTokens.Revoke(ctx, tokenHash)
	}
	return pair, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return ErrRefreshTokenInvalid
	}
	return s.refreshTokens.Revoke(ctx, hashToken(refreshToken))
}

func (s *authService) LogoutAll(ctx context.Context, userID uint) error {
	if userID == 0 {
		return ErrInvalidInput
	}
	return s.refreshTokens.RevokeByUser(ctx, userID)
}

func (s *authService) ParseAccessToken(token string) (*AccessClaims, error) {
	return s.tokens.ParseAccessToken(token)
}

func (s *authService) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	if userID == 0 {
		return ErrInvalidInput
	}
	oldPassword = strings.TrimSpace(oldPassword)
	newPassword = strings.TrimSpace(newPassword)
	if oldPassword == "" || newPassword == "" {
		return ErrInvalidInput
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if err := ComparePassword(user.Password, oldPassword); err != nil {
		return ErrInvalidCredentials
	}

	hashedPassword, err := HashPassword(newPassword, s.cfg.BcryptCost)
	if err != nil {
		return err
	}
	user.Password = hashedPassword
	return s.users.Update(ctx, user)
}

func (s *authService) issueTokenPair(ctx context.Context, user *model.Users) (*TokenPair, error) {
	accessToken, accessExpires, err := s.tokens.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}
	refreshToken, refreshExpires, err := s.tokens.GenerateRefreshToken(user)
	if err != nil {
		return nil, err
	}
	record := &model.RefreshTokens{
		UserId:    user.ID,
		Token:     hashToken(refreshToken),
		ExpiredAt: refreshExpires.UnixMilli(),
		Revoked:   false,
	}
	if err := s.refreshTokens.Create(ctx, record); err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:     accessToken,
		RefreshToken:    refreshToken,
		AccessExpiresAt: accessExpires,
		RefreshExpiresAt: refreshExpires,
	}, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func parseSubject(subject string) (uint, error) {
	parsed, err := strconv.ParseUint(subject, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}
