package ai

import (
	"fmt"
	"pixelpunk/internal/controllers/websocket"
	"pixelpunk/internal/models"
	"pixelpunk/internal/services/setting"
	ws "pixelpunk/internal/websocket"
	"pixelpunk/pkg/common"
	"pixelpunk/pkg/logger"
	"sync"
	"time"

	"gorm.io/gorm"
)

var globalTaggingService *TaggingService
var globalTaggingMutex sync.RWMutex // 保护 globalTaggingService 的并发访问
var hooksRegistered = false

var aiScanTimer *time.Timer
var aiScanMutex sync.Mutex

func RegisterAISettingHooks() {
	if hooksRegistered {
		return
	}
	hooksRegistered = true

	setting.RegisterSettingChangeHandler("ai", "ai_enabled", func(value string) {
		enabled := value != "" && value != "false" && value != "0" && value != "False" && value != "FALSE"

		svc := GetGlobalTaggingService()
		if enabled && svc == nil {
			if err := InitGlobalTaggingQueue(); err != nil {
				return
			}
			svc = GetGlobalTaggingService()
			if svc != nil && !svc.IsPaused() {
				go func() {
					if n, err := EnqueueAllPending(1000); err == nil && n > 0 {
					}
				}()
			}
		} else if !enabled && svc != nil {
			svc.Stop()
			SetGlobalTaggingService(nil)
		}
	})

	setting.RegisterSettingChangeHandler("ai", "ai_auto_processing_enabled", func(value string) {
		enabled := value != "" && value != "false" && value != "0" && value != "False" && value != "FALSE"

		svc := GetGlobalTaggingService()
		if enabled && svc == nil {
			if err := InitGlobalTaggingQueue(); err != nil {
				return
			}
			svc = GetGlobalTaggingService()
			if svc != nil {
				go func() {
					if n, err := EnqueueAllPending(1000); err == nil && n > 0 {
					}
				}()
			}
			return
		}

		if svc == nil {
			return
		}
		if enabled {
			svc.Resume()
		} else {
			svc.Pause()
		}
	})

	setting.RegisterSettingChangeHandler("ai", "ai_concurrency", func(value string) {
		svc := GetGlobalTaggingService()
		if svc == nil {
			return
		}
		// 直接从数据库读取配置（绕过缓存）
		concurrency := setting.GetIntDirectFromDB("ai", "ai_concurrency", 3)
		if concurrency > 0 {
			_ = svc.UpdateConcurrency(concurrency)
		}
	})

	handleAIConfigChange := func() {
		aiEnabled := setting.GetBoolDirectFromDB("ai", "ai_enabled", false)
		if !aiEnabled {
			return
		}

		svc := GetGlobalTaggingService()
		if svc == nil {
			if err := InitGlobalTaggingQueue(); err != nil {
				logger.Error("[AI服务] 初始化队列失败: %v", err)
				return
			}
			svc = GetGlobalTaggingService()
		}

		if svc != nil && !svc.IsPaused() {
			aiScanMutex.Lock()
			if aiScanTimer != nil {
				aiScanTimer.Stop()
			}
			aiScanTimer = time.AfterFunc(300*time.Millisecond, func() {
				if n, err := EnqueueAllPending(1000); err == nil && n > 0 {
				}
			})
			aiScanMutex.Unlock()
		}
	}

	aiCriticalKeys := []string{"ai_enabled", "ai_api_key", "ai_base_url", "ai_model"}
	for _, key := range aiCriticalKeys {
		setting.RegisterSettingChangeHandler("ai", key, func(value string) {
			handleAIConfigChange()
		})
	}

}

func SetGlobalTaggingService(service *TaggingService) {
	globalTaggingMutex.Lock()
	defer globalTaggingMutex.Unlock()
	globalTaggingService = service
}

func GetGlobalTaggingService() *TaggingService {
	globalTaggingMutex.RLock()
	defer globalTaggingMutex.RUnlock()
	return globalTaggingService
}

