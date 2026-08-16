package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"foodlog/server/internal/httpx"
	"foodlog/server/internal/service"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type registerRequest struct {
	Username string `json:"username" binding:"required,min=2,max=50"`
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type updateProfileRequest struct {
	Username  string `json:"username" binding:"omitempty,min=2,max=50"`
	Email     string `json:"email" binding:"omitempty,email,max=255"`
	Password  string `json:"password" binding:"omitempty,min=6,max=72"`
	AvatarURL string `json:"avatar_url"`
}

// Register 注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
		return
	}

	user, token, err := h.authService.Register(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}

	httpx.RespondOK(c, gin.H{"token": token, "user": user})
}

// Login 登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误")
		return
	}

	user, token, err := h.authService.Login(c.Request.Context(), strings.TrimSpace(req.Username), req.Password)
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}

	httpx.RespondOK(c, gin.H{"token": token, "user": user})
}

// Me 获取当前用户信息
func (h *AuthHandler) Me(c *gin.Context) {
	userID := httpx.GetUserID(c)
	profile, err := h.authService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, profile)
}

// UpdateMe 更新当前用户信息
func (h *AuthHandler) UpdateMe(c *gin.Context) {
	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
		return
	}

	userID := httpx.GetUserID(c)
	user, err := h.authService.UpdateProfile(c.Request.Context(), userID,
		req.Username, req.Email, req.Password, req.AvatarURL)
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, user)
}

// Health 健康检查
func (h *AuthHandler) Health(c *gin.Context) {
	httpx.RespondOK(c, gin.H{"status": "ok"})
}
