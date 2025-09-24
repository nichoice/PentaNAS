package services

import (
	"context"
	"os"
	"path/filepath"
	"pnas/internal/logging"
	"pnas/internal/models"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

// FilesystemMonitor 文件系统监控器
type FilesystemMonitor struct {
	watcher       *fsnotify.Watcher
	auditService  *AuditService
	watchedDirs   map[string]bool
	excludeRules  []string
	recursive     bool
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	mu            sync.RWMutex
	logger        *zap.Logger
	renameMap     map[string]time.Time // 用于处理重命名操作
	renameMu      sync.RWMutex
}

var (
	filesystemMonitorInstance *FilesystemMonitor
	filesystemMonitorOnce     sync.Once
)

// GetFilesystemMonitor 获取文件系统监控器单例
func GetFilesystemMonitor() *FilesystemMonitor {
	filesystemMonitorOnce.Do(func() {
		filesystemMonitorInstance = NewFilesystemMonitor()
	})
	return filesystemMonitorInstance
}

// NewFilesystemMonitor 创建新的文件系统监控器
func NewFilesystemMonitor() *FilesystemMonitor {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logging.Logger.Fatal("Failed to create filesystem watcher", zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())

	monitor := &FilesystemMonitor{
		watcher:      watcher,
		auditService: GetAuditService(),
		watchedDirs:  make(map[string]bool),
		excludeRules: []string{
			".DS_Store",
			".Spotlight-V100",
			".Trashes",
			".fseventsd",
			".TemporaryItems",
			"Thumbs.db",
			"desktop.ini",
			"~$*",
			".tmp",
			".temp",
			".swp",
			".~",
			"#*",
		},
		recursive: true,
		ctx:       ctx,
		cancel:    cancel,
		logger:    logging.Logger,
		renameMap: make(map[string]time.Time),
	}

	return monitor
}

// Start 启动文件系统监控
func (fm *FilesystemMonitor) Start() {
	fm.wg.Add(2)
	go fm.eventProcessor()
	go fm.renameCleanup()
	fm.logger.Info("Filesystem monitor started")
}

// Stop 停止文件系统监控
func (fm *FilesystemMonitor) Stop() {
	fm.cancel()
	fm.watcher.Close()
	fm.wg.Wait()
	fm.logger.Info("Filesystem monitor stopped")
}

// AddWatchPath 添加监控路径
func (fm *FilesystemMonitor) AddWatchPath(path string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// 清理路径
	cleanPath := filepath.Clean(path)

	// 检查路径是否存在
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return err
	}

	// 检查是否已经在监控
	if fm.watchedDirs[cleanPath] {
		return nil
	}

	// 添加监控
	if err := fm.watcher.Add(cleanPath); err != nil {
		return err
	}

	fm.watchedDirs[cleanPath] = true
	fm.logger.Info("Added watch path", zap.String("path", cleanPath))

	// 如果启用递归监控，添加所有子目录
	if fm.recursive {
		return fm.addSubDirectories(cleanPath)
	}

	return nil
}

// RemoveWatchPath 移除监控路径
func (fm *FilesystemMonitor) RemoveWatchPath(path string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	cleanPath := filepath.Clean(path)

	if !fm.watchedDirs[cleanPath] {
		return nil
	}

	if err := fm.watcher.Remove(cleanPath); err != nil {
		return err
	}

	delete(fm.watchedDirs, cleanPath)
	fm.logger.Info("Removed watch path", zap.String("path", cleanPath))

	return nil
}

// addSubDirectories 递归添加子目录监控
func (fm *FilesystemMonitor) addSubDirectories(rootPath string) error {
	return filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续处理其他目录
		}

		if info.IsDir() && path != rootPath {
			if fm.shouldExclude(filepath.Base(path)) {
				return filepath.SkipDir
			}

			if !fm.watchedDirs[path] {
				if err := fm.watcher.Add(path); err != nil {
					fm.logger.Warn("Failed to add watch path",
						zap.String("path", path),
						zap.Error(err))
				} else {
					fm.watchedDirs[path] = true
					fm.logger.Debug("Added recursive watch path", zap.String("path", path))
				}
			}
		}

		return nil
	})
}

// eventProcessor 事件处理器
func (fm *FilesystemMonitor) eventProcessor() {
	defer fm.wg.Done()

	for {
		select {
		case <-fm.ctx.Done():
			return

		case event, ok := <-fm.watcher.Events:
			if !ok {
				return
			}
			fm.processEvent(event)

		case err, ok := <-fm.watcher.Errors:
			if !ok {
				return
			}
			fm.logger.Error("Filesystem watcher error", zap.Error(err))
		}
	}
}

