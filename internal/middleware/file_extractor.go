package middleware

import (
	"pixelpunk/internal/models"
	"pixelpunk/pkg/assets"
	"pixelpunk/pkg/database"
	"pixelpunk/pkg/logger"
	"strings"

	"github.com/gin-gonic/gin"
)

/* FileInfoExtractorMiddleware 鎻愬彇鏂囦欢淇℃伅鍒颁笂涓嬫枃锛坒ile_info, isThumb锛?*/
func FileInfoExtractorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		var file models.File
		var err error
		var isThumb bool

		db := database.GetDB()
		if db == nil {
			logger.Error("[IMAGE_EXTRACTOR] 数据库未初始化，可能处于安装模式")
			assets.ServeDefaultFile(c, assets.FileTypeNotFound)
			return
		}

		if strings.HasPrefix(path, "/f/") {
			fileID := c.Param("fileID")
			if fileID == "" {
				logger.Error("[IMAGE_EXTRACTOR] 鏂囦欢ID鍙傛暟涓虹┖")
				assets.ServeDefaultFile(c, assets.FileTypeNotFound)
				return
			}

			err = db.Where("id = ?", fileID).First(&file).Error
			isThumb = false

		} else if strings.HasPrefix(path, "/t/") || strings.HasPrefix(path, "/thumb/") {
			fileID := c.Param("fileID")
			if fileID == "" {
				logger.Error("[IMAGE_EXTRACTOR] 鏂囦欢ID鍙傛暟涓虹┖")
				assets.ServeDefaultFile(c, assets.FileTypeNotFound)
				return
			}

			err = db.Where("id = ?", fileID).First(&file).Error
			isThumb = true

		} else if strings.HasPrefix(path, "/s/") {
			shortURL := c.Param("shortURL")
			if shortURL == "" {
				logger.Error("[IMAGE_EXTRACTOR] shortURL鍙傛暟涓虹┖")
				assets.ServeDefaultFile(c, assets.FileTypeNotFound)
				return
			}

			err = db.Where("short_url = ?", shortURL).First(&file).Error
			isThumb = false

		} else {
			logger.Error("[IMAGE_EXTRACTOR] 涓嶆敮鎸佺殑璺緞鏍煎紡: %s", path)
			assets.ServeDefaultFile(c, assets.FileTypeNotFound)
			return
		}

		if err != nil {
			logger.Error("[IMAGE_EXTRACTOR] 鏌ユ壘鏂囦欢澶辫触: %v", err)
			assets.ServeDefaultFile(c, assets.FileTypeNotFound)
			return
		}

		if file.Status == "pending_deletion" {
			assets.ServeDefaultFile(c, assets.FileTypeNotFound)
			return
		}

		c.Set("file_info", file)
		c.Set("isThumb", isThumb)

		c.Next()
	}
}
