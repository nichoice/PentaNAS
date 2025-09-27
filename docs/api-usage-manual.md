# PNAS API 接口使用手册

## 概述

本手册详细介绍了PNAS系统所有API接口的使用方法，包括请求参数、响应格式和使用示例。所有接口均基于RESTful设计理念，采用JSON格式进行数据交换。

## 基础信息

- **Base URL**: `http://localhost:8080/api/v1`
- **认证方式**: JWT Bearer Token
- **Content-Type**: `application/json`
- **Swagger文档**: `http://localhost:8080/swagger/index.html`

## 认证说明

### 获取访问令牌

除了登录接口外，所有API都需要在请求头中包含JWT Token：

```http
Authorization: Bearer <your_jwt_token>
```

---

## 1. 认证相关接口

### 1.1 用户登录

**接口**: `POST /login`

**功能**: 用户登录获取JWT令牌

**请求参数**:
```json
{
  "username": "admin",
  "password": "admin123456"
}
```

**响应示例**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user_id": "uuid-string",
  "username": "admin",
  "expires_at": "2024-01-01T12:00:00Z"
}
```

**使用示例**:
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123456"
  }'
```

---

## 2. 用户管理接口

### 2.1 创建用户

**接口**: `POST /users`

**功能**: 创建新用户

**请求参数**:
```json
{
  "username": "newuser",
  "password": "password123",
  "role": "normal_user",
  "is_active": true,
  "remark": "新用户备注"
}
```

**参数说明**:
- `username`: 用户名 (3-20字符，字母数字)
- `password`: 密码 (最少8字符)
- `role`: 用户角色 (`super_admin`, `normal_user`, `audit_user`, `ops_user`)
- `is_active`: 是否激活
- `remark`: 备注信息

**响应示例**:
```json
{
  "id": "uuid-string",
  "username": "newuser",
  "is_active": true,
  "remark": "新用户备注",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### 2.2 获取用户列表

**接口**: `GET /users`

**功能**: 获取用户列表，支持分页和筛选

**查询参数**:
- `username`: 用户名过滤
- `role`: 角色过滤
- `status`: 状态过滤 (`active`/`inactive`)
- `page`: 页码 (默认1)
- `page_size`: 每页大小 (默认20)

**使用示例**:
```bash
curl -X GET "http://localhost:8080/api/v1/users?page=1&page_size=10&status=active" \
  -H "Authorization: Bearer <token>"
