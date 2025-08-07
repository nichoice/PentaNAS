# PNAS 项目开发记录

## 项目概述

PNAS (Personal Network Access System) 是一个基于 Go 语言开发的个人网络访问系统，采用标准的 Go 项目布局，实现了完整的用户管理和认证授权体系。

## 技术栈

- **后端框架**: Gin Web Framework
- **数据库**: SQLite3 (开发/测试) / PostgreSQL (生产)
- **ORM**: GORM
- **认证**: JWT (JSON Web Token)
- **密码加密**: bcrypt
- **日志**: Zap 结构化日志
- **文档**: Swagger/OpenAPI
- **测试**: testify 测试框架

## 项目结构

```
pnas/
├── cmd/                    # 应用程序入口
│   └── main.go            # 主程序文件
├── internal/              # 私有应用代码
│   ├── config/           # 配置管理
│   │   ├── database.go   # 数据库配置
│   │   ├── jwt.go        # JWT配置
│   │   └── logger.go     # 日志配置
│   ├── controllers/      # 控制器层
│   │   ├── auth_controller.go      # 认证控制器
│   │   ├── health_controller.go    # 健康检查控制器
│   │   ├── user_controller.go      # 用户控制器
│   │   └── user_group_controller.go # 用户组控制器
│   ├── database/         # 数据库相关
│   │   └── migrate.go    # 数据库迁移和种子数据
│   ├── middlewares/      # 中间件
│   │   ├── auth.go       # JWT认证中间件
│   │   └── logger.go     # 日志中间件
│   ├── models/           # 数据模型
│   │   ├── base.go       # 基础模型
│   │   ├── health_check.go # 健康检查模型
│   │   ├── user.go       # 用户模型
│   │   └── user_group.go # 用户组模型
│   ├── repositories/     # 仓库层
│   │   ├── interfaces.go           # 仓库接口定义
│   │   ├── health_check_repository.go # 健康检查仓库
│   │   ├── user_repository.go      # 用户仓库
│   │   └── user_group_repository.go # 用户组仓库
│   ├── routes/           # 路由配置
│   │   └── routes.go     # 路由定义
│   └── services/         # 服务层
│       ├── auth_service.go # 认证服务
│       └── jwt_service.go  # JWT服务
├── api/                   # API定义
│   └── v1/               # API v1版本
│       ├── auth.go       # 认证API定义
│       ├── health.go     # 健康检查API定义
│       └── user.go       # 用户API定义
├── config/               # 配置文件
│   ├── database.yml      # 数据库配置
│   ├── jwt.yml          # JWT配置
│   ├── logger.yml       # 日志配置
│   ├── logger-dev.yml   # 开发环境日志配置
│   └── logger-prod.yml  # 生产环境日志配置
├── tests/                # 测试文件
│   ├── benchmark_test.go    # 性能测试
│   ├── controllers_test.go  # 控制器测试
│   ├── integration_test.go  # 集成测试
│   ├── models_test.go      # 模型测试
│   ├── repositories_test.go # 仓库测试
│   ├── setup_test.go       # 测试设置
│   ├── run_tests.sh        # 测试运行脚本
│   └── testdata/           # 测试数据
│       └── test_config.yml # 测试配置
├── logs/                 # 日志目录
├── docs/                 # Swagger生成的文档
├── go.mod               # Go模块定义
├── go.sum               # Go模块依赖
├── README.md            # 项目说明
└── DEVELOP.md           # 开发记录 (本文件)
```

## 开发历程

### 第一阶段：项目初始化和结构重构

**时间**: 2025-08-07

**主要改动**:
1. **项目结构重构**: 将单一的 `main.go` 文件按照标准 Go 项目布局进行分离
2. **创建核心目录结构**: 建立 `cmd/`, `internal/`, `api/`, `config/`, `tests/` 等标准目录
3. **分离关注点**: 将代码按照 MVC 架构分离到不同层次

**涉及文件**:
- `cmd/main.go` - 应用程序入口
- `internal/controllers/` - 控制器层
- `internal/routes/routes.go` - 路由配置
- `internal/models/` - 数据模型层

### 第二阶段：Swagger文档集成

**时间**: 2025-08-07

**主要改动**:
1. **集成Swagger**: 添加 API 文档自动生成功能
2. **API注释**: 为所有接口添加 Swagger 注释
3. **文档路由**: 配置 `/swagger/*` 路由访问文档

