package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xiaowumin-mark/AMLX/middleware"
	"github.com/xiaowumin-mark/AMLX/service"
)

type AuthHandler struct {
	auth service.AuthService
	users service.UserService
}

func NewAuthHandler(auth service.AuthService, users service.UserService) *AuthHandler {
	return &AuthHandler{auth: auth, users: users}
}

func (h *AuthHandler) Register(rg *gin.RouterGroup, authRequired gin.HandlerFunc) {
	group := rg.Group("/auth")
	group.POST("/register", h.register)
	group.POST("/login", h.login)
	group.POST("/refresh", h.refresh)
	group.POST("/logout", h.logout)

	group.Use(authRequired)
	group.GET("/me", h.me)
	group.POST("/logout_all", h.logoutAll)
	group.POST("/change_password", h.changePassword)
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	RoleID   uint   `json:"role_id"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type authResponse struct {
	User   userResponse   `json:"user"`
	Tokens service.TokenPair `json:"tokens"`
}

func (h *AuthHandler) register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	user, tokens, err := h.auth.Register(c.Request.Context(), service.RegisterRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		RoleID:   req.RoleID,
	})
	if err != nil {
		handleAuthError(c, err)
		return
	}
	c.JSON(http.StatusCreated, authResponse{
		User:   toUserResponse(user),
		Tokens: *tokens,
	})
}

func (h *AuthHandler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	user, tokens, err := h.auth.Login(c.Request.Context(), service.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		handleAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, authResponse{
		User:   toUserResponse(user),
		Tokens: *tokens,
	})
}

func (h *AuthHandler) refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	token := strings.TrimSpace(req.RefreshToken)
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token required"})
		return
	}

	tokens, err := h.auth.Refresh(c.Request.Context(), token)
	if err != nil {
		handleAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"tokens": tokens})
}

func (h *AuthHandler) logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	token := strings.TrimSpace(req.RefreshToken)
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token required"})
		return
	}
	if err := h.auth.Logout(c.Request.Context(), token); err != nil {
		handleAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AuthHandler) logoutAll(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if err := h.auth.LogoutAll(c.Request.Context(), userID); err != nil {
		handleAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AuthHandler) changePassword(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if err := h.auth.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		handleAuthError(c, err)
		return
	}
	_ = h.auth.LogoutAll(c.Request.Context(), userID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AuthHandler) me(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user, err := h.users.GetByID(c.Request.Context(), userID)
	if err != nil {
		handleAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": toUserResponse(user)})
}

func handleAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
	case errors.Is(err, service.ErrEmailExists):
		c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
	case errors.Is(err, service.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
	case errors.Is(err, service.ErrUserBanned):
		c.JSON(http.StatusForbidden, gin.H{"error": "user banned"})
	case errors.Is(err, service.ErrRegistrationClosed):
		c.JSON(http.StatusForbidden, gin.H{"error": "registration disabled"})
	case errors.Is(err, service.ErrRefreshTokenInvalid):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token invalid"})
	case errors.Is(err, service.ErrTokenExpired):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
	case errors.Is(err, service.ErrTokenInvalid):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token invalid"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
