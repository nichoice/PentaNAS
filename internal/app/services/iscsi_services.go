package services

import (
	"errors"
	"fmt"

	"pnas/internal/app/dto"
	"pnas/internal/database"
	"pnas/internal/infrastructure/iscsi"
	"pnas/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ISCSITargetService 负责处理与 iSCSI Target 相关的业务逻辑。
type ISCSITargetService struct {
	db  *gorm.DB
	cli iscsi.Client
}

// NewISCSITargetService 返回一个使用全局数据库连接的 Target 服务实例。
func NewISCSITargetService() *ISCSITargetService {
	return NewISCSITargetServiceWithDeps(database.DB, iscsi.NewTargetCLI())
}

// NewISCSITargetServiceWithDeps 允许在单元测试中注入依赖。
func NewISCSITargetServiceWithDeps(db *gorm.DB, cli iscsi.Client) *ISCSITargetService {
	return &ISCSITargetService{db: db, cli: cli}
}

// CreateTarget 基于请求参数创建新的 Target 记录，默认以未激活状态入库。
func (s *ISCSITargetService) CreateTarget(req *dto.CreateISCSITargetRequest) (*dto.ISCSITargetResponse, error) {
	if err := s.cli.CreateTarget(req.Name); err != nil {
		return nil, err
	}

	port := s.resolveTargetPort()
	if err := s.cli.EnsurePortal(req.Name, "0.0.0.0", port); err != nil {
		return nil, err
	}

	status := models.ISCSIStatusInactive
	if req.IsEnabled {
		if err := s.cli.EnableTarget(req.Name); err != nil {
			return nil, err
		}
		status = models.ISCSIStatusActive
	}

	target := models.ISCSITarget{
		Base:      models.Base{ID: uuid.New().String()},
		Name:      req.Name,
		Alias:     req.Alias,
		Comment:   req.Comment,
		IsEnabled: req.IsEnabled,
		Status:    status,
	}

	if err := s.db.Create(&target).Error; err != nil {
		return nil, fmt.Errorf("failed to create target: %w", err)
	}

	return s.convertTargetToResponse(&target), nil
}

// GetTargets 按分页及状态过滤返回 Target 列表。
func (s *ISCSITargetService) GetTargets(page, limit int, status string) ([]dto.ISCSITargetResponse, int64, error) {
	var targets []models.ISCSITarget
	var total int64

	query := s.db.Model(&models.ISCSITarget{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count targets: %w", err)
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&targets).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get targets: %w", err)
	}

	responses := make([]dto.ISCSITargetResponse, len(targets))
	for i, target := range targets {
		responses[i] = *s.convertTargetToResponse(&target)
	}

	return responses, total, nil
}

// GetTargetByID 根据主键 ID 查询单个 Target。
func (s *ISCSITargetService) GetTargetByID(id string) (*dto.ISCSITargetResponse, error) {
	var target models.ISCSITarget
	if err := s.db.First(&target, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("target not found")
		}
		return nil, fmt.Errorf("failed to get target: %w", err)
	}

	return s.convertTargetToResponse(&target), nil
}

// UpdateTarget 更新已有 Target 的别名、备注及启用状态。
func (s *ISCSITargetService) UpdateTarget(id string, req *dto.UpdateISCSITargetRequest) (*dto.ISCSITargetResponse, error) {
	var target models.ISCSITarget
	if err := s.db.First(&target, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("target not found: %w", err)
	}

	previousEnabled := target.IsEnabled
	if req.Alias != nil {
		target.Alias = *req.Alias
	}
	if req.Comment != nil {
		target.Comment = *req.Comment
	}
	if req.IsEnabled != nil {
		target.IsEnabled = *req.IsEnabled
	}

	if req.IsEnabled != nil && previousEnabled != target.IsEnabled {
		var err error
		if target.IsEnabled {
			err = s.cli.EnableTarget(target.Name)
			target.Status = models.ISCSIStatusActive
		} else {
			err = s.cli.DisableTarget(target.Name)
			target.Status = models.ISCSIStatusInactive
		}
		if err != nil {
			return nil, err
		}
	}

	if err := s.db.Save(&target).Error; err != nil {
		return nil, fmt.Errorf("failed to update target: %w", err)
	}

	return s.convertTargetToResponse(&target), nil
}

