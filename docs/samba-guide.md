# PNAS Samba 功能指南

## 概述

PNAS 系统集成了完整的 Samba 文件共享服务，支持以下核心功能：

- **角色和账号管理**：多种用户角色，精细化权限控制
- **文件共享管理**：灵活的共享配置，支持匿名访问和权限控制
- **时间机器支持**：为 macOS 设备提供 Time Machine 备份支持
- **回收站功能**：误删文件保护，支持文件恢复
- **多通道支持**：提高网络传输性能
- **版本管理**：支持多个 Samba 版本
- **实时监控**：连接状态、审计日志、性能监控

## API 接口

### 基础 URL
```
http://your-server:port/api/v1/samba
```

### 账号管理

#### 创建 Samba 账号
```http
POST /accounts
Content-Type: application/json

{
  "user_id": "user-uuid",
  "samba_user": "username",
  "password": "password123",
  "role": "samba_user",
  "description": "普通用户账号"
}
```

**角色类型：**
- `samba_admin`: Samba管理员
- `samba_user`: 普通用户
- `samba_guest`: 访客用户
- `samba_backup`: 备份用户
- `samba_timemachine`: 时间机器用户

#### 获取账号列表
```http
GET /accounts?offset=0&limit=10&role=samba_user
```

#### 更新账号
```http
PUT /accounts/{id}
Content-Type: application/json

{
  "password": "newpassword123",
  "role": "samba_admin",
  "is_enabled": true
}
```

#### 删除账号
```http
DELETE /accounts/{id}
```

### 共享管理

#### 创建共享
```http
POST /shares
Content-Type: application/json

{
  "name": "shared_folder",
  "path": "/data/shared",
  "comment": "共享文件夹",
  "is_enabled": true,
  "allow_guest": false,
  "writable": true,
  "create_mask": "0664",
  "directory_mask": "0775",
  "enable_time_machine": false,
  "enable_recycle_bin": true,
  "enable_multi_channel": false
}
```

#### 获取共享列表
```http
GET /shares?offset=0&limit=10&enabled=true
```

#### 更新共享
```http
PUT /shares/{id}
Content-Type: application/json

{
  "comment": "更新的共享描述",
  "writable": false,
  "enable_recycle_bin": true
}
```

#### 设置共享访问权限
```http
POST /shares/access
Content-Type: application/json

{
  "share_id": "share-uuid",
  "account_id": "account-uuid",
  "permission": "read"
}
```

**权限类型：**
- `read`: 只读权限
- `write`: 读写权限
- `admin`: 管理权限

### 时间机器功能

#### 启用时间机器
```http
PUT /shares/{id}/timemachine/enable?quota=102400
```

#### 禁用时间机器
```http
PUT /shares/{id}/timemachine/disable
```

#### 获取时间机器状态
```http
GET /shares/{id}/timemachine/status
```

#### 获取时间机器共享列表
```http
GET /timemachine/shares
```

### 回收站功能

#### 获取回收站条目
```http
GET /recycle/items?share_id=share-uuid&offset=0&limit=10
```

#### 恢复文件
```http
PUT /recycle/items/restore
Content-Type: application/json

{
  "item_id": "item-uuid",
  "restore_path": "/data/shared/restored_file.txt"
}
```

#### 永久删除文件
```http
DELETE /recycle/items/{id}
```

#### 清空回收站
```http
DELETE /recycle/shares/{shareId}/empty
```

#### 获取回收站统计
```http
GET /recycle/statistics?share_id=share-uuid
```

### 配置管理

#### 获取全局配置
```http
GET /config/global
```

#### 更新全局配置
```http
PUT /config/global
Content-Type: application/json

{
  "server_string": "PNAS Samba Server",
  "workgroup": "WORKGROUP",
  "security_level": "user",
  "enable_multi_channel": true,
  "max_channels": 4,
  "enable_auditing": true
}
```

#### 生成配置文件
```http
GET /config/generate
```

#### 写入配置文件
```http
POST /config/write?path=/etc/samba/smb.conf
```

#### 重载配置
```http
POST /config/reload
```

### 监控功能

#### 获取 Samba 整体状态
```http
GET /monitoring/status
```

#### 获取服务状态
```http
GET /monitoring/services
```

