package httpx

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"foodlog/server/internal/service"
)

// 错误码
const (
	CodeOK           = 0
	CodeBadRequest   = 1001
	CodeUnauthorized = 1002
	CodeForbidden    = 1003
	CodeNotFound     = 1004
	CodeConflict     = 1005
	CodeInternal     = 1006
	CodeUploadFailed = 1007
)

// 预定义错误
var ErrUnauthorized = errors.New("unauthorized")

// RespondOK 成功响应
func RespondOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"code": CodeOK, "message": "ok", "data": data})
}

// RespondError 错误响应
func RespondError(c *gin.Context, httpStatus, code int, message string) {
	c.JSON(httpStatus, gin.H{"code": code, "message": message, "data": nil})
}

// RespondErrorWithErr 根据业务错误返回合适的状态码
func RespondErrorWithErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		RespondError(c, http.StatusNotFound, CodeNotFound, "资源不存在")
	case errors.Is(err, service.ErrForbidden):
		RespondError(c, http.StatusForbidden, CodeForbidden, "无权操作此资源")
	case errors.Is(err, service.ErrUsernameTaken):
		RespondError(c, http.StatusConflict, CodeConflict, "用户名已被占用")
	case errors.Is(err, service.ErrEmailTaken):
		RespondError(c, http.StatusConflict, CodeConflict, "邮箱已被注册")
	case errors.Is(err, service.ErrInvalidCreds):
		RespondError(c, http.StatusUnauthorized, CodeUnauthorized, "用户名或密码错误")
	default:
		RespondError(c, http.StatusInternalServerError, CodeInternal, "服务器内部错误")
	}
}

// GetUserID 从 gin 上下文获取当前用户 ID（由 Auth 中间件写入）
func GetUserID(c *gin.Context) string {
	if v, ok := c.Get("user_id"); ok {
		return v.(string)
	}
	return ""
}
