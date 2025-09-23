# PNAS 监控系统使用指南

## 监控系统概述

PNAS 内置了完整的系统监控功能，可以实时采集和展示服务器的性能数据，包括 CPU、内存、磁盘、网络等核心指标，并支持存储协议状态监控。

## 监控功能特性

### 🔍 系统监控
- **CPU 监控**: 使用率、负载均衡、核心状态
- **内存监控**: 使用率、缓存、交换空间
- **磁盘监控**: 空间使用、I/O 性能、挂载点状态
- **网络监控**: 接口状态、流量统计、连接数

### 📊 存储协议监控
- **SMB/CIFS**: 连接数、共享状态、用户会话
- **NFS**: 导出状态、客户端连接、性能指标
- **iSCSI**: Target 状态、LUN 信息、会话管理

### 🚀 实时数据推送
- **WebSocket 连接**: 双向实时通信
- **订阅机制**: 按需订阅特定监控数据
- **自定义频率**: 不同数据类型的采集间隔可配置

### 📈 Prometheus 集成
- **标准指标**: 兼容 Prometheus 格式
- **自定义指标**: 支持业务相关指标
- **PromQL 查询**: 支持复杂的查询和聚合

## 使用方式

### 1. HTTP API 接口

#### 获取系统概览
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/monitoring/system-info
```

#### 获取实时 CPU 数据
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/monitoring/cpu
```

#### 获取完整监控数据
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/monitoring/complete
```

### 2. WebSocket 实时监控

#### JavaScript 客户端示例
```javascript
// 建立连接
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = function() {
    console.log('监控连接已建立');

    // 订阅所有监控数据
    ws.send(JSON.stringify({
        type: 'subscribe',
        data: { subscription: 'all' },
        timestamp: new Date().toISOString()
    }));
};

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);

    switch(data.type) {
        case 'monitoring_data':
            updateDashboard(data.data);
            break;
        case 'cpu':
            updateCPUChart(data.data);
            break;
        case 'memory':
            updateMemoryChart(data.data);
            break;
    }
};

function updateDashboard(monitoringData) {
    // 更新仪表板数据
    document.getElementById('cpu-usage').textContent =
        monitoringData.cpu_info.cpu_usage[0].user + '%';
    document.getElementById('memory-usage').textContent =
        monitoringData.memory_info.usage_percent.toFixed(1) + '%';
}
```

#### Python 客户端示例
```python
import asyncio
import websockets
import json

async def monitor_client():
    uri = "ws://localhost:8080/ws"

    async with websockets.connect(uri) as websocket:
        # 发送订阅请求
        subscribe_msg = {
            "type": "subscribe",
            "data": {"subscription": "cpu"},
            "timestamp": "2023-12-01T15:30:00Z"
        }
        await websocket.send(json.dumps(subscribe_msg))

        # 接收数据
        async for message in websocket:
            data = json.loads(message)
            if data['type'] == 'cpu':
                print(f"CPU 使用率: {data['data']['cpu_usage'][0]['user']:.1f}%")

# 运行客户端
asyncio.run(monitor_client())
```

### 3. 订阅类型说明

| 订阅类型 | 推送间隔 | 包含数据 |
|---------|---------|----------|
| `all` | 5秒 | 完整监控数据 |
| `system_info` | 10秒 | 系统基础信息 |
| `cpu` | 2秒 | CPU 使用率和负载 |
| `memory` | 3秒 | 内存使用情况 |
| `disk` | 5秒 | 磁盘使用和 I/O |
| `network` | 3秒 | 网络接口和流量 |
| `storage_protocols` | 10秒 | 存储协议状态 |

### 4. Prometheus 指标

#### 访问 Prometheus 端点
```bash
curl http://localhost:8080/api/v1/monitoring/metrics
```

#### 主要指标说明

**系统指标**
```prometheus
# CPU 使用率
pnas_cpu_usage_percent{cpu="cpu0",mode="user"} 25.5
pnas_cpu_usage_percent{cpu="cpu0",mode="system"} 10.2

# 内存使用
pnas_memory_usage_bytes{type="total"} 16777216000
pnas_memory_usage_bytes{type="used"} 8388608000
pnas_memory_usage_percent 50.0

# 磁盘使用
pnas_disk_usage_bytes{device="/dev/sda1",mountpoint="/"} 500000000000
pnas_disk_usage_percent{device="/dev/sda1",mountpoint="/"} 50.0

# 网络流量
pnas_network_bytes_sent{interface="eth0"} 1073741824
pnas_network_bytes_recv{interface="eth0"} 2147483648
```

**存储协议指标**
```prometheus
# SMB 连接数
pnas_smb_active_connections 5
pnas_smb_shares_count 3

# NFS 导出数
pnas_nfs_exports_count 2
pnas_nfs_client_connections 8

