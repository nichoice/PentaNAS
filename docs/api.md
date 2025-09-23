# PNAS API 文档

## API 概览

PNAS 提供完整的 RESTful API 接口和 WebSocket 实时通信功能，支持用户管理、存储管理、系统监控等核心功能。

## 基础信息

- **Base URL**: `http://localhost:8080`
- **API 版本**: v1
- **Content-Type**: `application/json`
- **Swagger 文档**: `http://localhost:8080/swagger/index.html`

## 认证机制

### JWT Token 认证

所有需要认证的 API 都需要在请求头中包含 JWT Token：

```http
Authorization: Bearer <jwt_token>
```

### 获取 Token

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123456"
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      "role": {
        "id": 1,
        "name": "超级管理员"
      }
    }
  }
}
```

## API 接口分类

### 1. 用户认证接口

#### 1.1 用户登录
```http
POST /api/v1/auth/login
```

**请求参数**:
```json
{
  "username": "string",     // 用户名，必填
  "password": "string"      // 密码，必填
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "登录成功",
  "data": {
    "token": "jwt_token_string",
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      "role": {
        "id": 1,
        "name": "超级管理员",
        "description": "拥有系统所有权限"
      }
    }
  }
}
```

#### 1.2 用户注册
```http
POST /api/v1/auth/register
```

**请求参数**:
```json
{
  "username": "string",     // 用户名，必填，3-32字符
  "password": "string",     // 密码，必填，6-32字符
  "email": "string",        // 邮箱，可选
  "role_id": 2              // 角色ID，可选，默认为普通用户
}
```

#### 1.3 用户登出
```http
POST /api/v1/auth/logout
Authorization: Bearer <token>
```

### 2. 存储管理接口

#### 2.1 获取磁盘列表
```http
GET /api/v1/storage/disks
Authorization: Bearer <token>
```

**响应示例**:
```json
{
  "code": 0,
  "message": "获取成功",
  "data": [
    {
      "name": "/dev/sda",
      "size": "1TB",
      "type": "SSD",
      "model": "Samsung SSD 970",
      "serial": "S4ELNX0M123456",
      "mount_point": "/",
      "file_system": "ext4",
      "used_space": "500GB",
      "available_space": "500GB",
      "usage_percentage": 50.0
    }
  ]
}
```

#### 2.2 获取卷组列表
```http
GET /api/v1/storage/vg
Authorization: Bearer <token>
```

**响应示例**:
```json
{
  "code": 0,
  "message": "获取成功",
  "data": [
    {
      "name": "vg_data",
      "size": "2TB",
      "free_size": "1TB",
      "pv_count": 2,
      "lv_count": 3,
      "physical_volumes": ["/dev/sdb", "/dev/sdc"]
    }
  ]
}
```

#### 2.3 获取逻辑卷列表
```http
GET /api/v1/storage/lv
Authorization: Bearer <token>
```

#### 2.4 创建卷组
```http
POST /api/v1/storage/vg
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "new_vg",              // 卷组名称，必填
  "physical_volumes": [          // 物理卷列表，必填
    "/dev/sdb",
    "/dev/sdc"
  ]
}
```

### 3. 系统监控接口

#### 3.1 获取系统信息
```http
GET /api/v1/monitoring/system-info
Authorization: Bearer <token>
```

**响应示例**:
```json
{
  "code": 0,
  "message": "获取成功",
  "data": {
    "hostname": "pnas-server",
    "os": "linux",
    "platform": "ubuntu",
    "platform_version": "20.04",
    "kernel_version": "5.4.0-74-generic",
    "architecture": "x86_64",
    "cpu_count": 8,
    "memory_total": 16777216,
    "uptime": 86400,
    "boot_time": "2023-12-01T10:00:00Z"
  }
}
```

#### 3.2 获取 CPU 信息
```http
GET /api/v1/monitoring/cpu
Authorization: Bearer <token>
```

**响应示例**:
```json
{
  "code": 0,
  "message": "获取成功",
  "data": {
    "cpu_count": 8,
    "cpu_usage": [
      {
        "cpu": "cpu0",
        "user": 25.5,
        "system": 10.2,
        "idle": 64.3,
        "iowait": 0.0,
        "irq": 0.0,
        "softirq": 0.0
      }
    ],
    "load_average": {
      "load1": 0.5,
      "load5": 0.3,
      "load15": 0.2
    }
  }
}
```

#### 3.3 获取内存信息
```http
GET /api/v1/monitoring/memory
Authorization: Bearer <token>
```

#### 3.4 获取磁盘信息
```http
GET /api/v1/monitoring/disk
Authorization: Bearer <token>
```

#### 3.5 获取网络信息
```http
GET /api/v1/monitoring/network
Authorization: Bearer <token>
```

#### 3.6 获取存储协议信息
```http
GET /api/v1/monitoring/storage-protocols
Authorization: Bearer <token>
```

**响应示例**:
```json
{
  "code": 0,
  "message": "获取成功",
  "data": {
    "smb": {
      "enabled": true,
      "version": "4.13.2",
      "active_connections": 5,
      "shares": [
        {
          "name": "data",
          "path": "/mnt/data",
          "users": ["user1", "user2"]
        }
      ]
    },
    "nfs": {
      "enabled": true,
      "version": "4.2",
      "exports": [
        {
          "path": "/mnt/nfs_share",
          "clients": ["192.168.1.0/24"],
          "options": "rw,sync,no_subtree_check"
        }
      ]
    },
    "iscsi": {
      "enabled": false,
      "targets": []
    }
  }
}
```

#### 3.7 获取完整监控数据
```http
GET /api/v1/monitoring/complete
Authorization: Bearer <token>
```

#### 3.8 Prometheus 指标
```http
GET /api/v1/monitoring/metrics
```

**响应格式**: Prometheus 格式的指标数据

#### 3.9 获取 WebSocket 客户端信息
```http
GET /api/v1/monitoring/clients
Authorization: Bearer <token>
```

**响应示例**:
```json
{
  "code": 0,
  "message": "获取成功",
  "data": {
    "total_clients": 3,
    "clients": [
      {
        "id": "uuid-string",
        "remote_addr": "192.168.1.100:54321",
        "user_agent": "Mozilla/5.0...",
        "connected_at": "2023-12-01T15:30:00Z",
        "last_ping": "2023-12-01T15:35:00Z"
      }
    ]
  }
}
```

## WebSocket 实时通信

### 连接地址
```
ws://localhost:8080/ws
```

### 连接流程

1. **建立连接**
   ```javascript
   const ws = new WebSocket('ws://localhost:8080/ws');
   ```

2. **接收欢迎消息**
   ```json
   {
     "type": "welcome",
     "data": {
       "client_id": "uuid-string",
       "message": "Connected to monitoring service"
     },
     "timestamp": "2023-12-01T15:30:00Z",
     "client_id": "uuid-string"
   }
   ```

### 消息类型

#### 1. 客户端发送消息

##### 1.1 心跳检测
```json
{
  "type": "ping",
  "data": {},
  "timestamp": "2023-12-01T15:30:00Z"
}
```

##### 1.2 订阅数据
```json
{
  "type": "subscribe",
  "data": {
    "subscription": "cpu"  // 可选值: system_info, cpu, memory, disk, network, storage_protocols, all
  },
  "timestamp": "2023-12-01T15:30:00Z"
}
```

##### 1.3 请求数据
```json
{
  "type": "request_data",
  "data": {
    "type": "system_info"  // 可选值同订阅类型
  },
  "timestamp": "2023-12-01T15:30:00Z"
}
```

#### 2. 服务器推送消息

##### 2.1 心跳响应
```json
{
  "type": "pong",
  "data": {
    "timestamp": "2023-12-01T15:30:00Z"
  },
  "timestamp": "2023-12-01T15:30:00Z",
  "client_id": "uuid-string"
}
```

##### 2.2 监控数据推送
```json
{
  "type": "monitoring_data",
  "data": {
    "system_info": { /* 系统信息 */ },
    "cpu_info": { /* CPU 信息 */ },
    "memory_info": { /* 内存信息 */ },
    "disk_info": [ /* 磁盘信息 */ ],
    "network_info": [ /* 网络信息 */ ],
    "storage_protocols": { /* 存储协议信息 */ }
  },
  "timestamp": "2023-12-01T15:30:00Z"
}
```

##### 2.3 特定数据推送
```json
{
  "type": "cpu",
  "data": {
    "cpu_count": 8,
    "cpu_usage": [
      /* CPU 使用率数据 */
    ],
    "load_average": {
      "load1": 0.5,
      "load5": 0.3,
      "load15": 0.2
    }
  },
  "timestamp": "2023-12-01T15:30:00Z",
  "client_id": "uuid-string"
}
```

##### 2.4 数据响应
```json
{
  "type": "data_response",
  "data": {
    "type": "system_info",
    "data": {
      /* 请求的数据内容 */
    }
  },
  "timestamp": "2023-12-01T15:30:00Z",
  "client_id": "uuid-string"
}
```

##### 2.5 错误消息
```json
{
  "type": "error",
  "data": {
    "error": "Unknown subscription type"
  },
  "timestamp": "2023-12-01T15:30:00Z",
  "client_id": "uuid-string"
}
```

### 订阅数据推送频率

| 数据类型 | 推送间隔 | 说明 |
|---------|---------|------|
| system_info | 10秒 | 系统基础信息 |
| cpu | 2秒 | CPU 使用率 |
| memory | 3秒 | 内存使用情况 |
| disk | 5秒 | 磁盘使用情况和 I/O |
| network | 3秒 | 网络接口和 I/O |
| storage_protocols | 10秒 | 存储协议状态 |
| all | 5秒 | 完整监控数据 |

## 错误码说明

| 错误码 | 说明 | HTTP状态码 |
|--------|------|-----------|
| 0 | 成功 | 200 |
| 4001 | 参数错误 | 400 |
| 4011 | 未认证 | 401 |
| 4031 | 权限不足 | 403 |
| 4041 | 资源不存在 | 404 |
| 4291 | 请求过于频繁 | 429 |
| 5001 | 服务器内部错误 | 500 |

## 使用示例

### JavaScript/TypeScript 示例

#### HTTP API 调用
```javascript
// 登录获取 token
async function login(username, password) {
  const response = await fetch('/api/v1/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ username, password }),
  });

  const result = await response.json();
  if (result.code === 0) {
    localStorage.setItem('token', result.data.token);
    return result.data;
  }
  throw new Error(result.message);
}

