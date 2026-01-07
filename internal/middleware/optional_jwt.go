package middleware

import (
	"strings"

	"pixelpunk/internal/services/auth"
	"pixelpunk/pkg/assets"

	"github.com/gin-gonic/gin"
)

// OptionalJWTAuth 可选 JWT 解析中间件：
// - 若请求未携带 token，则直接放行（不写入 AuthErrorKey）
// - 若携带 token，则尝试解析并写入 ContextPayloadKey
// - 若用户被禁用，则直接返回无权限默认图并中止请求
//
// 主要用于 /f /t /s 等文件直链路由：既支持匿名访问 public/private，又允许已登录用户访问 protected。
func OptionalJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			tokenString = c.Query("token")
			if tokenString == "" {
				tokenString, _ = c.Cookie("token")
			}
		} else {
			const prefix = "Bearer "
			if len(tokenString) > len(prefix) && tokenString[:len(prefix)] == prefix {
				tokenString = tokenString[len(prefix):]
			}
		}
		tokenString = strings.TrimSpace(tokenString)

		// 未携带 token，直接放行
		if tokenString == "" {
			c.Next()
			return
		}

		jwtSecret := strings.TrimSpace(getJWTSecret())
		if jwtSecret == "" {
			c.Set(AuthErrorKey, "系统配置错误：JWT密钥未设置")
			c.Next()
			return
		}

		claims, err := auth.ParseToken(tokenString, jwtSecret)
		if err != nil {
			c.Set(AuthErrorKey, "无效的认证令牌")
			c.Next()
			return
		}

		// 过期检查（防御性处理：ExpiresAt 可能为空）
		if claims.ExpiresAt == nil || claims.ExpiresAt.Unix() < auth.GetCurrentTimestamp() {
			c.Set(AuthErrorKey, "认证令牌已过期")
			c.Next()
			return
		}

		c.Set(ContextPayloadKey, claims)

		// 禁用用户即时拦截（文件直链使用默认图响应，避免返回 JSON）
		if !checkUserActive(claims) {
			assets.ServeDefaultFile(c, assets.FileTypeUnauthorized)
			return
		}

		c.Next()
	}
}