// processEvent 处理单个文件系统事件
func (fm *FilesystemMonitor) processEvent(event fsnotify.Event) {
	// 检查是否应该排除此文件
	if fm.shouldExclude(filepath.Base(event.Name)) {
		return
	}

	fm.logger.Debug("Filesystem event",
		zap.String("name", event.Name),
		zap.String("op", event.Op.String()))

	var operation models.AuditOperation
	var status models.AuditStatus = models.AuditStatusSuccess

	switch {
	case event.Has(fsnotify.Create):
		operation = models.AuditOperationCreate
		// 如果创建的是目录且启用递归监控，添加监控
		if info, err := os.Stat(event.Name); err == nil && info.IsDir() && fm.recursive {
			fm.AddWatchPath(event.Name)
		}

	case event.Has(fsnotify.Write):
		operation = models.AuditOperationUpdate

	case event.Has(fsnotify.Remove):
		operation = models.AuditOperationDelete
		// 如果删除的是目录，从监控中移除
		if fm.watchedDirs[event.Name] {
			fm.RemoveWatchPath(event.Name)
		}

	case event.Has(fsnotify.Rename):
		fm.handleRename(event.Name)
		return // 重命名单独处理，不记录审计日志

	case event.Has(fsnotify.Chmod):
		// 权限变化通常不需要审计，跳过
		return

	default:
		fm.logger.Debug("Unhandled filesystem event",
			zap.String("name", event.Name),
			zap.String("op", event.Op.String()))
		return
	}

	// 记录审计事件
	auditEvent := &AuditEvent{
		UserID:    "system", // 文件系统事件标记为系统操作
		Username:  "system",
		Operation: operation,
		FilePath:  event.Name,
		Status:    status,
		Source:    models.AuditSourceFilesystem,
		Duration:  0,
		Metadata: map[string]interface{}{
			"fs_event": event.Op.String(),
		},
	}

	fm.auditService.LogEvent(auditEvent)
}

// handleRename 处理重命名操作
func (fm *FilesystemMonitor) handleRename(path string) {
	fm.renameMu.Lock()
	defer fm.renameMu.Unlock()

	now := time.Now()

	// 检查是否有对应的重命名配对
	var foundPair string
	cutoff := now.Add(-time.Second) // 1秒内的重命名认为是配对的

	for oldPath, timestamp := range fm.renameMap {
		if timestamp.After(cutoff) {
			foundPair = oldPath
			delete(fm.renameMap, oldPath)
			break
		}
	}

	if foundPair != "" {
		// 这是重命名操作的目标文件
		auditEvent := &AuditEvent{
			UserID:    "system",
			Username:  "system",
			Operation: models.AuditOperationRename,
			FilePath:  path,
			Status:    models.AuditStatusSuccess,
			Source:    models.AuditSourceFilesystem,
			Metadata: map[string]interface{}{
				"old_path": foundPair,
				"new_path": path,
			},
		}
		fm.auditService.LogEvent(auditEvent)
	} else {
		// 这是重命名操作的源文件，等待配对
		fm.renameMap[path] = now
	}
}

// renameCleanup 清理过期的重命名记录
func (fm *FilesystemMonitor) renameCleanup() {
	defer fm.wg.Done()

	ticker := time.NewTicker(time.Minute * 5) // 每5分钟清理一次
	defer ticker.Stop()

	for {
		select {
		case <-fm.ctx.Done():
			return
		case <-ticker.C:
			fm.cleanupExpiredRenames()
		}
	}
}

// cleanupExpiredRenames 清理过期的重命名记录
func (fm *FilesystemMonitor) cleanupExpiredRenames() {
	fm.renameMu.Lock()
	defer fm.renameMu.Unlock()

	cutoff := time.Now().Add(-time.Minute * 10) // 10分钟前

	for path, timestamp := range fm.renameMap {
		if timestamp.Before(cutoff) {
			delete(fm.renameMap, path)
		}
	}
}

// shouldExclude 检查是否应该排除此文件/目录
func (fm *FilesystemMonitor) shouldExclude(name string) bool {
	for _, rule := range fm.excludeRules {
		if matched := fm.matchPattern(name, rule); matched {
			return true
		}
	}
	return false
}

// matchPattern 简单的模式匹配
func (fm *FilesystemMonitor) matchPattern(name, pattern string) bool {
	if pattern == name {
		return true
	}

	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(name, pattern[1:]) {
		return true
	}

	if strings.HasSuffix(pattern, "*") && strings.HasPrefix(name, pattern[:len(pattern)-1]) {
		return true
	}

	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			return strings.HasPrefix(name, parts[0]) && strings.HasSuffix(name, parts[1])
		}
	}

	return false
}

// SetExcludeRules 设置排除规则
func (fm *FilesystemMonitor) SetExcludeRules(rules []string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.excludeRules = rules
}

// GetWatchedDirs 获取当前监控的目录列表
func (fm *FilesystemMonitor) GetWatchedDirs() []string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	dirs := make([]string, 0, len(fm.watchedDirs))
	for dir := range fm.watchedDirs {
		dirs = append(dirs, dir)
	}
	return dirs
}

// SetRecursive 设置是否递归监控
func (fm *FilesystemMonitor) SetRecursive(recursive bool) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.recursive = recursive
}