# iSCSI Target 状态
pnas_iscsi_targets_count 1
pnas_iscsi_active_sessions 2
```

#### PromQL 查询示例

**CPU 使用率趋势**
```promql
rate(pnas_cpu_usage_percent{mode="user"}[5m])
```

**内存使用率告警**
```promql
pnas_memory_usage_percent > 80
```

**磁盘空间不足检测**
```promql
pnas_disk_usage_percent > 90
```

**网络流量峰值**
```promql
max_over_time(rate(pnas_network_bytes_sent[1m])[10m:])
```

## 监控数据格式

### 系统信息数据结构
```json
{
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
```

### CPU 监控数据结构
```json
{
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
```

### 内存监控数据结构
```json
{
  "total": 16777216,
  "available": 8388608,
  "used": 8388608,
  "free": 4194304,
  "usage_percent": 50.0,
  "buffers": 1048576,
  "cached": 2097152,
  "swap_total": 2097152,
  "swap_used": 0,
  "swap_free": 2097152
}
```

### 磁盘监控数据结构
```json
[
  {
    "device": "/dev/sda1",
    "mountpoint": "/",
    "filesystem": "ext4",
    "total": 1000000000,
    "used": 500000000,
    "free": 500000000,
    "usage_percent": 50.0
  }
]
```

### 网络监控数据结构
```json
[
  {
    "interface": "eth0",
    "bytes_sent": 1073741824,
    "bytes_recv": 2147483648,
    "packets_sent": 1000000,
    "packets_recv": 1500000,
    "errors_in": 0,
    "errors_out": 0,
    "drop_in": 0,
    "drop_out": 0
  }
]
```

### 存储协议监控数据结构
```json
{
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
```

## 配置监控系统

### 监控配置参数

在 `/etc/pnas/config.yaml` 中配置监控参数：

```yaml
monitoring:
  # 是否启用监控
  enable: true

  # 数据采集间隔
  interval: 5s

  # WebSocket 配置
  websocket:
    enable: true
    ping_interval: 54s
    buffer_size: 256
    max_clients: 100

  # Prometheus 配置
  prometheus:
    enable: true
    path: "/metrics"

  # 存储协议监控
  storage_protocols:
    smb:
      enable: true
      check_interval: 10s
    nfs:
      enable: true
      check_interval: 10s
    iscsi:
      enable: false
      check_interval: 15s
```

### 自定义监控指标

可以通过环境变量或配置文件添加自定义监控指标：

```yaml
monitoring:
  custom_metrics:
    # 自定义磁盘监控路径
    additional_disk_paths:
      - "/mnt/storage1"
      - "/mnt/storage2"

    # 自定义网络接口
    monitored_interfaces:
      - "eth0"
      - "eth1"
      - "bond0"

    # 监控阈值
    thresholds:
      cpu_warning: 80
      cpu_critical: 95
      memory_warning: 80
      memory_critical: 95
      disk_warning: 85
      disk_critical: 95
```

## 告警和通知

### 基于阈值的告警

PNAS 支持基于预设阈值的自动告警：

```json
{
  "type": "alert",
  "data": {
    "level": "warning",
    "metric": "cpu_usage",
    "value": 85.5,
    "threshold": 80.0,
    "message": "CPU 使用率超过警告阈值",
    "timestamp": "2023-12-01T15:30:00Z",
    "hostname": "pnas-server"
  }
}
```

### WebSocket 告警推送

告警会自动推送给所有连接的 WebSocket 客户端：

```javascript
ws.onmessage = function(event) {
    const data = JSON.parse(event.data);

    if (data.type === 'alert') {
        showAlert(data.data);
    }
};

function showAlert(alert) {
    const alertDiv = document.createElement('div');
    alertDiv.className = `alert alert-${alert.level}`;
    alertDiv.textContent = alert.message;
    document.getElementById('alerts').appendChild(alertDiv);
}
```

## 性能优化建议

### 1. 监控数据采集优化

- **调整采集间隔**: 根据实际需求调整不同指标的采集频率
- **选择性监控**: 只启用需要的监控功能
- **批量处理**: 使用批量数据处理减少系统开销

### 2. WebSocket 连接优化

```javascript
// 连接重试机制
function connectWithRetry() {
    const ws = new WebSocket('ws://localhost:8080/ws');

    ws.onclose = function() {
        console.log('连接断开，5秒后重试...');
        setTimeout(connectWithRetry, 5000);
    };

    ws.onerror = function(error) {
        console.error('WebSocket 错误:', error);
    };
}

// 心跳保持
function sendHeartbeat(ws) {
    setInterval(() => {
        if (ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({
                type: 'ping',
                timestamp: new Date().toISOString()
            }));
        }
    }, 30000);
}
```

### 3. 数据缓存策略

- **客户端缓存**: 缓存历史数据减少重复请求
- **数据压缩**: 对大量数据进行压缩传输
- **增量更新**: 只传输变化的数据部分

## 故障排除

### 常见监控问题

#### 1. 监控数据不准确
```bash
# 检查系统权限
sudo ls -la /proc/
sudo ls -la /sys/

# 验证 gopsutil 功能
go run -c "
import 'github.com/shirou/gopsutil/v3/cpu'
info, _ := cpu.Info()
fmt.Printf('%+v', info)
"
```

#### 2. WebSocket 连接失败
```bash
# 检查端口监听
sudo netstat -tlnp | grep :8080

# 测试 WebSocket 连接
wscat -c ws://localhost:8080/ws

# 检查防火墙规则
sudo ufw status verbose
```

#### 3. Prometheus 指标缺失
```bash
# 验证 metrics 端点
curl http://localhost:8080/api/v1/monitoring/metrics

# 检查日志
sudo journalctl -u pnas -f | grep -i prometheus
```

### 调试工具

#### 监控状态检查脚本
```bash
#!/bin/bash
# monitor-check.sh

echo "=== PNAS 监控系统检查 ==="

# 检查服务状态
echo "1. 服务状态:"
systemctl status pnas | grep Active

# 检查端口监听
echo "2. 端口监听:"
ss -tlnp | grep :8080

# 检查 API 响应
echo "3. API 响应测试:"
curl -s http://localhost:8080/api/v1/monitoring/system-info | jq .

# 检查 WebSocket
echo "4. WebSocket 测试:"
timeout 5 wscat -c ws://localhost:8080/ws

# 检查 Prometheus 指标
echo "5. Prometheus 指标:"
curl -s http://localhost:8080/api/v1/monitoring/metrics | head -10

echo "=== 检查完成 ==="
```

这份监控系统使用指南提供了完整的监控功能使用方法，帮助用户充分利用 PNAS 的监控能力，实现对系统状态的全面掌控。