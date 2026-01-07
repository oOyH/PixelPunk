package constants

import "time"

// ==================== AI处理相关常量 ====================

const (
	// AITaggingFailureThreshold AI打标失败阈值，超过此次数将暂停处理
	AITaggingFailureThreshold = 20

	// AIHeartbeatTimeout AI处理心跳超时时间
	AIHeartbeatTimeout = 2 * time.Minute

	// AIMaxRetries AI处理最大重试次数
	AIMaxRetries = 3

	// AIQueueCheckInterval AI队列检查间隔
	AIQueueCheckInterval = 5 * time.Second
)

// ==================== 分片上传相关常量 ====================

const (
	// ChunkSizeMin 最小分片大小 1MB
	ChunkSizeMin = 1 * 1024 * 1024

	// ChunkSizeMax 最大分片大小 10MB
	ChunkSizeMax = 10 * 1024 * 1024

	// ChunkSizeDefault 默认分片大小 5MB
	ChunkSizeDefault = 5 * 1024 * 1024

	// ChunkedUploadSessionTimeout 分片上传会话默认超时时间
	ChunkedUploadSessionTimeout = 24 * time.Hour
)

// ==================== 缓存相关常量 ====================

const (
	// PublicFileCacheMaxAgeDays 公开文件缓存时间（天）
	PublicFileCacheMaxAgeDays = 60

	// PrivateFileCacheMaxAgeHours 私有文件缓存时间（小时）
	PrivateFileCacheMaxAgeHours = 24

	// PublicFileCacheMaxAgeSeconds 公开文件缓存时间（秒）
	PublicFileCacheMaxAgeSeconds = PublicFileCacheMaxAgeDays * 24 * 60 * 60

	// PrivateFileCacheMaxAgeSeconds 私有文件缓存时间（秒）
	PrivateFileCacheMaxAgeSeconds = PrivateFileCacheMaxAgeHours * 60 * 60
)

// ==================== 批量操作相关常量 ====================

const (
	// BatchUploadMaxConcurrent 批量上传最大并发数
	BatchUploadMaxConcurrent = 5

	// BatchDeleteMaxCount 批量删除最大数量
	BatchDeleteMaxCount = 100

	// BatchMoveMaxCount 批量移动最大数量
	BatchMoveMaxCount = 100
)

// ==================== 安全相关常量 ====================

const (
	// MinJWTSecretLength JWT密钥最小长度
	MinJWTSecretLength = 32

	// DefaultLoginExpireHours 默认登录过期时间（小时）
	DefaultLoginExpireHours = 24

	// MaxLoginAttempts 最大登录尝试次数
	MaxLoginAttempts = 5

	// AccountLockoutMinutes 账户锁定时间（分钟）
	AccountLockoutMinutes = 30
)

// ==================== 分页相关常量 ====================

const (
	// DefaultPageSize 默认分页大小
	DefaultPageSize = 20

	// MaxPageSize 最大分页大小
	MaxPageSize = 100
)

// ==================== 文件状态常量 ====================

const (
	// FileStatusNormal 正常状态
	FileStatusNormal = "normal"

	// FileStatusPendingReview 待审核状态
	FileStatusPendingReview = "pending_review"

	// FileStatusPendingDeletion 待删除状态
	FileStatusPendingDeletion = "pending_deletion"

	// FileStatusDeleted 已删除状态
	FileStatusDeleted = "deleted"
)

// ==================== 访问级别常量 ====================

const (
	// AccessLevelPublic 公开访问
	AccessLevelPublic = "public"

	// AccessLevelPrivate 私有访问
	AccessLevelPrivate = "private"

	// AccessLevelProtected 受保护访问
	AccessLevelProtected = "protected"
)

