package routes

import (
	"encoding/json"
	"net/http"
	"os"
	"pixelpunk/internal/middleware"
	"pixelpunk/internal/static"
	"strings"

	"github.com/gin-gonic/gin"
)

func RegisterClientRoutes(r *gin.Engine) {
	distFS := static.GetDistFS()

	r.GET("/runtime-config.js", func(c *gin.Context) {
		apiBaseURL := strings.TrimSpace(firstNonEmpty(os.Getenv("VITE_API_BASE_URL"), os.Getenv("PIXELPUNK_API_BASE_URL")))
		siteDomain := strings.TrimSpace(firstNonEmpty(os.Getenv("VITE_SITE_DOMAIN"), os.Getenv("PIXELPUNK_SITE_DOMAIN")))

		if apiBaseURL == "" {
			apiBaseURL = "/api/v1"
		}

		if siteDomain == "" {
			siteDomain = getRequestOrigin(c)
		}

		payload, _ := json.Marshal(map[string]string{
			"VITE_API_BASE_URL": apiBaseURL,
			"VITE_SITE_DOMAIN":  siteDomain,
		})

		c.Header("Content-Type", "application/javascript; charset=utf-8")
		c.Header("Cache-Control", "no-store")
		c.Status(http.StatusOK)

		_, _ = c.Writer.Write([]byte("window.__VITE_RUNTIME_CONFIG__ = "))
		_, _ = c.Writer.Write(payload)
		_, _ = c.Writer.Write([]byte(";\n"))
		_, _ = c.Writer.Write([]byte("window.__VITE_API_BASE_URL__ = window.__VITE_RUNTIME_CONFIG__.VITE_API_BASE_URL;\n"))
		_, _ = c.Writer.Write([]byte("window.__VITE_SITE_DOMAIN__ = window.__VITE_RUNTIME_CONFIG__.VITE_SITE_DOMAIN;\n"))
	})

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/debug/") {
			c.Next()
			return
		}
		middleware.StaticFileHandler(distFS)(c)
	})
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func getRequestOrigin(c *gin.Context) string {
	proto := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto"))
	if proto != "" {
		// Only keep the first value, e.g. "https, http"
		if idx := strings.Index(proto, ","); idx >= 0 {
			proto = strings.TrimSpace(proto[:idx])
		}
	}
	if proto == "" {
		if c.Request.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}

	host := strings.TrimSpace(c.GetHeader("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(c.Request.Host)
	}
	if host == "" {
		return ""
	}
	return proto + "://" + host
}
