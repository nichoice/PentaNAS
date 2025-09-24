# PNAS 文件审计系统文档

## 概述

PNAS 文件审计系统提供企业级的文件操作监控和安全分析能力，通过API操作审计和文件系统实时监控双重机制，全面记录和分析文件相关的所有操作。

## 核心特性

### 🔍 **双重审计监控**
- **API操作审计**: 自动拦截和记录所有通过API的文件操作
- **文件系统监控**: 基于fsnotify的实时文件系统变化监控
- **全覆盖监控**: 确保任何文件操作都被记录，无遗漏

### ⚡ **高性能处理**
- **异步处理**: 批量写入，不影响系统响应速度
- **事件缓冲**: 10000条记录队列缓冲，防止数据丢失
- **批量优化**: 默认100条记录批量处理，5秒自动刷新

### 📊 **智能分析**
- **多维度查询**: 用户、时间、操作类型、文件路径等多维度过滤
- **统计分析**: 操作类型分布、用户活动分析、时间趋势分析
- **热度分析**: 文件访问频率热度图，识别热点文件
- **异常检测**: 自动识别暴力破解、批量删除、异常下载等风险行为

### 🛡️ **安全防护**
- **风险评分**: 基于异常活动计算系统风险等级
- **实时告警**: 自动生成安全威胁警报
- **行为追踪**: 完整的用户操作时间线追踪

## 系统架构

### 数据模型

#### 1. 文件审计日志 (FileAuditLog)
```go
type FileAuditLog struct {
    ID        string    `gorm:"primaryKey"`              // 唯一标识
    UserID    string    `gorm:"index"`                   // 用户ID
    Username  string    `gorm:"index"`                   // 用户名
    Operation string    `gorm:"index;not null"`          // 操作类型
    FilePath  string    `gorm:"not null"`                // 文件路径
    FileName  string    `gorm:"index"`                   // 文件名
    FileSize  int64                                      // 文件大小
    ClientIP  string                                     // 客户端IP
    UserAgent string                                     // 用户代理
    Status    string    `gorm:"index;not null"`          // 操作状态
    ErrorMsg  string                                     // 错误信息
    Duration  int64                                      // 操作耗时(ms)
    Hash      string                                     // 文件哈希
    Source    string    `gorm:"index;not null"`          // 来源(api/filesystem)
    Metadata  string                                     // 元数据
    CreatedAt time.Time                                  // 创建时间
}
```

#### 2. 文件访问统计 (FileAccessStats)
```go
type FileAccessStats struct {
    ID          string    `gorm:"primaryKey"`
    FilePath    string    `gorm:"uniqueIndex;not null"`  // 文件路径
    FileName    string    `gorm:"index"`                 // 文件名
    AccessCount int64     `gorm:"default:0"`            // 访问次数
    UserCount   int64     `gorm:"default:0"`            // 访问用户数
    LastAccess  time.Time                               // 最后访问时间
    FirstAccess time.Time                               // 首次访问时间
}
```

#### 3. 审计汇总 (AuditSummary)
```go
type AuditSummary struct {
    ID           string    `gorm:"primaryKey"`
    Date         string    `gorm:"uniqueIndex;not null"` // 日期
    UserID       string    `gorm:"index"`                // 用户ID
    Username     string    `gorm:"index"`                // 用户名
    Operation    string    `gorm:"index"`                // 操作类型
    TotalCount   int64     `gorm:"default:0"`           // 总操作数
    SuccessCount int64     `gorm:"default:0"`           // 成功操作数
    FailedCount  int64     `gorm:"default:0"`           // 失败操作数
    TotalSize    int64     `gorm:"default:0"`           // 总文件大小
}
```

### 核心组件

#### 1. 审计服务 (AuditService)
```go
type AuditService struct {
    eventChan     chan *AuditEvent    // 事件队列
    batchSize     int                 // 批量大小
    flushInterval time.Duration       // 刷新间隔
    db            *gorm.DB           // 数据库连接
}
```

**主要功能**:
- 异步事件处理
- 批量数据写入
- 定期统计汇总
- 文件访问统计更新

#### 2. 文件系统监控 (FilesystemMonitor)
```go
type FilesystemMonitor struct {
    watcher       *fsnotify.Watcher   // 文件系统监控器
    watchedDirs   map[string]bool     // 监控目录
    excludeRules  []string            // 排除规则
    recursive     bool                // 递归监控
}
```

**主要功能**:
- 实时文件系统事件监控
- 递归目录监控
- 智能文件过滤
- 重命名操作识别

#### 3. 审计中间件 (AuditMiddleware)
```go
func AuditMiddleware() gin.HandlerFunc {
    // API操作拦截和记录
}
```

**主要功能**:
- HTTP请求拦截
- 文件操作识别
- 请求响应分析
- 审计事件生成

## 操作类型

系统支持以下操作类型的审计：

| 操作类型 | 说明 | 来源 |
|---------|------|------|
| `create` | 文件创建 | API/Filesystem |
| `read` | 文件读取 | API |
| `update` | 文件更新 | API/Filesystem |
| `delete` | 文件删除 | API/Filesystem |
| `upload` | 文件上传 | API |
| `download` | 文件下载 | API |
| `rename` | 文件重命名 | API/Filesystem |
| `move` | 文件移动 | API/Filesystem |
| `copy` | 文件复制 | API |
| `compress` | 文件压缩 | API |
| `extract` | 文件解压 | API |

## 配置说明