**涉及文件**:
- `api/v1/health.go` - 健康检查API定义
- `internal/controllers/health_controller.go` - 添加Swagger注释
- `internal/routes/routes.go` - 添加Swagger路由

**依赖包**:
```go
github.com/swaggo/swag/cmd/swag
github.com/swaggo/gin-swagger
github.com/swaggo/files
```

### 第三阶段：Zap日志系统集成

**时间**: 2025-08-07

**主要改动**:
1. **Zap日志库**: 集成结构化日志系统
2. **多环境配置**: 支持开发和生产环境不同的日志配置
3. **彩色日志**: 开发环境支持彩色日志输出
4. **日志中间件**: 添加HTTP请求日志记录

**涉及文件**:
- `config/logger.yml`, `config/logger-dev.yml`, `config/logger-prod.yml` - 日志配置
- `internal/config/logger.go` - 日志配置结构
- `internal/middlewares/logger.go` - 日志中间件
- `logs/.gitignore` - 日志目录配置

**依赖包**:
```go
go.uber.org/zap
gopkg.in/yaml.v3
```

### 第四阶段：数据库支持

**时间**: 2025-08-07

**主要改动**:
1. **GORM集成**: 添加ORM支持
2. **多数据库支持**: SQLite3(开发/测试) + PostgreSQL(生产)
3. **自动迁移**: 数据库表结构自动创建和更新
4. **仓库模式**: 实现Repository Pattern进行数据访问抽象

**涉及文件**:
- `config/database.yml` - 数据库配置
- `internal/config/database.go` - 数据库配置和连接
- `internal/models/base.go` - 基础模型定义
- `internal/repositories/` - 仓库层实现
- `internal/database/migrate.go` - 数据库迁移

**依赖包**:
```go
gorm.io/gorm
gorm.io/driver/sqlite
gorm.io/driver/postgres
```

### 第五阶段：用户管理系统

**时间**: 2025-08-07

**主要改动**:
1. **用户模型优化**: 简化用户字段（用户名、密码、状态、用户类型）
2. **用户组功能**: 添加用户组管理，支持用户与用户组关联
3. **三员管理**: 实现系统管理员、安全管理员、审计管理员的分离

**涉及文件**:
- `internal/models/user.go` - 用户模型
- `internal/models/user_group.go` - 用户组模型
- `internal/repositories/user_repository.go` - 用户仓库
- `internal/repositories/user_group_repository.go` - 用户组仓库

**用户类型定义**:
- `UserTypeSystemAdmin = 1` - 系统管理员
- `UserTypeSecurityAdmin = 2` - 安全管理员
- `UserTypeAuditAdmin = 3` - 审计管理员
- `UserTypeNormal = 4` - 普通用户

### 第六阶段：完整API接口

**时间**: 2025-08-07

**主要改动**:
1. **用户CRUD接口**: 完整的用户增删改查功能
2. **用户组CRUD接口**: 完整的用户组管理功能
3. **关联查询**: 支持用户组及其用户列表查询
4. **Swagger文档更新**: 为所有接口添加完整的API文档

**涉及文件**:
- `api/v1/user.go` - 用户API定义
- `internal/controllers/user_controller.go` - 用户控制器
- `internal/controllers/user_group_controller.go` - 用户组控制器

**API接口列表**:
- `POST /api/v1/users` - 创建用户
- `GET /api/v1/users` - 获取用户列表
- `GET /api/v1/users/:id` - 获取用户详情
- `PUT /api/v1/users/:id` - 更新用户信息
- `DELETE /api/v1/users/:id` - 删除用户
- `POST /api/v1/user-groups` - 创建用户组
- `GET /api/v1/user-groups` - 获取用户组列表
- `GET /api/v1/user-groups/:id` - 获取用户组详情
- `GET /api/v1/user-groups/:id/users` - 获取用户组及其用户列表
- `PUT /api/v1/user-groups/:id` - 更新用户组信息
- `DELETE /api/v1/user-groups/:id` - 删除用户组

### 第七阶段：单元测试体系

**时间**: 2025-08-07

**主要改动**:
1. **完整测试覆盖**: 模型、仓库、控制器、集成测试
2. **性能测试**: 基准测试和性能分析
3. **测试数据**: 独立的测试数据库配置
4. **测试脚本**: 自动化测试运行脚本

