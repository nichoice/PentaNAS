# PNAS - Personal Network Attached Storage

## 项目概述

PNAS (Personal Network Attached Storage) 是一个基于 Go 语言开发的个人网络存储系统，采用现代化的微服务架构设计，提供完整的存储管理、系统监控和性能分析功能。

## 核心特性

### 🚀 存储管理
- **LVM 存储管理**: 支持磁盘、卷组(VG)、逻辑卷(LV)的完整生命周期管理
- **存储协议支持**: 集成 SMB、NFS、iSCSI 等主流存储协议
- **存储池管理**: 灵活的存储资源分配和管理

### 📊 系统监控
- **实时性能监控**: CPU、内存、磁盘、网络等系统资源实时监控
- **系统信息采集**: 系统版本、启动时间、硬件信息等详细数据
- **WebSocket 实时推送**: 支持实时数据流推送和订阅
- **Prometheus 集成**: 兼容 Prometheus 生态系统，支持 PromQL 查询

### 🔐 用户与权限
- **RBAC 权限模型**: 基于角色的访问控制系统
- **JWT 认证**: 安全的 Token 认证机制
- **默认账号初始化**: 首次启动自动创建管理员账号

### 🛡️ 文件审计系统
- **双重审计监控**: API操作 + 文件系统实时监控
- **异步日志处理**: 高性能批量写入，不影响系统响应
- **智能分析**: 多维度查询、统计分析、异常检测
- **安全防护**: 实时风险评分、安全告警、行为追踪

### 🌐 API 与接口
- **RESTful API**: 完整的 REST API 接口设计
- **Swagger 文档**: 自动生成的 API 文档
- **WebSocket 支持**: 双向实时通信能力

## 技术栈

### 后端框架
- **Go 1.21+**: 主要开发语言
- **Gin**: HTTP Web 框架
- **GORM**: ORM 数据库操作库
- **SQLite**: 轻量级数据库

### 监控与审计
- **gopsutil**: 系统信息采集库
- **gorilla/websocket**: WebSocket 通信库
- **Prometheus**: 监控指标收集
- **fsnotify**: 文件系统事件监控

### 开发工具
- **Swag**: Swagger 文档自动生成
- **UUID**: 唯一标识符生成
- **bcrypt**: 密码加密

## 项目结构

```
pnas/
├── cmd/                    # 应用程序入口
│   ├── main.go            # 主程序入口
│   ├── server.go          # 服务器启动配置
│   └── docs/              # Swagger 生成文档
├── internal/              # 内部应用代码
│   ├── app/               # 应用层
│   │   ├── dto/           # 数据传输对象
│   │   └── models/        # 数据模型
│   ├── controllers/       # 控制器层
│   │   ├── auth_controller.go    # 用户认证控制器
│   │   ├── storage_controller.go # 存储管理控制器
│   │   ├── monitoring_controller.go # 系统监控控制器
│   │   └── audit_controller.go   # 审计管理控制器
│   ├── middleware/        # 中间件
│   │   ├── auth.go               # 认证中间件
│   │   ├── logging.go            # 日志中间件
│   │   └── audit.go              # 审计中间件
│   ├── models/            # 数据模型
│   │   ├── user.go               # 用户模型
│   │   ├── role.go               # 角色模型
│   │   └── audit.go              # 审计模型
│   ├── routes/            # 路由配置
│   ├── services/          # 服务层
│   │   ├── init_service.go       # 系统初始化服务
│   │   ├── monitoring_service.go # 监控服务
│   │   ├── audit_service.go      # 审计服务
│   │   └── filesystem_monitor.go # 文件系统监控服务
│   └── utils/             # 工具函数
├── docs/                  # 项目文档
├── logs/                  # 日志文件
├── uploads/               # 上传文件存储
└── web/                   # 前端静态资源
```

## 快速开始

### 环境要求
- Go 1.21 或更高版本
- Linux 操作系统（推荐 Ubuntu 20.04+）
- Root 权限（用于 LVM 和系统监控功能）

### 安装依赖
```bash
go mod download
```

### 编译项目
```bash
go build -o pnas cmd/main.go
```

### 运行服务
```bash
sudo ./pnas
```

### 访问服务
- API 服务: http://localhost:8080
- Swagger 文档: http://localhost:8080/swagger/index.html
- WebSocket 连接: ws://localhost:8080/ws

## 默认账号

系统首次启动时会自动创建默认管理员账号：
- 用户名: `admin`
- 密码: `admin123456`

**重要**: 请在首次登录后立即修改默认密码！

## 文档导航

- [架构设计](./architecture.md) - 详细的系统架构设计文档
- [API 文档](./api.md) - API 接口使用说明
- [审计系统](./audit.md) - 文件审计系统详细说明
- [部署指南](./deployment.md) - 生产环境部署配置
- [监控指南](./monitoring.md) - 系统监控功能使用说明

## 开发指南

### 代码规范
- 遵循 Go 官方代码规范
- 使用 DDD (领域驱动设计) 架构模式
- 每个功能模块包含完整的单元测试

### 提交规范
- feat: 新功能
- fix: 错误修复
- docs: 文档更新
- refactor: 代码重构
- test: 测试相关

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](../LICENSE) 文件了解详情。

## 贡献

欢迎提交 Pull Request 和 Issue！

## 联系方式

如有问题或建议，请通过 GitHub Issues 联系我们。