#### 获取活跃连接
```http
GET /monitoring/connections
```

#### 获取连接历史
```http
GET /monitoring/connections/history?hours=24&offset=0&limit=10
```

#### 获取审计日志
```http
GET /monitoring/audit/logs?username=user&operation=read&offset=0&limit=10
```

### 多通道功能

#### 启用多通道
```http
PUT /shares/{id}/multichannel/enable
```

#### 禁用多通道
```http
PUT /shares/{id}/multichannel/disable
```

#### 获取多通道状态
```http
GET /monitoring/multichannel/status
```

## 使用示例

### 1. 创建基本的文件共享

```bash
# 1. 创建用户账号
curl -X POST http://localhost:8080/api/v1/samba/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-123",
    "samba_user": "testuser",
    "password": "password123",
    "role": "samba_user"
  }'

# 2. 创建共享
curl -X POST http://localhost:8080/api/v1/samba/shares \
  -H "Content-Type: application/json" \
  -d '{
    "name": "documents",
    "path": "/data/documents",
    "comment": "文档共享",
    "writable": true,
    "enable_recycle_bin": true
  }'

# 3. 设置访问权限
curl -X POST http://localhost:8080/api/v1/samba/shares/access \
  -H "Content-Type: application/json" \
  -d '{
    "share_id": "share-uuid",
    "account_id": "account-uuid",
    "permission": "write"
  }'
```

### 2. 配置时间机器备份

```bash
# 1. 创建时间机器专用用户
curl -X POST http://localhost:8080/api/v1/samba/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "tm-user-123",
    "samba_user": "timemachine",
    "password": "tm_password123",
    "role": "samba_timemachine"
  }'

# 2. 创建时间机器共享
curl -X POST http://localhost:8080/api/v1/samba/shares \
  -H "Content-Type: application/json" \
  -d '{
    "name": "TimeMachine",
    "path": "/data/timemachine",
    "comment": "Time Machine Backup",
    "writable": true,
    "enable_time_machine": true,
    "time_machine_quota": 1048576
  }'

# 3. 启用时间机器功能
curl -X PUT "http://localhost:8080/api/v1/samba/shares/share-uuid/timemachine/enable?quota=1048576"
```

### 3. 启用多通道支持

```bash
# 1. 更新全局配置启用多通道
curl -X PUT http://localhost:8080/api/v1/samba/config/global \
  -H "Content-Type: application/json" \
  -d '{
    "enable_multi_channel": true,
    "max_channels": 8
  }'

# 2. 为特定共享启用多通道
curl -X PUT http://localhost:8080/api/v1/samba/shares/share-uuid/multichannel/enable

# 3. 检查多通道状态
curl http://localhost:8080/api/v1/samba/monitoring/multichannel/status
```

## 注意事项

1. **系统要求**：需要安装 Samba 4.4+ 版本以支持多通道功能
2. **权限配置**：确保 Samba 进程有足够权限访问共享目录
3. **网络配置**：多通道功能需要多个网络接口支持
4. **存储空间**：时间机器和回收站功能会占用额外存储空间
5. **安全考虑**：定期更新账号密码，监控访问日志

## 故障排除

### 常见问题

1. **连接失败**
   - 检查服务状态：`GET /monitoring/services`
   - 验证网络配置
   - 确认账号状态

2. **权限错误**
   - 验证共享权限设置
   - 检查文件系统权限
   - 确认用户角色配置

3. **时间机器无法连接**
   - 确认 fruit 模块已加载
   - 验证时间机器配置
   - 检查磁盘空间

4. **性能问题**
   - 启用多通道支持
   - 调整缓存设置
   - 监控系统资源使用

### 日志查看

```bash
# 获取审计日志
curl "http://localhost:8080/api/v1/samba/monitoring/audit/logs?limit=100"

# 获取系统状态
curl http://localhost:8080/api/v1/samba/monitoring/status

# 获取连接状态
curl http://localhost:8080/api/v1/samba/monitoring/connections
```

## 更多信息

- [Samba官方文档](https://www.samba.org/samba/docs/)
- [PNAS项目地址](https://github.com/your-repo/pnas)
- [API完整文档](http://localhost:8080/swagger/)