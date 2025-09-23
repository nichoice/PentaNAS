# PNAS 部署指南

## 系统要求

### 硬件要求

#### 最低配置
- **CPU**: 2核心
- **内存**: 4GB RAM
- **存储**: 20GB 可用空间
- **网络**: 100Mbps 网络连接

#### 推荐配置
- **CPU**: 4核心以上
- **内存**: 8GB RAM 以上
- **存储**: 100GB+ SSD 存储
- **网络**: 1Gbps 网络连接

### 软件要求

#### 操作系统
- **Linux**: Ubuntu 20.04+ / CentOS 8+ / Debian 11+
- **架构**: x86_64 / amd64

#### 系统依赖
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y lvm2 samba nfs-kernel-server open-iscsi

# CentOS/RHEL
sudo yum update
sudo yum install -y lvm2 samba nfs-utils iscsi-initiator-utils
```

#### 开发依赖（可选）
- **Go**: 1.21+ （仅开发环境需要）
- **Git**: 版本控制
- **Docker**: 容器化部署（可选）

## 安装方式

### 方式一：二进制安装（推荐）

#### 1. 下载预编译二进制文件
```bash
# 创建安装目录
sudo mkdir -p /opt/pnas
cd /opt/pnas

# 下载最新版本（替换为实际版本）
wget https://github.com/your-org/pnas/releases/download/v1.0.0/pnas-linux-amd64.tar.gz

# 解压
tar -xzf pnas-linux-amd64.tar.gz
```

#### 2. 配置权限
```bash
sudo chown root:root pnas
sudo chmod +x pnas
```

#### 3. 创建系统用户
```bash
sudo useradd -r -s /bin/false pnas
sudo mkdir -p /var/lib/pnas
sudo chown pnas:pnas /var/lib/pnas
```

### 方式二：源码编译

#### 1. 安装 Go 环境
```bash
# 下载 Go
wget https://golang.org/dl/go1.21.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.4.linux-amd64.tar.gz

# 配置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

#### 2. 克隆代码并编译
```bash
git clone https://github.com/your-org/pnas.git
cd pnas

# 安装依赖
go mod download

# 编译
go build -o pnas cmd/main.go

# 安装到系统目录
sudo mv pnas /opt/pnas/
```

### 方式三：Docker 部署

#### 1. 创建 Dockerfile
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o pnas cmd/main.go

FROM alpine:latest

# 安装系统依赖
RUN apk --no-cache add ca-certificates lvm2 samba nfs-utils

WORKDIR /root/

# 复制二进制文件
COPY --from=builder /app/pnas .

# 创建必要目录
RUN mkdir -p /var/lib/pnas /var/log/pnas

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["./pnas"]
```

#### 2. 构建和运行
```bash
# 构建镜像
docker build -t pnas:latest .

# 运行容器
docker run -d \
  --name pnas \
  --privileged \
  -p 8080:8080 \
  -v /dev:/dev \
  -v /var/lib/pnas:/var/lib/pnas \
  -v /var/log/pnas:/var/log/pnas \
  pnas:latest
```

## 配置文件

### 主配置文件 `/etc/pnas/config.yaml`

```yaml
# 服务器配置
server:
  host: "0.0.0.0"
  port: 8080
  debug: false
  read_timeout: 30s
  write_timeout: 30s

# 数据库配置
database:
  driver: "sqlite"
  dsn: "/var/lib/pnas/pnas.db"
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600s

# JWT 配置
jwt:
  secret: "your-secret-key-change-this"
  expire_hours: 24

# 日志配置
logging:
  level: "info"
  file: "/var/log/pnas/pnas.log"
  max_size: 100    # MB
  max_backups: 10
  max_age: 30      # days
  compress: true

# 监控配置
monitoring:
  enable: true
  interval: 5s
  websocket:
    enable: true
    ping_interval: 54s
    buffer_size: 256
  prometheus:
    enable: true
    path: "/metrics"

# 存储配置
storage:
  protocols:
    smb:
      enable: true
      config_path: "/etc/samba/smb.conf"
    nfs:
      enable: true
      exports_path: "/etc/exports"
    iscsi:
      enable: false
      config_path: "/etc/iscsi/iscsid.conf"

# 安全配置
security:
  rate_limit:
    enable: true
    requests_per_minute: 60
  cors:
    enable: true
    allowed_origins: ["*"]
    allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allowed_headers: ["Authorization", "Content-Type"]
