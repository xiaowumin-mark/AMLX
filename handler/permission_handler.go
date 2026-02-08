package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiaowumin-mark/AMLX/service"
)

type PermissionHandler struct {
	svc service.PermissionService
}

func NewPermissionHandler(svc service.PermissionService) *PermissionHandler {
	return &PermissionHandler{svc: svc}
}

func (h *PermissionHandler) Register(rg *gin.RouterGroup) {
	group := rg.Group("/permissions")
	group.POST("", h.createPermission)

	roleGroup := rg.Group("/roles")
	roleGroup.POST("", h.createRole)
	roleGroup.GET("/:id/permissions", h.listRolePermissions)
	roleGroup.POST("/:id/permissions", h.addPermissionToRole)
	roleGroup.DELETE("/:id/permissions/:perm_id", h.removePermissionFromRole)
}

type createPermissionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type createRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type addPermissionRequest struct {
	PermissionID uint `json:"permission_id"`
}

func (h *PermissionHandler) createPermission(c *gin.Context) {
	var req createPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	perm, err := h.svc.CreatePermission(c.Request.Context(), req.Name, req.Description)
	if err != nil {
		handlePermissionError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"permission": perm})
}

func (h *PermissionHandler) createRole(c *gin.Context) {
	var req createRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	role, err := h.svc.CreateRole(c.Request.Context(), req.Name, req.Description)
	if err != nil {
		handlePermissionError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"role": role})
}

func (h *PermissionHandler) addPermissionToRole(c *gin.Context) {
	roleID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req addPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if err := h.svc.AddPermissionToRole(c.Request.Context(), roleID, req.PermissionID); err != nil {
		handlePermissionError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *PermissionHandler) removePermissionFromRole(c *gin.Context) {
	roleID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	permID, err := parseUintParam(c, "perm_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid perm_id"})
		return
	}
	if err := h.svc.RemovePermissionFromRole(c.Request.Context(), roleID, permID); err != nil {
		handlePermissionError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *PermissionHandler) listRolePermissions(c *gin.Context) {
	roleID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	perms, err := h.svc.ListPermissionsByRole(c.Request.Context(), roleID)
	if err != nil {
		handlePermissionError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"permissions": perms})
}

func handlePermissionError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
	case errors.Is(err, service.ErrRoleExists), errors.Is(err, service.ErrPermissionExists):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrRoleNotFound), errors.Is(err, service.ErrPermissionNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
