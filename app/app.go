package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/xiaowumin-mark/AMLX/config"
	"github.com/xiaowumin-mark/AMLX/database"
	"github.com/xiaowumin-mark/AMLX/handler"
	"github.com/xiaowumin-mark/AMLX/logx"
	"github.com/xiaowumin-mark/AMLX/router"
	"github.com/xiaowumin-mark/AMLX/service"
	"github.com/xiaowumin-mark/AMLX/store"
	"gorm.io/gorm"
)

type App struct {
	Config *config.Config
	DB     *gorm.DB
	Router http.Handler
}

func New(cfg *config.Config) (*App, error) {
	db, err := database.Open(cfg.MySQL) // 连接数据库
	if err != nil {
		return nil, err
	}
	if err := database.AutoMigrate(db); err != nil { // 自动迁移
		return nil, err
	}

	userStore := store.NewUserStore(db)                     // 创建用户store
	roleStore := store.NewRoleStore(db)                     // 创建角色store
	permissionStore := store.NewPermissionStore(db)         // 创建权限store
	rolePermissionStore := store.NewRolePermissionStore(db) // 创建角色权限store
	refreshTokenStore := store.NewRefreshTokenStore(db)     // 创建刷新令牌store

	userService := service.NewUserService(userStore, cfg.Auth.BcryptCost) // 创建用户服务
	jwtManager, err := service.NewJWTManager(cfg.Auth)                    // 创建JWT管理器
	if err != nil {
		return nil, err
	}
	authService := service.NewAuthService(cfg.Auth, userStore, refreshTokenStore, jwtManager)          // 创建认证服务
	permissionService := service.NewPermissionService(roleStore, permissionStore, rolePermissionStore) // 创建权限服务

	if cfg.Auth.BootstrapAdminRoleValue() { // 如果允许注册，则创建管理员角色
		adminRoleID, err := permissionService.EnsureAdminRole(context.Background()) // 确保管理员角色
		if err != nil {
			return nil, err
		}
		if cfg.Auth.AllowRegisterValue() { // 如果允许注册，则设置默认角色为管理员角色
			count, err := userStore.Count(context.Background()) // 获取用户数量
			if err != nil {
				return nil, err
			}
			if count == 0 && cfg.Auth.DefaultRoleID == 0 { // 如果没有用户且默认角色为0，则将默认角色设置为管理员角色
				cfg.Auth.DefaultRoleID = adminRoleID
			}
		}
	}

	userHandler := handler.NewUserHandler(userService)                   // 创建用户处理器
	authHandler := handler.NewAuthHandler(authService, userService)      // 创建认证处理器
	permissionHandler := handler.NewPermissionHandler(permissionService) // 创建权限处理器

	engine := router.New(cfg, userHandler, authHandler, permissionHandler, authService, permissionService) // 创建路由

	logx.L().Info("mysql connected and migrated")

	return &App{
		Config: cfg,
		DB:     db,
		Router: engine,
	}, nil
}

func (a *App) Run() error { // 启动服务
	addr := fmt.Sprintf(":%d", a.Config.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      a.Router,
		ReadTimeout:  a.Config.Server.ReadTimeout,
		WriteTimeout: a.Config.Server.WriteTimeout,
		IdleTimeout:  a.Config.Server.IdleTimeout,
	}
	return srv.ListenAndServe()
}