```

### 环境变量配置

```bash
# /etc/environment 或 ~/.bashrc
export PNAS_CONFIG=/etc/pnas/config.yaml
export PNAS_LOG_LEVEL=info
export PNAS_JWT_SECRET=your-secret-key
export PNAS_DB_PATH=/var/lib/pnas/pnas.db
```

## 系统服务配置

### SystemD 服务文件 `/etc/systemd/system/pnas.service`

```ini
[Unit]
Description=PNAS - Personal Network Attached Storage
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=/opt/pnas
ExecStart=/opt/pnas/pnas
ExecReload=/bin/kill -HUP $MAINPID
KillMode=process
Restart=on-failure
RestartSec=5s

# 环境变量
Environment=PNAS_CONFIG=/etc/pnas/config.yaml
Environment=PNAS_LOG_LEVEL=info

# 安全设置
NoNewPrivileges=false
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/pnas /var/log/pnas /tmp

# 资源限制
LimitNOFILE=65536
LimitNPROC=32768

[Install]
WantedBy=multi-user.target
```

### 启动和管理服务

```bash
# 重新加载 systemd 配置
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start pnas

# 开机自启
sudo systemctl enable pnas

# 查看服务状态
sudo systemctl status pnas

# 查看日志
sudo journalctl -u pnas -f

# 重启服务
sudo systemctl restart pnas

# 停止服务
sudo systemctl stop pnas
```

## 数据库初始化

### 首次启动检查

PNAS 首次启动时会自动：
1. 创建数据库文件
2. 执行数据库迁移
3. 创建默认角色和管理员账号

### 手动数据库操作

```bash
# 备份数据库
sudo cp /var/lib/pnas/pnas.db /var/lib/pnas/pnas.db.backup.$(date +%Y%m%d_%H%M%S)

# 重置数据库（谨慎操作）
sudo rm /var/lib/pnas/pnas.db
sudo systemctl restart pnas
```

## 网络配置

### 防火墙设置

#### UFW (Ubuntu)
```bash
# 开放 PNAS 端口
sudo ufw allow 8080/tcp

# 开放存储协议端口
sudo ufw allow 445/tcp   # SMB
sudo ufw allow 2049/tcp  # NFS
sudo ufw allow 3260/tcp  # iSCSI

# 重新加载防火墙
sudo ufw reload
```

#### FirewallD (CentOS/RHEL)
```bash
# 开放端口
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --permanent --add-port=445/tcp
sudo firewall-cmd --permanent --add-port=2049/tcp
sudo firewall-cmd --permanent --add-port=3260/tcp

# 重新加载
sudo firewall-cmd --reload
```

### 反向代理配置

#### Nginx 配置示例
```nginx
server {
    listen 80;
    server_name pnas.example.com;

    # HTTP API
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # WebSocket
    location /ws {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # Swagger 文档
    location /swagger/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
    }

    # 静态文件
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
    }
}
```

## 存储协议配置

### SMB/CIFS 配置

#### Samba 配置文件 `/etc/samba/smb.conf`
```ini
[global]
    workgroup = WORKGROUP
    server string = PNAS Server
    security = user
    map to guest = bad user
    dns proxy = no
    wide links = yes
    unix extensions = no

[data]
    comment = Data Share
    path = /mnt/data
    browseable = yes
    writable = yes
    guest ok = no
    valid users = @users
    create mask = 0755
    directory mask = 0755
```

### NFS 配置

#### 导出文件 `/etc/exports`
```
/mnt/nfs_share 192.168.1.0/24(rw,sync,no_subtree_check,no_root_squash)
/mnt/data      *(ro,sync,no_subtree_check)
```

#### 启动 NFS 服务
```bash
sudo systemctl enable nfs-kernel-server
sudo systemctl start nfs-kernel-server
sudo exportfs -ra
```

### iSCSI 配置

#### Target 配置（需要 targetcli）
```bash
# 安装 targetcli
sudo apt install targetcli-fb  # Ubuntu
sudo yum install targetcli     # CentOS

