# PNAS 系统架构设计

## 架构概览

PNAS 采用分层架构设计，基于 DDD (领域驱动设计) 原则，确保代码的可维护性、可扩展性和业务逻辑的清晰分离。

## 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                      Client Layer                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  Web UI     │  │  Mobile App │  │  Third Party Apps   │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                               │
                         HTTP/WebSocket
                               │
┌─────────────────────────────────────────────────────────────┐
│                    API Gateway Layer                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   Router    │  │ Middleware  │  │    WebSocket Hub    │  │
│  │   (Gin)     │  │  (CORS/     │  │                     │  │
│  │             │  │   Auth/     │  │                     │  │
│  │             │  │   Rate      │  │                     │  │
│  │             │  │   Limit)    │  │                     │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                               │
┌─────────────────────────────────────────────────────────────┐
│                   Controller Layer                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   User      │  │  Storage    │  │    Monitoring       │  │
│  │ Controller  │  │ Controller  │  │    Controller       │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                               │
┌─────────────────────────────────────────────────────────────┐
│                   Service Layer                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │    User     │  │   Storage   │  │    Monitoring       │  │
│  │   Service   │  │   Service   │  │     Service         │  │
│  │             │  │             │  │                     │  │
│  │  ┌───────┐  │  │  ┌───────┐  │  │  ┌───────────────┐  │  │
│  │  │ Auth  │  │  │  │  LVM  │  │  │  │  System Info  │  │  │
│  │  │Service│  │  │  │Service│  │  │  │   Collection  │  │  │
│  │  └───────┘  │  │  └───────┘  │  │  └───────────────┘  │  │
│  │             │  │             │  │                     │  │
│  │  ┌───────┐  │  │  ┌───────┐  │  │  ┌───────────────┐  │  │
│  │  │ Init  │  │  │  │Protocol│  │  │  │   WebSocket   │  │  │
│  │  │Service│  │  │  │Service│  │  │  │    Service    │  │  │
│  │  └───────┘  │  │  └───────┘  │  │  └───────────────┘  │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                               │
┌─────────────────────────────────────────────────────────────┐
│                   Domain Layer                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │    Models   │  │     DTOs    │  │      Utils          │  │
│  │             │  │             │  │                     │  │
│  │  ┌───────┐  │  │  ┌───────┐  │  │  ┌───────────────┐  │  │
│  │  │ User  │  │  │  │WebSkt │  │  │  │     Crypto    │  │  │
│  │  │ Role  │  │  │  │ Msg   │  │  │  │     Helper    │  │  │
│  │  │Storage│  │  │  │Monitor│  │  │  │               │  │  │
│  │  │       │  │  │  │ Data  │  │  │  │               │  │  │
│  │  └───────┘  │  │  └───────┘  │  │  └───────────────┘  │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                               │
┌─────────────────────────────────────────────────────────────┐
│                Infrastructure Layer                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  Database   │  │   System    │  │    External APIs    │  │
│  │  (SQLite)   │  │   (Linux)   │  │                     │  │
│  │             │  │             │  │  ┌───────────────┐  │  │
│  │   GORM      │  │  LVM/VG/LV  │  │  │  Prometheus   │  │  │
│  │             │  │             │  │  │    Metrics    │  │  │
│  │             │  │  SMB/NFS/   │  │  └───────────────┘  │  │
│  │             │  │   iSCSI     │  │                     │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## 核心组件设计

### 1. API Gateway Layer (API 网关层)

#### Router (路由层)
- **技术**: Gin HTTP 框架
- **职责**:
  - HTTP 请求路由分发
  - RESTful API 设计规范
  - 路径参数解析和验证

#### Middleware (中间件层)
- **CORS 中间件**: 跨域资源共享配置
- **Authentication 中间件**: JWT Token 验证
- **Rate Limiting 中间件**: API 请求频率限制
- **Logging 中间件**: 请求日志记录

#### WebSocket Hub
- **技术**: gorilla/websocket
- **功能**:
  - 客户端连接管理
  - 实时数据广播
  - 订阅模式支持

### 2. Controller Layer (控制器层)

#### User Controller
- 用户注册、登录、登出
- 用户信息管理
- 权限验证

#### Storage Controller
- 磁盘扫描和管理
- LVM 卷组和逻辑卷操作
- 存储协议配置

#### Monitoring Controller
- 系统性能数据采集
- 实时监控数据推送
- Prometheus 指标暴露

### 3. Service Layer (服务层)

#### User Service
```go
type UserService struct {
    userRepo UserRepository
    authService *AuthService
}
```
- **AuthService**: JWT Token 生成和验证
- **InitService**: 系统初始化，创建默认管理员

#### Storage Service
```go
type StorageService struct {
    // LVM 管理相关方法
}
```
- **LVM 操作**: 卷组、逻辑卷的创建、删除、扩容
- **协议服务**: SMB、NFS、iSCSI 配置管理

#### Monitoring Service
```go
type MonitoringService struct {
    // 系统监控相关方法
}
```
- **系统信息采集**: CPU、内存、磁盘、网络
- **实时数据处理**: 定时采集和数据聚合
- **存储协议监控**: SMB、NFS、iSCSI 状态监控

