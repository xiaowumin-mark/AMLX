package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaowumin-mark/AMLX/model"
	"github.com/xiaowumin-mark/AMLX/service"
)

type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) Register(rg *gin.RouterGroup) {
	group := rg.Group("/users")
	group.POST("", h.create)
	group.GET("/:id", h.getByID)
	group.GET("", h.getByEmail)
	group.PATCH("/:id", h.update)
	group.PUT("/:id/ban", h.setBan)
}

type createUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	RoleID   uint   `json:"role_id"`
}

type updateUserRequest struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
	RoleID   *uint   `json:"role_id"`
}

type setBanRequest struct {
	Ban bool `json:"ban"`
}

type userResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	RoleID    uint   `json:"role_id"`
	Ban       bool   `json:"ban"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (h *UserHandler) create(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	user, err := h.svc.Create(c.Request.Context(), service.CreateUserRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		RoleID:   req.RoleID,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toUserResponse(user))
}

func (h *UserHandler) getByID(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	user, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toUserResponse(user))
}

func (h *UserHandler) getByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	user, err := h.svc.GetByEmail(c.Request.Context(), email)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toUserResponse(user))
}

func (h *UserHandler) update(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	user, err := h.svc.Update(c.Request.Context(), id, service.UpdateUserRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		RoleID:   req.RoleID,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toUserResponse(user))
}

func (h *UserHandler) setBan(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req setBanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	if err := h.svc.SetBan(c.Request.Context(), id, req.Ban); err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *UserHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
	case errors.Is(err, service.ErrEmailExists):
		c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
	case errors.Is(err, service.ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}

func parseUintParam(c *gin.Context, name string) (uint, error) {
	value := c.Param(name)
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}

func toUserResponse(user *model.Users) userResponse {
	return userResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		RoleID:    user.RoleId,
		Ban:       user.Ban,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}