# 配置 iSCSI target
sudo targetcli
/> backstores/fileio create disk01 /var/lib/pnas/disk01.img 10G
/> iscsi/ create iqn.2023-12.com.example:target01
/> iscsi/iqn.2023-12.com.example:target01/tpg1/luns/ create /backstores/fileio/disk01
/> iscsi/iqn.2023-12.com.example:target01/tpg1/acls/ create iqn.2023-12.com.example:initiator01
/> exit
```

## 监控和日志

### 日志管理

#### 日志轮转配置 `/etc/logrotate.d/pnas`
```
/var/log/pnas/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    sharedscripts
    postrotate
        systemctl reload pnas
    endscript
}
```

### 系统监控

#### Prometheus 配置
```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'pnas'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/api/v1/monitoring/metrics'
    scrape_interval: 10s
```

#### Grafana 仪表板
导入预配置的 Grafana 仪表板 JSON 文件，包含：
- 系统资源监控
- 存储性能监控
- 网络流量监控
- 服务健康状态

## 备份和恢复

### 数据备份脚本
```bash
#!/bin/bash
# /opt/pnas/backup.sh

BACKUP_DIR="/backup/pnas"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# 备份数据库
cp /var/lib/pnas/pnas.db $BACKUP_DIR/pnas_${DATE}.db

# 备份配置文件
tar -czf $BACKUP_DIR/config_${DATE}.tar.gz /etc/pnas/

# 备份日志（最近7天）
find /var/log/pnas/ -name "*.log" -mtime -7 -exec cp {} $BACKUP_DIR/ \;

# 清理旧备份（保留30天）
find $BACKUP_DIR -type f -mtime +30 -delete

echo "Backup completed: $BACKUP_DIR"
```

### 定时备份
```bash
# 添加到 crontab
sudo crontab -e

# 每天凌晨2点备份
0 2 * * * /opt/pnas/backup.sh
```

## 安全建议

### 1. 用户权限
```bash
# 更改默认管理员密码
curl -X POST http://localhost:8080/api/v1/user/change-password \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"old_password":"admin123456","new_password":"your-strong-password"}'
```

### 2. SSL/TLS 配置
使用 Let's Encrypt 或自签名证书启用 HTTPS：
```bash
# 安装 certbot
sudo apt install certbot

# 获取证书
sudo certbot certonly --standalone -d pnas.example.com
```

### 3. 网络隔离
- 将 PNAS 部署在独立的 VLAN 中
- 限制管理端口只能从特定网段访问
- 使用 VPN 连接管理界面

### 4. 审计日志
启用系统审计日志记录所有管理操作：
```bash
sudo apt install auditd
sudo systemctl enable auditd
```

## 故障排除

### 常见问题

#### 1. 服务无法启动
```bash
# 检查日志
sudo journalctl -u pnas -n 50

# 检查配置文件
sudo pnas --config-check

# 检查端口占用
sudo netstat -tulpn | grep :8080
```

#### 2. 数据库连接错误
```bash
# 检查数据库文件权限
ls -la /var/lib/pnas/

# 重新创建数据库
sudo rm /var/lib/pnas/pnas.db
sudo systemctl restart pnas
```

#### 3. WebSocket 连接失败
```bash
# 检查防火墙
sudo ufw status

# 测试 WebSocket 连接
wscat -c ws://localhost:8080/ws
```

### 性能优化

#### 1. 数据库优化
```sql
-- 定期执行 VACUUM（SQLite）
PRAGMA auto_vacuum = INCREMENTAL;
PRAGMA journal_mode = WAL;
```

#### 2. 系统优化
```bash
# 增加文件描述符限制
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf

# 优化网络参数
echo "net.core.rmem_max = 134217728" >> /etc/sysctl.conf
echo "net.core.wmem_max = 134217728" >> /etc/sysctl.conf
sysctl -p
```

## 升级指南

### 版本升级步骤
1. 备份当前数据和配置
2. 停止 PNAS 服务
3. 替换二进制文件
4. 启动服务并验证
5. 检查数据库迁移

```bash
# 升级脚本示例
#!/bin/bash
sudo systemctl stop pnas
sudo cp /opt/pnas/pnas /opt/pnas/pnas.backup
sudo wget -O /opt/pnas/pnas https://github.com/your-org/pnas/releases/download/v1.1.0/pnas-linux-amd64
sudo chmod +x /opt/pnas/pnas
sudo systemctl start pnas
sudo systemctl status pnas
```

这份部署指南提供了从基础安装到生产环境运维的完整流程，确保 PNAS 系统能够稳定、安全地运行在各种环境中。