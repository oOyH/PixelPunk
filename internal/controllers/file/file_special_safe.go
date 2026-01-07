package file

import (
	"os"
	"path/filepath"
	"strings"

	filesvc "pixelpunk/internal/services/file"
	"pixelpunk/pkg/errors"

	"github.com/gin-gonic/gin"
)

func ServeAvatarFileSafe(c *gin.Context) {
	fileName, ok := sanitizeSingleFileName(c.Param("fileName"))
	if !ok {
		errors.HandleError(c, errors.New(errors.CodeInvalidParameter, "非法文件名"))
		return
	}

	baseDir := filepath.Clean(filesvc.AvatarUploadDir)
	fullPath := filepath.Clean(filepath.Join(baseDir, fileName))
	if !isPathUnderBase(fullPath, baseDir) {
		errors.HandleError(c, errors.New(errors.CodeInvalidParameter, "非法文件路径"))
		return
	}

	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		errors.HandleError(c, errors.New(errors.CodeFileNotFound, "头像文件不存在"))
		return
	}
	if err != nil {
		errors.HandleError(c, errors.New(errors.CodeInternal, "读取头像文件失败: "+err.Error()))
		return
	}
	if info.IsDir() {
		errors.HandleError(c, errors.New(errors.CodeInvalidParameter, "非法文件路径"))
		return
	}

	c.File(fullPath)
}

func ServeAdminFileSafe(c *gin.Context) {
	fileName, ok := sanitizeSingleFileName(c.Param("fileName"))
	if !ok {
		errors.HandleError(c, errors.New(errors.CodeInvalidParameter, "非法文件名"))
		return
	}

	baseDir := filepath.Clean(filesvc.FileUploadDir)
	fullPath := filepath.Clean(filepath.Join(baseDir, fileName))
	if !isPathUnderBase(fullPath, baseDir) {
		errors.HandleError(c, errors.New(errors.CodeInvalidParameter, "非法文件路径"))
		return
	}

	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		errors.HandleError(c, errors.New(errors.CodeFileNotFound, "文件不存在"))
		return
	}
	if err != nil {
		errors.HandleError(c, errors.New(errors.CodeInternal, "读取文件失败: "+err.Error()))
		return
	}
	if info.IsDir() {
		errors.HandleError(c, errors.New(errors.CodeInvalidParameter, "非法文件路径"))
		return
	}

	c.File(fullPath)
}

func sanitizeSingleFileName(fileName string) (string, bool) {
	fileName = strings.TrimSpace(fileName)
	if fileName == "" || fileName == "." || fileName == ".." {
		return "", false
	}
	// 只允许单段文件名，禁止分隔符与 Windows 驱动器前缀
	if strings.ContainsAny(fileName, "/\\") || strings.Contains(fileName, ":") || strings.ContainsRune(fileName, '\x00') {
		return "", false
	}
	return fileName, true
}

func isPathUnderBase(fullPath, baseDir string) bool {
	fullPath = filepath.Clean(fullPath)
	baseDir = filepath.Clean(baseDir)

	baseWithSep := baseDir
	if !strings.HasSuffix(baseWithSep, string(os.PathSeparator)) {
		baseWithSep += string(os.PathSeparator)
	}
	return strings.HasPrefix(fullPath, baseWithSep)
}
