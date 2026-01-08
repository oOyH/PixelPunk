package middleware

import (
	"net"
	"pixelpunk/internal/models"
	"pixelpunk/pkg/common"
	"pixelpunk/pkg/database"
	"pixelpunk/pkg/logger"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const userActivityMinUpdateInterval = 60 * time.Second

var lastUserActivityUpdateAt sync.Map // map[uint]int64 (unix nanos)

func getClientIP(c *gin.Context) string {
	if xForwardedFor := c.GetHeader("X-Forwarded-For"); xForwardedFor != "" {
		if ip, _, err := net.SplitHostPort(xForwardedFor); err == nil {
			return ip
		}
		return xForwardedFor
	}

	if xRealIP := c.GetHeader("X-Real-IP"); xRealIP != "" {
		return xRealIP
	}

	if ip, _, err := net.SplitHostPort(c.Request.RemoteAddr); err == nil {
		return ip
	}

	return c.Request.RemoteAddr
}

func shouldUpdateUserActivity(userID uint, now time.Time) bool {
	if userID == 0 {
		return false
	}

	nowUnix := now.UnixNano()
	if last, ok := lastUserActivityUpdateAt.Load(userID); ok {
		if lastUnix, ok := last.(int64); ok && nowUnix-lastUnix < userActivityMinUpdateInterval.Nanoseconds() {
			return false
		}
	}

	lastUserActivityUpdateAt.Store(userID, nowUnix)
	return true
}

func updateUserActivity(userID uint, clientIP string, now time.Time) {
	go func() {
		db := database.GetDB()
		if db == nil {
			logger.Error("无法获取数据库连接，跳过用户活动更新")
			lastUserActivityUpdateAt.Delete(userID)
			return
		}

		nowJSON := common.JSONTime(now)
		result := db.Model(&models.User{}).
			Where("id = ?", userID).
			Updates(map[string]interface{}{
				"last_activity_at": &nowJSON,
				"last_activity_ip": clientIP,
				"updated_at":       &nowJSON,
			})

		if result.Error != nil {
			logger.Error("更新用户 %d 活动信息失败: %v", userID, result.Error)
		}
	}()
}

func TrackUserActivity() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if _, hasAuthError := c.Get(AuthErrorKey); !hasAuthError {
			if claims := GetCurrentUser(c); claims != nil {
				now := time.Now()
				if !shouldUpdateUserActivity(claims.UserID, now) {
					return
				}

				clientIP := getClientIP(c)

				updateUserActivity(claims.UserID, clientIP, now)
			}
		}
	}
}
