package model

import (
	"gorm.io/gorm"
)

// 用户数据表
type Users struct {
	gorm.Model
	Name     string `gorm:"not null;unique;size:50"`
	Email    string `gorm:"not null;unique;size:225"`
	Password string `gorm:"not null;size:225"`
	RoleId   uint   `gorm:"not null"`
	Ban      bool   `gorm:"not null;default:false"`
}

// 角色数据表
type Roles struct {
	gorm.Model
	Name        string `gorm:"not null;unique;size:50"`
	Description string `gorm:"not null;size:225"`
}

// 权限数据表
type Permissions struct {
	gorm.Model
	Name        string `gorm:"not null;unique;size:50"`
	Description string `gorm:"not null;size:225"`
}

// 角色权限数据表
type RolePermissions struct {
	gorm.Model
	RoleId       uint `gorm:"not null"`
	PermissionId uint `gorm:"not null"`
}

// 刷新令牌数据表
type RefreshTokens struct {
	gorm.Model
	UserId    uint   `gorm:"not null"`
	Token     string `gorm:"not null;unique;size:225"`
	ExpiredAt int64  `gorm:"not null;autoUpdateTime:milli"`
	Revoked   bool   `gorm:"not null;default:false"`
}
