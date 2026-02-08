package app

import (
	"fmt"
	"net/http"
	"context"

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
	db, err := database.Open(cfg.MySQL)
	if err != nil {
		return nil, err
	}
	if err := database.AutoMigrate(db); err != nil {
		return nil, err
	}

	userStore := store.NewUserStore(db)
	roleStore := store.NewRoleStore(db)
	permissionStore := store.NewPermissionStore(db)
	rolePermissionStore := store.NewRolePermissionStore(db)
	refreshTokenStore := store.NewRefreshTokenStore(db)

	userService := service.NewUserService(userStore, cfg.Auth.BcryptCost)
	jwtManager, err := service.NewJWTManager(cfg.Auth)
	if err != nil {
		return nil, err
	}
	authService := service.NewAuthService(cfg.Auth, userStore, refreshTokenStore, jwtManager)
	permissionService := service.NewPermissionService(roleStore, permissionStore, rolePermissionStore)

	if cfg.Auth.BootstrapAdminRoleValue() {
		adminRoleID, err := permissionService.EnsureAdminRole(context.Background())
		if err != nil {
			return nil, err
		}
		if cfg.Auth.AllowRegisterValue() {
			count, err := userStore.Count(context.Background())
			if err != nil {
				return nil, err
			}
			if count == 0 && cfg.Auth.DefaultRoleID == 0 {
				cfg.Auth.DefaultRoleID = adminRoleID
			}
		}
	}

	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(authService, userService)
	permissionHandler := handler.NewPermissionHandler(permissionService)

	engine := router.New(cfg, userHandler, authHandler, permissionHandler, authService, permissionService)

	logx.L().Info("mysql connected and migrated")

	return &App{
		Config: cfg,
		DB:     db,
		Router: engine,
	}, nil
}

func (a *App) Run() error {
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
