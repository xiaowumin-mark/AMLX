package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/xiaowumin-mark/AMLX/config"
	"github.com/xiaowumin-mark/AMLX/model"
)

var (
	ErrTokenInvalid = errors.New("token invalid")
	ErrTokenExpired = errors.New("token expired")
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type AccessClaims struct {
	jwt.RegisteredClaims
	RoleID uint     `json:"role_id"`
	Email  string   `json:"email"`
	Type   TokenType `json:"typ"`
}

type JWTManager struct {
	secret     []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJWTManager(cfg config.AuthConfig) (*JWTManager, error) {
	if cfg.JWTSecret == "" {
		return nil, errors.New("auth.jwt_secret is required")
	}
	return &JWTManager{
		secret:     []byte(cfg.JWTSecret),
		issuer:     cfg.Issuer,
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}, nil
}

func (m *JWTManager) GenerateAccessToken(user *model.Users) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.accessTTL)
	claims := AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   formatSubject(user.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
		},
		RoleID: user.RoleId,
		Email:  user.Email,
		Type:   TokenTypeAccess,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	return signed, expiresAt, err
}

func (m *JWTManager) GenerateRefreshToken(user *model.Users) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.refreshTTL)
	claims := AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   formatSubject(user.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
		},
		RoleID: user.RoleId,
		Email:  user.Email,
		Type:   TokenTypeRefresh,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	return signed, expiresAt, err
}

func (m *JWTManager) ParseAccessToken(tokenStr string) (*AccessClaims, error) {
	claims, err := m.parse(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.Type != TokenTypeAccess {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}

func (m *JWTManager) ParseRefreshToken(tokenStr string) (*AccessClaims, error) {
	claims, err := m.parse(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.Type != TokenTypeRefresh {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}

func (m *JWTManager) parse(tokenStr string) (*AccessClaims, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	claims := &AccessClaims{}
	token, err := parser.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return m.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}
	if !token.Valid {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}

func formatSubject(id uint) string {
	return fmt.Sprintf("%d", id)
}