#### WebSocket Service
```go
type WebSocketHub struct {
    clients    map[string]*WebSocketClient
    broadcast  chan WebSocketMessage
    register   chan *WebSocketClient
    unregister chan *WebSocketClient
}
```
- **连接管理**: 客户端注册、注销、心跳检测
- **消息广播**: 支持全局广播和定向推送
- **订阅机制**: 按类型订阅特定监控数据

### 4. Domain Layer (领域层)

#### Models (数据模型)
```go
// 用户模型
type User struct {
    ID       uint   `gorm:"primaryKey"`
    Username string `gorm:"unique;not null"`
    Password string `gorm:"not null"`
    Email    string `gorm:"unique"`
    RoleID   uint
    Role     Role
}

// 角色模型
type Role struct {
    ID          uint   `gorm:"primaryKey"`
    Name        string `gorm:"unique;not null"`
    Description string
    Users       []User
}
```

#### DTOs (数据传输对象)
```go
// WebSocket 消息
type WebSocketMessage struct {
    Type      string      `json:"type"`
    Data      interface{} `json:"data"`
    Timestamp time.Time   `json:"timestamp"`
    ClientID  string      `json:"client_id,omitempty"`
}

// 监控数据
type MonitoringData struct {
    SystemInfo         *SystemInfo         `json:"system_info"`
    CPUInfo           *CPUInfo            `json:"cpu_info"`
    MemoryInfo        *MemoryInfo         `json:"memory_info"`
    DiskInfo          []DiskInfo          `json:"disk_info"`
    NetworkInfo       []NetworkInfo       `json:"network_info"`
    StorageProtocols  *StorageProtocolInfo `json:"storage_protocols"`
}
```

### 5. Infrastructure Layer (基础设施层)

#### Database (数据库层)
- **SQLite**: 轻量级关系数据库
- **GORM**: ORM 框架，自动迁移和关系映射
- **连接池**: 数据库连接管理

#### System Integration (系统集成)
- **LVM 命令**: 通过系统调用操作 Linux LVM
- **系统信息**: 使用 gopsutil 库采集系统数据
- **存储协议**: 集成 Samba、NFS、Open-iSCSI

## 数据流架构

### 1. HTTP API 请求流
```
Client Request → Router → Middleware → Controller → Service → Repository → Database
```

### 2. WebSocket 实时数据流
```
System Data → Monitoring Service → WebSocket Hub → Connected Clients
```

### 3. 初始化流程
```
Application Start → Init Service → Check Users → Create Default Admin/Roles → Start Services
```

## 安全架构

### 1. 认证与授权
- **JWT Token**: 无状态身份验证
- **RBAC**: 基于角色的访问控制
- **密码加密**: bcrypt 算法

### 2. API 安全
- **CORS 配置**: 跨域请求控制
- **Rate Limiting**: 防止 API 滥用
- **输入验证**: 参数合法性检查

### 3. 系统安全
- **权限隔离**: 最小权限原则
- **安全日志**: 操作审计追踪

## 性能设计

### 1. 并发处理
- **Goroutine**: 异步任务处理
- **Channel**: 安全的并发通信
- **Context**: 请求生命周期管理

### 2. 缓存策略
- **内存缓存**: 热点数据缓存
- **连接复用**: 数据库连接池

### 3. 监控优化
- **数据采集间隔**: 按重要性配置不同采集频率
- **WebSocket 优化**: 连接管理和消息队列

## 扩展性设计

### 1. 模块化设计
- **服务解耦**: 各服务模块独立
- **接口抽象**: 便于功能扩展

### 2. 配置化
- **环境配置**: 支持多环境部署
- **功能开关**: 可配置的功能启用/禁用

### 3. 插件化
- **存储协议**: 可扩展的协议支持
- **监控指标**: 可定制的监控维度

## 部署架构

### 1. 单体部署
```
┌─────────────────────────────────────┐
│            Linux Server             │
│  ┌─────────────────────────────────┐ │
│  │         PNAS Binary             │ │
│  │  ┌─────┐ ┌─────┐ ┌───────────┐  │ │
│  │  │ API │ │ WS  │ │ Monitoring│  │ │
│  │  └─────┘ └─────┘ └───────────┘  │ │
│  └─────────────────────────────────┘ │
│  ┌─────────────────────────────────┐ │
│  │          SQLite DB              │ │
│  └─────────────────────────────────┘ │
│  ┌─────────────────────────────────┐ │
│  │      Storage Services           │ │
│  │    (SMB/NFS/iSCSI)             │ │
│  └─────────────────────────────────┘ │
└─────────────────────────────────────┘
```

### 2. 容器化部署
```
┌─────────────────────────────────────┐
│            Docker Host              │
│  ┌─────────────────────────────────┐ │
│  │         PNAS Container          │ │
│  │  ┌─────────────────────────────┐ │ │
│  │  │        Application          │ │ │
│  │  └─────────────────────────────┘ │ │
│  │  ┌─────────────────────────────┐ │ │
│  │  │        Volume Mounts        │ │ │
│  │  │    /data → /host/data       │ │ │
│  │  │    /logs → /host/logs       │ │ │
│  │  └─────────────────────────────┘ │ │
│  └─────────────────────────────────┘ │
└─────────────────────────────────────┘
```

此架构设计确保了系统的高内聚、低耦合，支持水平扩展和功能迭代，为 PNAS 的长期发展提供了坚实的技术基础。