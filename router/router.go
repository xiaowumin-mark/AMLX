package router

import (
	"github.com/gin-gonic/gin"
	"github.com/xiaowumin-mark/AMLX/config"
	"github.com/xiaowumin-mark/AMLX/handler"
	"github.com/xiaowumin-mark/AMLX/logx"
	"github.com/xiaowumin-mark/AMLX/middleware"
	"github.com/xiaowumin-mark/AMLX/service"
)

func New(cfg *config.Config, userHandler *handler.UserHandler, authHandler *handler.AuthHandler, permissionHandler *handler.PermissionHandler, authSvc service.AuthService, permSvc service.PermissionService) *gin.Engine {
	engine := gin.New()
	if cfg.Server.Log {
		engine.Use(gin.LoggerWithWriter(logx.Writer()))
	}
	engine.Use(gin.Recovery())

	api := engine.Group("/api/v1")
	auth := middleware.NewAuth(authSvc)
	authHandler.Register(api, auth.Required())

	adminOnly := middleware.RequirePermission(permSvc, "admin")
	protected := api.Group("")
	protected.Use(auth.Required())

	userGroup := protected.Group("")
	userGroup.Use(adminOnly)
	userHandler.Register(userGroup)

	permGroup := protected.Group("")
	permGroup.Use(adminOnly)
	permissionHandler.Register(permGroup)

	return engine
}