func InitGlobalTaggingQueue() error {
	if GetGlobalTaggingService() != nil {
		return nil
	}

	aiEnabled := setting.GetBool("ai", "ai_enabled", false)
	if !aiEnabled {
		return nil
	}

	db := GetDBFromContext()
	if db == nil {
		return fmt.Errorf("无法获取数据库连接")
	}

	svc := NewTaggingServiceWithConfig(db)
	if svc == nil {
		return fmt.Errorf("创建TaggingService失败")
	}

	SetGlobalTaggingService(svc)

	autoProcessing := setting.GetBool("ai", "ai_auto_processing_enabled", true)
	if !autoProcessing {
		svc.Pause()
	}

	go func() {
		_, _ = RecoverPendingOnStartup(1000)
		_, _ = EnqueueAllPending(1000)
	}()

	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if s := GetGlobalTaggingService(); s != nil {
				threshold := setting.GetInt("ai", "pending_stuck_threshold_minutes", 5)
				count, err := ResetStuckPendingFiles(threshold)
				if err != nil {
					continue
				}
				if count > 0 {
					if !s.IsPaused() {
						_, _ = EnqueueAllPending(1000)
					}
				}
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			retention := setting.GetInt("ai", "ai_job_retention_days", 14)
			_, _ = CleanOldAIJobs(retention)
		}
	}()

	return nil
}

func RecoverPendingOnStartup(limit int) (int, error) {
	db := GetDBFromContext()
	if db == nil {
		return 0, fmt.Errorf("无法获取数据库连接")
	}

	var affectedIDs []string
	err := db.Transaction(func(tx *gorm.DB) error {
		q := tx.Table("file").Where("ai_tagging_status = ?", common.AITaggingStatusPending)
		if limit > 0 {
			q = q.Limit(limit)
		}
		if err := q.Pluck("id", &affectedIDs).Error; err != nil {
			return err
		}
		if len(affectedIDs) == 0 {
			return nil
		}
		return tx.Model(&models.File{}).Where("id IN ?", affectedIDs).Updates(map[string]interface{}{
			"ai_tagging_status": common.AITaggingStatusNone,
			"ai_tagging_tries":  0,
		}).Error
	})
	if err != nil {
		return 0, err
	}

	if n := len(affectedIDs); n > 0 {
		if svc := GetGlobalTaggingService(); svc != nil {
			var images []models.File
			if err := db.Where("id IN ?", affectedIDs).Find(&images).Error; err == nil && len(images) > 0 {
				svc.BatchProcessFiles(images)
			}
		}
	}
	return len(affectedIDs), nil
}

func RefreshGlobalConcurrency() error {
	svc := GetGlobalTaggingService()
	if svc == nil {
		return fmt.Errorf("全局TaggingService未初始化")
	}
	return svc.RefreshConcurrencyFromConfig()
}

func AddFileToQueue(file models.File) error {
	svc := GetGlobalTaggingService()
	if svc == nil {
		_ = InitGlobalTaggingQueue()
		svc = GetGlobalTaggingService()
		if svc == nil {
			return nil
		}
	}

	db := GetDBFromContext()
	if db == nil {
		return fmt.Errorf("无法获取数据库连接")
	}

	if err := db.Model(&models.File{}).Where("id = ?", file.ID).
		Updates(map[string]interface{}{
			"ai_tagging_status":    common.AITaggingStatusPending,
			"ai_last_heartbeat_at": time.Now(),
		}).Error; err != nil {
		logger.Error("更新文件AI状态失败: %v", err)
	}

	if svc.taskQueue != nil {
		if err := svc.taskQueue.EnqueueUnique(file.ID, 0); err != nil {
			if svc.IsPaused() {
				return nil
			}
			svc.ProcessSingleFile(file)
			return nil
		}
		svc.notifyQueueStatsChange()
		if svc.IsPaused() {
			websocket.BroadcastToAdmins(ws.MessageTypeAnnouncement, map[string]interface{}{
				"title":   "AI 队列暂停（已入队）",
				"content": fmt.Sprintf("文件 %s 已加入AI队列。当前处于暂停状态，恢复后自动执行。", file.ID),
				"ts":      time.Now().Unix(),
			})
		}
		return nil
	}
	if svc.IsPaused() {
		return nil
	}
	svc.ProcessSingleFile(file)
	return nil
}