// DeleteTarget 根据 ID 删除 Target 记录。
func (s *ISCSITargetService) DeleteTarget(id string) error {
	var target models.ISCSITarget
	if err := s.db.First(&target, "id = ?", id).Error; err != nil {
		return fmt.Errorf("target not found: %w", err)
	}

	if err := s.cli.DisableTarget(target.Name); err != nil {
		return err
	}
	if err := s.cli.DeleteTarget(target.Name); err != nil {
		return err
	}

	if err := s.db.Delete(&models.ISCSITarget{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete target: %w", err)
	}
	return nil
}

// StartTarget 调用 targetcli 激活 Target。
func (s *ISCSITargetService) StartTarget(id string) error {
	var target models.ISCSITarget
	if err := s.db.First(&target, "id = ?", id).Error; err != nil {
		return fmt.Errorf("target not found: %w", err)
	}

	if err := s.cli.EnableTarget(target.Name); err != nil {
		return err
	}

	if err := s.db.Model(&models.ISCSITarget{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     models.ISCSIStatusActive,
		"is_enabled": true,
	}).Error; err != nil {
		return fmt.Errorf("failed to start target: %w", err)
	}
	return nil
}

// StopTarget 调用 targetcli 停止 Target。
func (s *ISCSITargetService) StopTarget(id string) error {
	var target models.ISCSITarget
	if err := s.db.First(&target, "id = ?", id).Error; err != nil {
		return fmt.Errorf("target not found: %w", err)
	}

	if err := s.cli.DisableTarget(target.Name); err != nil {
		return err
	}

	if err := s.db.Model(&models.ISCSITarget{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     models.ISCSIStatusInactive,
		"is_enabled": false,
	}).Error; err != nil {
		return fmt.Errorf("failed to stop target: %w", err)
	}
	return nil
}

// GetTargetStatus 返回 Target 的当前状态及统计概览。
func (s *ISCSITargetService) GetTargetStatus(id string) (*dto.ISCSITargetStatusResponse, error) {
	var target models.ISCSITarget
	if err := s.db.First(&target, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("target not found: %w", err)
	}

	// TODO: Get actual session and LUN counts
	return &dto.ISCSITargetStatusResponse{
		ID:             target.ID,
		Name:           target.Name,
		Status:         target.Status,
		ActiveSessions: 0,
		TotalLUNs:      0,
	}, nil
}

func (s *ISCSITargetService) convertTargetToResponse(target *models.ISCSITarget) *dto.ISCSITargetResponse {
	return &dto.ISCSITargetResponse{
		ID:        target.ID,
		Name:      target.Name,
		Alias:     target.Alias,
		Status:    target.Status,
		Comment:   target.Comment,
		IsEnabled: target.IsEnabled,
		CreatedAt: target.CreatedAt,
		UpdatedAt: target.UpdatedAt,
	}
}

func (s *ISCSITargetService) resolveTargetPort() int {
	const fallbackPort = 3260
	var config models.ISCSIGlobalConfig
	if err := s.db.Where("is_active = ?", true).First(&config).Error; err == nil {
		if config.TargetPort > 0 {
			return config.TargetPort
		}
	}
	return fallbackPort
}

func (s *ISCSILUNService) convertLUNToResponse(lun *models.ISCSILUN) *dto.ISCSILUNResponse {
	return &dto.ISCSILUNResponse{
		ID:         lun.ID,
		Name:       lun.Name,
		DeviceType: lun.DeviceType,
		Size:       lun.Size,
		DevicePath: lun.DevicePath,
		Comment:    lun.Comment,
		IsEnabled:  lun.IsEnabled,
		ReadOnly:   lun.ReadOnly,
		BlockSize:  lun.BlockSize,
		CreatedAt:  lun.CreatedAt,
		UpdatedAt:  lun.UpdatedAt,
	}
}

// ISCSILUNService 负责管理 LUN 的生命周期与映射关系。
type ISCSILUNService struct {
	db  *gorm.DB
	cli iscsi.Client
}

// NewISCSILUNService 返回一个 LUN 服务实例。
func NewISCSILUNService() *ISCSILUNService {
	return NewISCSILUNServiceWithDeps(database.DB, iscsi.NewTargetCLI())
}

// NewISCSILUNServiceWithDeps 允许在单元测试中注入依赖。
func NewISCSILUNServiceWithDeps(db *gorm.DB, cli iscsi.Client) *ISCSILUNService {
	return &ISCSILUNService{db: db, cli: cli}
}

// CreateLUN 创建新的 LUN 并关联 backstore。
func (s *ISCSILUNService) CreateLUN(req *dto.CreateISCSILUNRequest) (*dto.ISCSILUNResponse, error) {
	// 生成 backstore 名称
	backstoreName := fmt.Sprintf("backstore_%s", req.Name)

	// 根据设备类型创建 backstore
	switch req.DeviceType {
	case models.ISCSIDeviceFile:
		if req.DevicePath == "" {
			// 如果没有指定路径，在临时目录创建文件（更安全）
			req.DevicePath = fmt.Sprintf("/tmp/iscsi_%s.img", req.Name)
		}
		if err := s.cli.CreateFileBackstore(backstoreName, req.DevicePath, req.Size); err != nil {
			return nil, fmt.Errorf("failed to create file backstore: %w", err)
		}
	case models.ISCSIDeviceBlock:
		if req.DevicePath == "" {
			return nil, fmt.Errorf("device path is required for block device")
		}
		if err := s.cli.CreateBlockBackstore(backstoreName, req.DevicePath); err != nil {
			return nil, fmt.Errorf("failed to create block backstore: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported device type: %s", req.DeviceType)
	}

	// 创建数据库记录
	lun := models.ISCSILUN{
		Base:        models.Base{ID: uuid.New().String()},
		Name:        req.Name,
		DeviceType:  req.DeviceType,
		DevicePath:  req.DevicePath,
		Size:        req.Size,
		BlockSize:   req.BlockSize,
		IsEnabled:   req.IsEnabled,
		ReadOnly:    req.ReadOnly,
		Comment:     req.Comment,
	}

	if lun.BlockSize == 0 {
		lun.BlockSize = 512 // 默认块大小
	}

	if err := s.db.Create(&lun).Error; err != nil {
		// 清理已创建的 backstore
		backendType := "fileio"
		if req.DeviceType == models.ISCSIDeviceBlock {
			backendType = "block"
		}
		_ = s.cli.DeleteBackstore(backendType, backstoreName)
		return nil, fmt.Errorf("failed to create LUN record: %w", err)
	}

	return s.convertLUNToResponse(&lun), nil
}

// GetLUNs 按类型与分页条件返回 LUN 列表。
func (s *ISCSILUNService) GetLUNs(page, limit int, deviceType string) ([]dto.ISCSILUNResponse, int64, error) {
	// TODO: Implement
	return nil, 0, fmt.Errorf("not implemented")
}

// GetLUNByID 预期根据 ID 查询单个 LUN。
func (s *ISCSILUNService) GetLUNByID(id string) (*dto.ISCSILUNResponse, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

// UpdateLUN 预期修改 LUN 的只读状态等属性。
func (s *ISCSILUNService) UpdateLUN(id string, req *dto.UpdateISCSILUNRequest) (*dto.ISCSILUNResponse, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

// DeleteLUN 预留删除 LUN 的实现。
func (s *ISCSILUNService) DeleteLUN(id string) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

// MapLUNToTarget 实现 LUN 到 Target 的映射关系。
func (s *ISCSILUNService) MapLUNToTarget(lunID, targetID string, lun int) (*dto.ISCSILUNMappingResponse, error) {
	// 查找 LUN 记录
	var lunRecord models.ISCSILUN
	if err := s.db.First(&lunRecord, "id = ?", lunID).Error; err != nil {
		return nil, fmt.Errorf("LUN not found: %w", err)
	}

	// 查找 Target 记录
	var target models.ISCSITarget
	if err := s.db.First(&target, "id = ?", targetID).Error; err != nil {
		return nil, fmt.Errorf("target not found: %w", err)
	}

	// 生成 backstore 名称和类型
	backstoreName := fmt.Sprintf("backstore_%s", lunRecord.Name)
	backstoreType := "fileio"
	if lunRecord.DeviceType == models.ISCSIDeviceBlock {
		backstoreType = "block"
	}

	// 在 targetcli 中创建 LUN 映射
	if err := s.cli.CreateLUN(target.Name, lun, backstoreName, backstoreType); err != nil {
		return nil, fmt.Errorf("failed to map LUN to target: %w", err)
	}

	// 更新数据库中的映射关系
	lunRecord.TargetID = targetID
	lunRecord.LUN = lun
	if err := s.db.Save(&lunRecord).Error; err != nil {
		// 回滚 targetcli 操作
		_ = s.cli.DeleteLUN(target.Name, lun)
		return nil, fmt.Errorf("failed to update LUN mapping: %w", err)
	}

	return &dto.ISCSILUNMappingResponse{
		ID:       lunRecord.ID,
		TargetID: targetID,
		LUNID:    lunID,
		LUN:      lun,
		CreatedAt: lunRecord.UpdatedAt,
	}, nil
}

// UnmapLUNFromTarget 负责解除 LUN 的映射。
func (s *ISCSILUNService) UnmapLUNFromTarget(lunID, targetID string) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

// GetLUNsByTarget 将返回指定 Target 已挂载的 LUN 列表。
func (s *ISCSILUNService) GetLUNsByTarget(targetID string) ([]dto.ISCSILUNResponse, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

// ISCSIACLService 负责管理 Initiator 访问控制策略。
type ISCSIACLService struct {
	db *gorm.DB
}

// NewISCSIACLService 返回 ACL 服务实例。
func NewISCSIACLService() *ISCSIACLService {
	return &ISCSIACLService{db: database.DB}
}

// CreateACL 计划创建新的 ACL 记录。
func (s *ISCSIACLService) CreateACL(req *dto.CreateISCSIACLRequest) (*dto.ISCSIACLResponse, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

// GetACLs 将支持按 Target 过滤并分页返回 ACL。
func (s *ISCSIACLService) GetACLs(page, limit int, targetID string) ([]dto.ISCSIACLResponse, int64, error) {
	// TODO: Implement
	return nil, 0, fmt.Errorf("not implemented")
}

// GetACLByID 预留通过 ID 查询 ACL 的逻辑。
func (s *ISCSIACLService) GetACLByID(id string) (*dto.ISCSIACLResponse, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

// UpdateACL 负责更新 ACL 权限及认证信息。
func (s *ISCSIACLService) UpdateACL(id string, req *dto.UpdateISCSIACLRequest) (*dto.ISCSIACLResponse, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

// DeleteACL 删除 ACL 记录。
func (s *ISCSIACLService) DeleteACL(id string) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

// ISCSIConfigService 负责读取与修改全局 iSCSI 配置。
type ISCSIConfigService struct {
	db *gorm.DB
}

// NewISCSIConfigService 返回配置服务实例。
func NewISCSIConfigService() *ISCSIConfigService {
	return &ISCSIConfigService{db: database.DB}
}

// GetGlobalConfig 预留查询全局配置的实现。
func (s *ISCSIConfigService) GetGlobalConfig() (*dto.ISCSIGlobalConfigResponse, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

// UpdateGlobalConfig 计划根据请求更新配置并返回最新值。
func (s *ISCSIConfigService) UpdateGlobalConfig(req *dto.UpdateISCSIGlobalConfigRequest) (*dto.ISCSIGlobalConfigResponse, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

// ResetGlobalConfig 将恢复默认配置值。
func (s *ISCSIConfigService) ResetGlobalConfig() (*dto.ISCSIGlobalConfigResponse, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

// Stub services for compilation
type ISCSIServiceService struct{ db *gorm.DB }
type ISCSISessionService struct{ db *gorm.DB }
type ISCSIConnectionService struct{ db *gorm.DB }
type ISCSIStoragePoolService struct{ db *gorm.DB }
type ISCSIPerformanceService struct{ db *gorm.DB }
type ISCSIAuditService struct{ db *gorm.DB }

// NewISCSIServiceService 返回服务状态管理实例。
func NewISCSIServiceService() *ISCSIServiceService {
	return &ISCSIServiceService{db: database.DB}
}

// GetServiceStatus 预留查询服务状态的实现。
func (s *ISCSIServiceService) GetServiceStatus() (*dto.ISCSIServiceStatusResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

// StartService 计划启动 iSCSI 后台服务。
func (s *ISCSIServiceService) StartService() error {
	return fmt.Errorf("not implemented")
}

// StopService 计划停止 iSCSI 后台服务。
func (s *ISCSIServiceService) StopService() error {
	return fmt.Errorf("not implemented")
}

// RestartService 计划重启 iSCSI 服务。
func (s *ISCSIServiceService) RestartService() error {
	return fmt.Errorf("not implemented")
}

// NewISCSISessionService 返回会话管理实例。
func NewISCSISessionService() *ISCSISessionService {
	return &ISCSISessionService{db: database.DB}
}

// GetSessions 预留分页查询会话的实现。
func (s *ISCSISessionService) GetSessions(page, limit int, targetID string) ([]dto.ISCSISessionResponse, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

// GetSessionByID 预留查询单个会话的实现。
func (s *ISCSISessionService) GetSessionByID(id string) (*dto.ISCSISessionResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

// TerminateSession 计划实现会话强制下线。
func (s *ISCSISessionService) TerminateSession(id string) error {
	return fmt.Errorf("not implemented")
}

// NewISCSIConnectionService 返回连接监控实例。
func NewISCSIConnectionService() *ISCSIConnectionService {
	return &ISCSIConnectionService{db: database.DB}
}

// GetConnections 预留查询当前连接的实现。
func (s *ISCSIConnectionService) GetConnections(page, limit int) ([]dto.ISCSIConnectionResponse, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

// GetConnectionHistory 预留查询历史连接记录的实现。
func (s *ISCSIConnectionService) GetConnectionHistory(page, limit int, from, to string) ([]dto.ISCSIConnectionResponse, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

// NewISCSIStoragePoolService 返回存储池管理实例。
func NewISCSIStoragePoolService() *ISCSIStoragePoolService {
	return &ISCSIStoragePoolService{db: database.DB}
}

// CreateStoragePool 计划创建新的存储池。
func (s *ISCSIStoragePoolService) CreateStoragePool(req *dto.CreateISCSIStoragePoolRequest) (*dto.ISCSIStoragePoolResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetStoragePools 预留分页查询存储池的实现。
func (s *ISCSIStoragePoolService) GetStoragePools(page, limit int) ([]dto.ISCSIStoragePoolResponse, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

// GetStoragePoolByID 预留查询单个存储池的实现。
func (s *ISCSIStoragePoolService) GetStoragePoolByID(id string) (*dto.ISCSIStoragePoolResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

// UpdateStoragePool 计划更新存储池属性。
func (s *ISCSIStoragePoolService) UpdateStoragePool(id string, req *dto.UpdateISCSIStoragePoolRequest) (*dto.ISCSIStoragePoolResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

// DeleteStoragePool 预留删除存储池的实现。
func (s *ISCSIStoragePoolService) DeleteStoragePool(id string) error {
	return fmt.Errorf("not implemented")
}

// NewISCSIPerformanceService 返回性能统计服务实例。
func NewISCSIPerformanceService() *ISCSIPerformanceService {
	return &ISCSIPerformanceService{db: database.DB}
}

// GetPerformanceStats 预留获取整体性能指标的实现。
func (s *ISCSIPerformanceService) GetPerformanceStats(from, to string) (*dto.ISCSIPerformanceStatsResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetTargetStats 计划返回指定 Target 的性能数据。
func (s *ISCSIPerformanceService) GetTargetStats(targetID, from, to string) (*dto.ISCSITargetPerformanceResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetLUNStats 计划返回指定 LUN 的性能数据。
func (s *ISCSIPerformanceService) GetLUNStats(lunID, from, to string) (*dto.ISCSILUNPerformanceResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

// NewISCSIAuditService 返回审计日志服务实例。
func NewISCSIAuditService() *ISCSIAuditService {
	return &ISCSIAuditService{db: database.DB}
}

// GetAuditLogs 预留查询审计日志的实现。
func (s *ISCSIAuditService) GetAuditLogs(page, limit int, action, userID, from, to string) ([]dto.ISCSIAuditLogResponse, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

// GetAuditStats 预留统计审计信息的实现。
func (s *ISCSIAuditService) GetAuditStats(from, to string) (*dto.ISCSIAuditStatsResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