**涉及文件**:
- `tests/models_test.go` - 模型测试
- `tests/repositories_test.go` - 仓库测试
- `tests/controllers_test.go` - 控制器测试
- `tests/integration_test.go` - 集成测试
- `tests/benchmark_test.go` - 性能测试
- `tests/setup_test.go` - 测试设置
- `tests/run_tests.sh` - 测试运行脚本
- `tests/testdata/test_config.yml` - 测试配置

**依赖包**:
```go
github.com/stretchr/testify
```

### 第八阶段：JWT认证系统

**时间**: 2025-08-08

**主要改动**:
1. **JWT认证**: 实现完整的JWT登录认证系统
2. **密码加密**: 使用bcrypt对密码进行安全哈希
3. **认证中间件**: JWT Token验证和用户信息提取
4. **访问控制**: 普通用户登录限制，接口权限保护
5. **种子数据**: 自动创建默认管理员账户

**涉及文件**:
- `config/jwt.yml` - JWT配置文件
- `internal/config/jwt.go` - JWT配置结构
- `internal/services/jwt_service.go` - JWT服务
- `internal/services/auth_service.go` - 认证服务
- `api/v1/auth.go` - 认证API定义
- `internal/controllers/auth_controller.go` - 认证控制器
- `internal/middlewares/auth.go` - JWT认证中间件
- `internal/database/migrate.go` - 种子数据初始化

**依赖包**:
```go
github.com/golang-jwt/jwt/v5
golang.org/x/crypto/bcrypt
```

**认证功能**:
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/refresh` - Token刷新
- `POST /api/v1/auth/logout` - 用户登出

**默认管理员账户**:
- 系统管理员: `sysadmin` / `admin123`
- 安全管理员: `secadmin` / `admin123`
- 审计管理员: `auditadmin` / `admin123`

**安全特性**:
- 普通用户不能登录系统
- 除 `/ping` 和 `/api/v1/health/ping` 外，所有接口都需要JWT认证
- 密码使用bcrypt哈希存储
- JWT Token有过期时间限制
- 完整的认证和授权日志记录

## 配置说明

### 数据库配置 (config/database.yml)
```yaml
development:
  driver: sqlite
  dsn: "./data/pnas_dev.db"
  
production:
  driver: postgres
  dsn: "host=localhost user=pnas password=pnas dbname=pnas port=5432 sslmode=disable"
```

### JWT配置 (config/jwt.yml)
```yaml
secret_key: "your-secret-key-here"
expires_in: 24h
refresh_expires_in: 168h
issuer: "pnas"
```

### 日志配置 (config/logger-dev.yml)
```yaml
level: debug
encoding: console
output_paths:
  - stdout
  - logs/app.log
development: true
color: true
```

## 运行说明

### 开发环境启动
```bash
# 安装依赖
go mod tidy

# 生成Swagger文档
swag init -g cmd/main.go

# 启动服务
go run cmd/main.go
```

### 测试运行
```bash
# 运行所有测试
./tests/run_tests.sh

# 或者手动运行
go test ./tests/... -v
```

### 访问地址
- **API服务**: http://localhost:8080
- **Swagger文档**: http://localhost:8080/swagger/index.html
- **健康检查**: http://localhost:8080/ping

## API使用示例

### 登录获取Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"sysadmin","password":"admin123"}'
```

### 使用Token访问受保护接口
```bash
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 项目特色

1. **标准Go项目布局**: 遵循Go社区最佳实践
2. **完整的认证授权**: JWT + bcrypt安全认证
3. **三员管理**: 系统、安全、审计管理员分离
4. **多环境支持**: 开发、测试、生产环境配置
5. **完整测试覆盖**: 单元测试、集成测试、性能测试
6. **结构化日志**: Zap日志库，支持彩色输出
7. **API文档**: Swagger自动生成文档
8. **仓库模式**: 数据访问层抽象
9. **中间件支持**: 认证、日志、恢复中间件
10. **数据库迁移**: 自动表结构创建和种子数据

## 后续规划

1. **权限系统**: 基于角色的访问控制(RBAC)
2. **审计日志**: 操作审计和日志分析
3. **API限流**: 接口访问频率限制
4. **缓存系统**: Redis缓存集成
5. **配置中心**: 动态配置管理
6. **监控告警**: 系统监控和告警机制
7. **容器化**: Docker和Kubernetes支持
8. **CI/CD**: 自动化构建和部署

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。