```

**响应示例**:
```json
{
  "users": [
    {
      "id": "uuid-string",
      "username": "admin",
      "is_active": true,
      "remark": "管理员",
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z",
      "last_login": "2024-01-01T12:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 10,
  "total_page": 1
}
```

### 2.3 获取单个用户信息

**接口**: `GET /users/{id}`

**功能**: 根据用户ID获取详细信息

**使用示例**:
```bash
curl -X GET http://localhost:8080/api/v1/users/uuid-string \
  -H "Authorization: Bearer <token>"
```

### 2.4 更新用户信息

**接口**: `PUT /users/{id}`

**功能**: 更新用户信息

**请求参数**:
```json
{
  "role": "ops_user",
  "is_active": false,
  "remark": "更新的备注",
  "password": "newpassword123"
}
```

### 2.5 删除用户

**接口**: `DELETE /users/{id}`

**功能**: 删除指定用户

### 2.6 更新用户状态

**接口**: `PUT /users/{id}/status`

**功能**: 更新用户激活状态

**请求参数**:
```json
{
  "is_active": false
}
```

---

## 3. 角色管理接口

### 3.1 分配角色

**接口**: `POST /auth/assign`

**功能**: 为用户分配角色

**请求参数**:
```json
{
  "user_id": "uuid-string",
  "role_id": "normal_user"
}
```

### 3.2 撤销角色

**接口**: `POST /auth/revoke`

**功能**: 撤销用户角色

**请求参数**:
```json
{
  "user_id": "uuid-string",
  "role_id": "normal_user"
}
```

### 3.3 获取用户角色

**接口**: `GET /auth/users/{user_id}/roles`

**功能**: 获取用户拥有的所有角色

### 3.4 获取角色用户

**接口**: `GET /auth/roles/{role_id}/users`

**功能**: 获取拥有指定角色的所有用户

### 3.5 获取角色列表

**接口**: `GET /roles`

**功能**: 获取系统所有角色

### 3.6 获取角色详情

**接口**: `GET /roles/{id}`

**功能**: 获取指定角色的详细信息

---

## 4. 存储管理接口

### 4.1 获取磁盘信息

**接口**: `GET /storage/disks`

**功能**: 获取系统所有磁盘信息

**响应示例**:
```json
{
  "disks": [
    {
      "name": "/dev/sda",
      "size": "1TB",
      "used": "500GB",
      "available": "500GB",
      "mount_point": "/",
      "file_system": "ext4"
    }
  ]
}
```

### 4.2 获取卷组信息

**接口**: `GET /storage/vgs`

**功能**: 获取LVM卷组信息

### 4.3 获取逻辑卷信息

**接口**: `GET /storage/lvs`

**功能**: 获取LVM逻辑卷信息

### 4.4 创建卷组

**接口**: `POST /storage/create_vg`

**功能**: 创建新的LVM卷组

**请求参数**:
```json
{
  "vg_name": "my_volume_group",
  "devices": ["/dev/sdb", "/dev/sdc"]
}
```

---

## 5. 系统监控接口

### 5.1 获取系统信息

**接口**: `GET /monitoring/system`

**功能**: 获取系统基本信息

**响应示例**:
```json
{
  "hostname": "pnas-server",
  "os": "Linux",
  "kernel": "5.4.0-42-generic",
  "uptime": "7 days, 12:34:56",
  "cpu_cores": 8,
  "total_memory": "16GB"
}
```

### 5.2 获取CPU信息

**接口**: `GET /monitoring/cpu`

**功能**: 获取CPU使用情况

**响应示例**:
```json
{
  "usage_percent": 45.6,
  "cores": 8,
  "model": "Intel(R) Core(TM) i7-8700K",
  "frequency": "3.70GHz"
}
```

### 5.3 获取内存信息

**接口**: `GET /monitoring/memory`

**功能**: 获取内存使用情况

**响应示例**:
```json
{
  "total": "16GB",
  "used": "8GB",
  "free": "8GB",
  "usage_percent": 50.0,
  "swap_total": "2GB",
  "swap_used": "0GB"
}
```

### 5.4 获取磁盘信息

**接口**: `GET /monitoring/disk`

**功能**: 获取磁盘使用情况

### 5.5 获取网络信息

**接口**: `GET /monitoring/network`

**功能**: 获取网络接口信息

### 5.6 获取存储协议信息

**接口**: `GET /monitoring/storage-protocols`

**功能**: 获取存储协议状态

### 5.7 获取完整监控数据

**接口**: `GET /monitoring/complete`

**功能**: 获取所有监控数据的汇总

### 5.8 获取服务状态

**接口**: `GET /monitoring/service/status`

**功能**: 获取系统服务运行状态

### 5.9 获取指标数据

**接口**: `GET /monitoring/metrics`

**功能**: 获取Prometheus格式的指标数据

---

## 6. WebSocket接口

### 6.1 建立WebSocket连接

**接口**: `GET /monitoring/websocket`

**功能**: 建立WebSocket连接进行实时数据推送

**使用示例**:
```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/monitoring/websocket');

ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log('实时监控数据:', data);
};
```

### 6.2 获取WebSocket信息

**接口**: `GET /monitoring/websocket/info`

**功能**: 获取WebSocket连接信息

### 6.3 广播消息

**接口**: `POST /monitoring/websocket/broadcast`

**功能**: 向所有WebSocket客户端广播消息

### 6.4 发送到指定客户端

**接口**: `POST /monitoring/websocket/client/{client_id}`

**功能**: 向指定客户端发送消息

---

## 7. 审计管理接口

### 7.1 获取审计日志

**接口**: `GET /audit/logs`

**功能**: 获取系统审计日志

**查询参数**:
- `user_id`: 用户ID过滤
- `username`: 用户名过滤
- `operation`: 操作类型过滤
- `file_path`: 文件路径过滤
- `status`: 状态过滤
- `source`: 来源过滤
- `start_time`: 开始时间
- `end_time`: 结束时间
- `page`: 页码
- `page_size`: 每页大小
- `order_by`: 排序字段
- `order_dir`: 排序方向

**使用示例**:
```bash
curl -X GET "http://localhost:8080/api/v1/audit/logs?operation=create&page=1&page_size=20" \
  -H "Authorization: Bearer <token>"
```

### 7.2 获取审计统计

**接口**: `GET /audit/stats`

**功能**: 获取审计统计数据

**响应示例**:
```json
{
  "total_count": 1000,
  "success_count": 950,
  "failed_count": 50,
  "operation_stats": [
    {
      "operation": "create",
      "count": 300
    }
  ],
  "user_stats": [
    {
      "user_id": "uuid",
      "username": "admin",
      "count": 500
    }
  ]
}
```

### 7.3 获取文件访问热力图

**接口**: `GET /audit/heatmap`

**功能**: 获取文件访问热力图数据

### 7.4 异常检测

**接口**: `GET /audit/anomalies`

**功能**: 检测异常行为和安全风险

**响应示例**:
```json
{
  "risk_score": 75.5,
  "suspicious_activities": [
    {
      "type": "unusual_access_pattern",
      "description": "异常访问模式",
      "severity": "medium"
    }
  ],
  "alerts": [
    {
      "type": "failed_login_attempts",
      "message": "连续登录失败",
      "count": 5
    }
  ]
}
```

### 7.5 获取用户活动时间线

**接口**: `GET /audit/users/{user_id}/timeline`

**功能**: 获取指定用户的活动时间线

---

## 8. iSCSI管理接口

### 8.1 获取iSCSI ACL列表

**接口**: `GET /iscsi/acls`

**功能**: 获取iSCSI访问控制列表

**查询参数**:
- `page`: 页码 (默认1)
- `page_size`: 每页大小 (默认10)
- `target_id`: Target ID过滤
- `permission`: 权限过滤
- `auth_type`: 认证类型过滤
- `is_enabled`: 是否启用过滤

### 8.2 创建iSCSI ACL

**接口**: `POST /iscsi/acls`

**功能**: 创建新的iSCSI访问控制规则

### 8.3 获取iSCSI ACL详情

**接口**: `GET /iscsi/acls/{id}`

**功能**: 获取指定ACL的详细信息

### 8.4 删除iSCSI ACL

**接口**: `DELETE /iscsi/acls/{id}`

**功能**: 删除指定的iSCSI ACL

---

## 9. 错误码说明

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 201 | 创建成功 |
| 204 | 删除成功 |
| 400 | 请求参数错误 |
| 401 | 未授权，需要登录 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 10. 通用响应格式

### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

### 错误响应

```json
{
  "error": "error message",
  "details": "detailed error information"
}
```

### 分页响应

```json
{
  "items": [ ... ],
  "total": 100,
  "page": 1,
  "page_size": 20,
  "total_page": 5
}
```

## 11. 使用建议

### 11.1 认证最佳实践

1. **Token管理**: 及时刷新即将过期的Token
2. **安全存储**: 不要在客户端明文存储Token
3. **权限检查**: 确保用户有足够权限访问相应接口

### 11.2 性能优化

1. **分页查询**: 大数据量查询时使用分页参数
2. **字段过滤**: 只请求需要的字段
3. **缓存机制**: 合理使用缓存减少重复请求

### 11.3 错误处理

1. **重试机制**: 对于网络错误实现合理的重试策略
2. **降级处理**: 服务不可用时的降级方案
3. **日志记录**: 记录API调用日志便于问题排查

## 12. 常见问题

### Q: 如何处理Token过期？
A: Token过期时会返回401状态码，需要重新登录获取新Token。

### Q: 接口调用频率有限制吗？
A: 存储相关接口有速率限制（5次/10秒），其他接口暂无限制。

### Q: 如何获取实时监控数据？
A: 使用WebSocket接口建立长连接，系统会自动推送实时数据。

### Q: 支持批量操作吗？
A: 目前大部分接口只支持单条记录操作，批量操作需要循环调用。

---

## 示例代码

### Python示例

```python
import requests
import json

# 登录获取Token
login_data = {
    "username": "admin",
    "password": "admin123456"
}

response = requests.post(
    "http://localhost:8080/api/v1/login",
    json=login_data
)

token = response.json()["token"]

# 使用Token调用其他接口
headers = {
    "Authorization": f"Bearer {token}",
    "Content-Type": "application/json"
}

# 获取用户列表
users_response = requests.get(
    "http://localhost:8080/api/v1/users",
    headers=headers
)

print(users_response.json())
```

### JavaScript示例

```javascript
// 登录获取Token
async function login() {
  const response = await fetch('http://localhost:8080/api/v1/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      username: 'admin',
      password: 'admin123456'
    })
  });

  const data = await response.json();
  return data.token;
}

// 获取系统信息
async function getSystemInfo(token) {
  const response = await fetch('http://localhost:8080/api/v1/monitoring/system', {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });

  return await response.json();
}
```

---

*本API手册持续更新，如有疑问请查看最新的Swagger文档或联系开发团队。*