package file

import (
	"context"
	"sync"
)

/* Centralized constants (no behavior change) */

const (
	StatusPendingDeletion = "pending_deletion"

	AccessPublic    = "public"
	AccessPrivate   = "private"
	AccessProtected = "protected"

	DefaultPageSize = 20

	// BatchUploadMaxConcurrent 批量上传最大并发数
	BatchUploadMaxConcurrent = 5
)

/* 全局Context管理，用于优雅关闭 */
var (
	globalCtx       context.Context
	globalCancel    context.CancelFunc
	globalCtxMutex  sync.RWMutex
	serviceInitOnce sync.Once
)

// InitUploadService 初始化上传服务的全局context
func InitUploadService() {
	serviceInitOnce.Do(func() {
		globalCtxMutex.Lock()
		defer globalCtxMutex.Unlock()
		globalCtx, globalCancel = context.WithCancel(context.Background())
	})
}

// ShutdownUploadService 优雅关闭上传服务，取消所有后台任务
func ShutdownUploadService() {
	globalCtxMutex.Lock()
	defer globalCtxMutex.Unlock()
	if globalCancel != nil {
		globalCancel()
	}
}

// GetServiceContext 获取服务的全局context
func GetServiceContext() context.Context {
	globalCtxMutex.RLock()
	defer globalCtxMutex.RUnlock()
	if globalCtx == nil {
		return context.Background()
	}
	return globalCtx
}

// IsServiceShuttingDown 检查服务是否正在关闭
func IsServiceShuttingDown() bool {
	ctx := GetServiceContext()
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