// 获取系统信息
async function getSystemInfo() {
  const token = localStorage.getItem('token');
  const response = await fetch('/api/v1/monitoring/system-info', {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });

  const result = await response.json();
  return result.data;
}
```

#### WebSocket 实时监控
```javascript
class PNASMonitor {
  constructor() {
    this.ws = null;
    this.clientId = null;
  }

  connect() {
    this.ws = new WebSocket('ws://localhost:8080/ws');

    this.ws.onopen = () => {
      console.log('WebSocket connected');
    };

    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      this.handleMessage(message);
    };

    this.ws.onclose = () => {
      console.log('WebSocket disconnected');
      // 重连逻辑
      setTimeout(() => this.connect(), 5000);
    };
  }

  handleMessage(message) {
    switch (message.type) {
      case 'welcome':
        this.clientId = message.client_id;
        this.subscribeTo('all');
        break;
      case 'monitoring_data':
        this.updateDashboard(message.data);
        break;
      case 'cpu':
        this.updateCPUChart(message.data);
        break;
      // 处理其他消息类型
    }
  }

  subscribeTo(type) {
    this.send({
      type: 'subscribe',
      data: { subscription: type },
      timestamp: new Date().toISOString()
    });
  }

  send(message) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    }
  }
}

// 使用示例
const monitor = new PNASMonitor();
monitor.connect();
```

## Postman 集合

可以导入以下 Postman 集合来快速测试 API：

```json
{
  "info": {
    "name": "PNAS API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "auth": {
    "type": "bearer",
    "bearer": [
      {
        "key": "token",
        "value": "{{jwt_token}}",
        "type": "string"
      }
    ]
  },
  "variable": [
    {
      "key": "base_url",
      "value": "http://localhost:8080"
    },
    {
      "key": "jwt_token",
      "value": ""
    }
  ]
}
```

## 开发调试

### 启用 Swagger UI
访问 `http://localhost:8080/swagger/index.html` 可以使用交互式 API 文档进行测试。

### 生成 API 文档
```bash
# 安装 swag
go install github.com/swaggo/swag/cmd/swag@latest

# 生成文档
swag init -g cmd/main.go -o cmd/docs
```

### WebSocket 调试工具
推荐使用以下工具调试 WebSocket 连接：
- WebSocket King (Chrome 扩展)
- wscat (命令行工具)
- Postman (支持 WebSocket)

这份 API 文档涵盖了 PNAS 系统的所有 HTTP 接口和 WebSocket 通信协议，为前端开发和第三方集成提供了完整的技术参考。