### 审计配置 (config.yaml)
```yaml
audit:
  # 是否启用审计功能
  enabled: true

  # 监控的目录列表
  watch_paths:
    - "/home"
    - "/var/data"
    - "/opt/shared"

  # 排除的文件模式
  exclude_patterns:
    - ".DS_Store"
    - ".git"
    - "*.log"
    - "*.tmp"
    - "*.swp"

  # 性能参数
  batch_size: 100        # 批量大小
  flush_interval: 5      # 刷新间隔(秒)

  # 功能开关
  recursive_watch: true  # 递归监控
  enable_api: true       # API审计
  enable_filesystem: true # 文件系统监控

  # 数据保留
  retention_days: 90     # 日志保留天数
```

### 配置参数说明

| 参数 | 类型 | 默认值 | 说明 |
|------|------|-------|------|
| `enabled` | bool | true | 是否启用审计系统 |
| `watch_paths` | []string | ["/home", "/var/data", "/opt/shared"] | 监控目录列表 |
| `exclude_patterns` | []string | 见示例 | 排除文件模式 |
| `batch_size` | int | 100 | 批量处理大小 |
| `flush_interval` | int | 5 | 刷新间隔(秒) |
| `recursive_watch` | bool | true | 是否递归监控子目录 |
| `enable_api` | bool | true | 是否启用API审计 |
| `enable_filesystem` | bool | true | 是否启用文件系统监控 |
| `retention_days` | int | 90 | 日志保留天数 |

## API 接口

### 1. 获取审计日志
```http
GET /api/v1/audit/logs
```

**查询参数**:
- `user_id`: 用户ID过滤
- `username`: 用户名模糊匹配
- `operation`: 操作类型过滤
- `file_path`: 文件路径模糊匹配
- `status`: 操作状态 (success/failed)
- `source`: 来源 (api/filesystem)
- `start_time`: 开始时间 (RFC3339格式)
- `end_time`: 结束时间 (RFC3339格式)
- `page`: 页码，默认1
- `page_size`: 每页大小，默认20
- `order_by`: 排序字段，默认created_at
- `order_dir`: 排序方向，默认desc

### 2. 获取审计统计
```http
GET /api/v1/audit/stats
```

**功能**:
- 操作类型统计
- 用户活动统计
- 时间分布统计
- 热门文件统计
- 最近活动记录

### 3. 获取文件访问热度图
```http
GET /api/v1/audit/heatmap
```

**功能**:
- 文件访问频率排名
- 访问用户数统计
- 热度等级分类
- 文件大小信息

### 4. 异常行为检测
```http
GET /api/v1/audit/anomalies
```

**检测类型**:
- 频繁失败操作 (可能的暴力破解)
- 异常大量操作 (可能的批量操作)
- 异常下载活动 (可能的数据泄露)
- 批量删除活动 (可能的恶意删除)

### 5. 用户活动时间线
```http
GET /api/v1/audit/users/{user_id}/timeline
```

**功能**:
- 用户操作历史
- 活动统计信息
- 时间范围分析

## 安全分析

### 风险评分算法
```go
// 风险评分计算
riskScore := 0.0
for _, activity := range suspiciousActivities {
    switch activity.RiskLevel {
    case "high":   riskScore += 30
    case "medium": riskScore += 15
    case "low":    riskScore += 5
    }
}
for _, alert := range alerts {
    switch alert.Severity {
    case "high":   riskScore += 40
    case "medium": riskScore += 20
    case "low":    riskScore += 10
    }
}
```

### 异常检测规则

#### 1. 暴力破解检测
- 条件: 24小时内失败操作超过50次
- 风险等级: High
- 告警类型: brute_force

#### 2. 数据泄露检测
- 条件: 24小时内下载操作超过200次
- 风险等级: Medium
- 告警类型: data_exfiltration

#### 3. 批量删除检测
- 条件: 24小时内删除操作超过50次
- 风险等级: High
- 告警类型: mass_deletion

#### 4. 异常活动检测
- 条件: 用户单一操作类型超过100次
- 风险等级: Medium
- 描述: 异常大量的操作，可能存在批量操作

## 性能优化

### 1. 异步处理
- 事件队列缓冲10000条记录
- 批量写入减少数据库IO
- 非阻塞事件记录

### 2. 数据库优化
- 关键字段建立索引
- 批量插入优化
- 定期数据归档

### 3. 内存管理
- 重命名操作内存清理
- 定期清理过期数据
- 连接池管理

## 部署和维护

### 启动流程
1. 初始化审计服务
2. 启动文件系统监控
3. 配置监控目录
4. 启动API审计中间件

### 日常维护
- 定期清理过期审计日志
- 监控系统性能指标
- 分析安全风险报告
- 调整监控配置

### 故障排查
- 检查监控目录权限
- 验证数据库连接
- 查看服务日志
- 监控队列状态

## 最佳实践

### 1. 监控目录配置
- 选择关键业务目录进行监控
- 避免监控系统临时目录
- 合理配置排除规则

### 2. 性能调优
- 根据系统负载调整批量大小
- 合理设置刷新间隔
- 定期清理历史数据

### 3. 安全策略
- 定期查看异常行为报告
- 关注高风险用户活动
- 及时响应安全告警

### 4. 数据治理
- 设置合理的数据保留期
- 定期备份审计数据
- 建立数据访问权限控制

PNAS文件审计系统为您的NAS环境提供了企业级的文件安全监控和风险分析能力，确保数据安全和合规要求。