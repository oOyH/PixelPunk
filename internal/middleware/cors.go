package middleware

import (
	"net/url"
	"pixelpunk/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

/* CORSMiddleware 跨域(CORS)中间件：
 * - 默认允许任意 Origin，但不允许携带凭证（避免反射 Origin + credentials 的安全风险）
 * - 当 Origin 与当前 Host / 配置的 baseUrl 匹配时，才允许携带凭证
 */
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.Request.Header.Get("Origin"))

		allowOrigin := ""
		allowCredentials := false

		if origin == "" {
			// 非浏览器请求通常没有 Origin
			allowOrigin = "*"
		} else {
			allowOrigin = origin
			allowCredentials = isTrustedOrigin(origin, c.Request.Host, utils.GetBaseUrl())
		}

		if allowOrigin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			if allowOrigin != "*" {
				c.Writer.Header().Set("Vary", "Origin")
			}
			if allowCredentials {
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")

		requestedHeaders := c.Request.Header.Get("Access-Control-Request-Headers")
		baseAllowedHeaders := "Authorization, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Accept, X-Requested-With, X-API-Key, x-pixelpunk-key"
		if requestedHeaders != "" {
			c.Writer.Header().Set("Access-Control-Allow-Headers", baseAllowedHeaders+", "+requestedHeaders)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Headers", baseAllowedHeaders)
		}

		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Disposition, Content-Type, X-Request-Id, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func isTrustedOrigin(origin, requestHost, baseURL string) bool {
	originURL, err := url.Parse(origin)
	if err != nil {
		return false
	}

	originHost := strings.ToLower(originURL.Hostname())
	if originHost == "" {
		return false
	}

	reqHost := strings.ToLower(hostnameFromHostPort(requestHost))
	if reqHost != "" && reqHost == originHost {
		return true
	}

	baseURL = strings.TrimSpace(baseURL)
	if baseURL != "" {
		if baseParsed, err := url.Parse(baseURL); err == nil {
			baseHost := strings.ToLower(baseParsed.Hostname())
			if baseHost != "" && baseHost == originHost {
				return true
			}
		}
	}

	isLocal := func(h string) bool { return h == "localhost" || h == "127.0.0.1" }
	if isLocal(originHost) && isLocal(reqHost) {
		return true
	}

	return false
}

func hostnameFromHostPort(hostport string) string {
	hostport = strings.TrimSpace(hostport)
	if hostport == "" {
		return ""
	}
	// IPv6: [::1]:9520
	if strings.HasPrefix(hostport, "[") {
		if idx := strings.Index(hostport, "]"); idx > 1 {
			return hostport[1:idx]
		}
	}
	if idx := strings.Index(hostport, ":"); idx > 0 {
		return hostport[:idx]
	}
	return hostport
}
