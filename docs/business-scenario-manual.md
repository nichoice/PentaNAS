# PNAS 业务场景使用手册

## 概述

本手册以实际业务场景为导向，详细说明如何使用PNAS系统完成具体的存储服务配置。每个场景都包含完整的操作流程、组件关联关系和配置要点。

---

## 场景一：创建完整的iSCSI块存储服务

### 业务目标
为远程服务器提供块存储服务，客户端可以通过iSCSI协议挂载和使用远程磁盘。

### 组件关系图
```
存储池 (Storage Pool)
    ↓
逻辑单元 (LUN)
    ↓ 映射到
目标端 (Target/IQN)
    ↓ 控制访问
访问控制列表 (ACL)
    ↓ 包含
CHAP认证配置
```

### 完整操作流程

#### 第1步：准备底层存储
**目的**: 为iSCSI服务准备存储空间

1. **查看可用磁盘**
```bash
curl -X GET "http://localhost:8080/api/v1/storage/disks" \
  -H "Authorization: Bearer <token>"
```

2. **创建存储池** (可选，也可以直接使用文件)
```bash
curl -X POST "http://localhost:8080/api/v1/iscsi/pools" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "iscsi-pool-1",
    "type": "file",
    "path": "/data/iscsi/pool1",
    "size": 107374182400,
    "comment": "iSCSI存储池"
  }'
```

#### 第2步：创建iSCSI目标端 (Target)
**目的**: 创建iSCSI服务的唯一标识符(IQN)

```bash
curl -X POST "http://localhost:8080/api/v1/iscsi/targets" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "iqn.2024-01.com.example:storage.target01",
    "alias": "存储服务器01",
    "comment": "用于生产环境的iSCSI目标",
    "is_enabled": true
  }'
```

**关键信息**:
- IQN格式: `iqn.年份-月份.域名反写:标识符`
- 记录返回的target_id，后续步骤需要使用

#### 第3步：创建逻辑单元 (LUN)
**目的**: 创建实际的存储设备，客户端将访问这个逻辑单元

```bash
curl -X POST "http://localhost:8080/api/v1/iscsi/luns" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "lun-disk-01",
    "device_type": "file",
    "size": 53687091200,
    "device_path": "/data/iscsi/lun01.img",
    "comment": "50GB存储磁盘",
    "is_enabled": true,
    "read_only": false,
    "block_size": 512
  }'
```

**参数说明**:
- `device_type`: 设备类型
  - `file`: 文件型LUN (推荐用于测试)
  - `block`: 块设备型LUN (推荐用于生产)
  - `tcmu`: TCMU用户空间LUN
- `size`: 以字节为单位的LUN大小

#### 第4步：将LUN映射到Target
**目的**: 将逻辑单元关联到目标端，使其可被访问

```bash
curl -X POST "http://localhost:8080/api/v1/iscsi/luns/{lun_id}/map" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "target_id": "{target_id}",
    "lun_number": 0
  }'
```

**说明**:
- `lun_number`: LUN编号 (0-255)，同一个Target下的LUN编号必须唯一

#### 第5步：配置访问控制列表 (ACL)
**目的**: 控制哪些客户端可以访问iSCSI服务

```bash
curl -X POST "http://localhost:8080/api/v1/iscsi/acls" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "target_id": "{target_id}",
    "initiator_name": "iqn.2024-01.com.client:node01",
    "permission": "rw",
    "auth_type": "chap",
    "username": "iscsi_user",
    "password": "SecurePassword123",
    "mutual_auth": false,
    "is_enabled": true,
    "comment": "客户端服务器01的访问权限"
  }'
```

**参数详解**:
- `initiator_name`: 客户端的IQN，需要客户端提供
- `permission`: 访问权限
  - `rw`: 读写权限
  - `ro`: 只读权限
  - `deny`: 拒绝访问
- `auth_type`: 认证类型
  - `none`: 无认证
  - `chap`: CHAP认证 (推荐)
- `mutual_auth`: 是否启用双向认证

#### 第6步：启动Target服务
**目的**: 激活iSCSI目标端，使其可被客户端发现和连接

```bash
curl -X POST "http://localhost:8080/api/v1/iscsi/targets/{target_id}/start" \
  -H "Authorization: Bearer <token>"
```

#### 第7步：验证服务状态
**目的**: 确认iSCSI服务正常运行

```bash
# 检查Target状态
curl -X GET "http://localhost:8080/api/v1/iscsi/targets/{target_id}/status" \
  -H "Authorization: Bearer <token>"

# 检查服务整体状态
curl -X GET "http://localhost:8080/api/v1/iscsi/service/status" \
  -H "Authorization: Bearer <token>"
```

### 客户端连接配置

完成服务端配置后，客户端需要进行以下配置：

**Linux客户端**:
```bash
# 1. 发现iSCSI目标
sudo iscsiadm -m discovery -t st -p <服务器IP>:3260

# 2. 登录到目标
sudo iscsiadm -m node --targetname "iqn.2024-01.com.example:storage.target01" --login

# 3. 查看新增的块设备
lsblk
```

**Windows客户端**:
```cmd
# 1. 打开iSCSI Initiator
# 2. 在Discovery选项卡中添加目标门户：<服务器IP>:3260
# 3. 在Targets选项卡中连接到发现的目标
# 4. 在磁盘管理中初始化新磁盘
```

---

## 场景二：配置Samba文件共享服务

### 业务目标
为Windows和Linux客户端提供文件共享服务，支持用户认证、权限控制和高级功能。

### 组件关系图
```
全局配置 (Global Config)
    ↓
Samba账号 (Account)
    ↓ 授权访问
文件共享 (Share)
    ↓ 增强功能
时间机器/多通道/回收站
```

### 完整操作流程

#### 第1步：配置Samba全局设置
**目的**: 设置Samba服务的全局参数

```bash
curl -X PUT "http://localhost:8080/api/v1/samba/config/global" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "workgroup": "WORKGROUP",
    "server_string": "PNAS File Server",
    "security": "user",
    "enable_recycling": true,
    "enable_audit": true,
    "max_connections": 100,
    "enable_multichannel": true
  }'
```

#### 第2步：创建Samba账号
**目的**: 创建用于文件共享访问的用户账号

```bash
curl -X POST "http://localhost:8080/api/v1/samba/accounts" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "password": "SecurePass123",
    "full_name": "John Doe",
    "comment": "销售部门用户",
    "is_enabled": true,
    "is_admin": false
  }'
```

#### 第3步：创建文件共享
**目的**: 创建可被网络访问的共享目录

```bash
curl -X POST "http://localhost:8080/api/v1/samba/shares" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "sales_documents",
    "path": "/data/shares/sales",
    "comment": "销售部门文档共享",
    "read_only": false,
    "guest_ok": false,
    "browseable": true,
    "enable_recycling": true,
    "enable_audit": true,
    "create_mask": "0664",
    "directory_mask": "0775"
  }'
```

#### 第4步：配置共享访问权限
**目的**: 为特定用户分配共享访问权限

```bash
curl -X POST "http://localhost:8080/api/v1/samba/shares/access" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "share_id": "{share_id}",
    "account_id": "{account_id}",
    "permission": "read_write",
    "comment": "销售经理权限"
  }'
```

**权限类型**:
- `read_only`: 只读权限
- `read_write`: 读写权限
- `deny`: 拒绝访问

#### 第5步：启用时间机器功能 (macOS备份)
**目的**: 为macOS客户端提供Time Machine备份支持

```bash
curl -X PUT "http://localhost:8080/api/v1/samba/shares/{share_id}/timemachine/enable" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "quota_size": 107374182400,
    "model": "TimeCapsule8,119"
  }'
```

#### 第6步：启用多通道功能
**目的**: 提高网络传输性能和可靠性

```bash
curl -X PUT "http://localhost:8080/api/v1/samba/shares/{share_id}/multichannel/enable" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "max_channels": 4,
    "rss_capable": true
  }'
```

#### 第7步：重载配置并启动服务
**目的**: 应用配置变更并启动Samba服务

```bash
# 重载配置
curl -X POST "http://localhost:8080/api/v1/samba/config/reload" \
  -H "Authorization: Bearer <token>"

# 检查服务状态
curl -X GET "http://localhost:8080/api/v1/samba/monitoring/status" \
  -H "Authorization: Bearer <token>"
```

### 客户端访问配置

**Windows客户端**:
```
1. 打开资源管理器
2. 在地址栏输入：\\<服务器IP>\sales_documents
3. 输入Samba账号和密码
```

**macOS客户端**:
```
1. 打开Finder
2. 按Cmd+K，输入：smb://<服务器IP>/sales_documents
3. 输入Samba账号和密码
```

**Linux客户端**:
```bash
# 临时挂载
sudo mount -t cifs //<服务器IP>/sales_documents /mnt/share \
  -o username=john_doe,password=SecurePass123

# 永久挂载 (添加到/etc/fstab)
//<服务器IP>/sales_documents /mnt/share cifs username=john_doe,password=SecurePass123,uid=1000,gid=1000 0 0
```

---

## 场景三：管理系统存储

### 业务目标
管理物理磁盘，创建卷组和逻辑卷，为上层存储服务提供底层存储支持。

### 组件关系图
```
物理磁盘 (Physical Disk)
    ↓ 加入
卷组 (Volume Group)
    ↓ 创建
逻辑卷 (Logical Volume)
    ↓ 格式化挂载
文件系统
```

### 完整操作流程

#### 第1步：查看可用磁盘
**目的**: 了解系统中的物理磁盘状态

```bash
curl -X GET "http://localhost:8080/api/v1/storage/disks" \
  -H "Authorization: Bearer <token>"
```

#### 第2步：创建卷组 (VG)
**目的**: 将一个或多个物理磁盘组成卷组

```bash
curl -X POST "http://localhost:8080/api/v1/storage/create_vg" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "vg_name": "data_vg",
    "devices": ["/dev/sdb", "/dev/sdc"]
  }'
```

#### 第3步：查看卷组状态
**目的**: 确认卷组创建成功

```bash
curl -X GET "http://localhost:8080/api/v1/storage/vgs" \
  -H "Authorization: Bearer <token>"
```

#### 第4步：创建逻辑卷 (LV)
**目的**: 在卷组中创建可用的逻辑卷

```bash
curl -X POST "http://localhost:8080/api/v1/storage/create_lv" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "lv_name": "shares_lv",
    "vg_name": "data_vg",
    "size": "50G"
  }'
```

#### 第5步：格式化和挂载 (需要系统级操作)
**目的**: 创建文件系统并挂载逻辑卷

```bash
# 以下操作需要在服务器上直接执行
sudo mkfs.ext4 /dev/data_vg/shares_lv
sudo mkdir -p /data/shares
sudo mount /dev/data_vg/shares_lv /data/shares
```

---

## 场景四：监控和故障排查

### 业务目标
监控各种存储服务的运行状态，及时发现和解决问题。

### 监控检查清单

#### 1. 系统整体状态检查
```bash
# 系统信息
curl -X GET "http://localhost:8080/api/v1/monitoring/system" \
  -H "Authorization: Bearer <token>"

# CPU和内存状态
curl -X GET "http://localhost:8080/api/v1/monitoring/cpu" \
  -H "Authorization: Bearer <token>"

curl -X GET "http://localhost:8080/api/v1/monitoring/memory" \
  -H "Authorization: Bearer <token>"
```

#### 2. iSCSI服务监控
```bash
# 服务状态
curl -X GET "http://localhost:8080/api/v1/iscsi/service/status" \
  -H "Authorization: Bearer <token>"

# 活跃会话
curl -X GET "http://localhost:8080/api/v1/iscsi/sessions" \
  -H "Authorization: Bearer <token>"

# 性能统计
curl -X GET "http://localhost:8080/api/v1/iscsi/monitoring/performance" \
  -H "Authorization: Bearer <token>"
```

#### 3. Samba服务监控
```bash
# 服务状态
curl -X GET "http://localhost:8080/api/v1/samba/monitoring/status" \
  -H "Authorization: Bearer <token>"

# 活跃连接
curl -X GET "http://localhost:8080/api/v1/samba/monitoring/connections" \
  -H "Authorization: Bearer <token>"
```

#### 4. 审计日志检查
```bash
# 系统审计日志
curl -X GET "http://localhost:8080/api/v1/audit/logs" \
  -H "Authorization: Bearer <token>"

# 异常检测
curl -X GET "http://localhost:8080/api/v1/audit/anomalies" \
  -H "Authorization: Bearer <token>"
```

### 常见问题排查

#### iSCSI连接问题
1. **检查Target状态**: 确认Target是否启动
2. **检查ACL配置**: 验证客户端IQN是否在ACL中
3. **检查CHAP认证**: 确认用户名密码是否正确
4. **检查网络连通性**: 测试3260端口是否可达

#### Samba访问问题
1. **检查账号状态**: 确认Samba账号是否启用
2. **检查共享权限**: 验证用户是否有共享访问权限
3. **检查文件系统权限**: 确认底层目录权限设置
4. **检查防火墙**: 确认139和445端口是否开放

---

## 最佳实践建议

### 安全配置
1. **使用CHAP认证**: iSCSI服务建议启用CHAP认证
2. **最小权限原则**: 只分配必要的访问权限
3. **定期更换密码**: 定期更新认证密码
4. **启用审计**: 开启详细的审计日志记录

### 性能优化
1. **合理规划存储**: 根据业务需求分配存储资源
2. **启用多通道**: 对于高带宽需求启用SMB多通道
3. **调整块大小**: 根据访问模式调整块大小
4. **监控性能指标**: 定期检查IOPS和带宽使用情况

### 维护管理
1. **定期备份配置**: 备份系统配置文件
2. **监控存储使用**: 及时清理或扩容存储空间
3. **更新系统**: 保持系统和服务的最新版本
4. **文档记录**: 记录配置变更和故障处理过程

---

## 总结

通过本手册，你可以：
1. 从零搭建完整的iSCSI块存储服务
2. 配置功能完善的Samba文件共享
3. 管理底层存储资源
4. 监控和维护各种存储服务

每个业务场景都遵循了组件化的设计思路，先准备底层资源，再配置核心服务，最后添加访问控制和增强功能。这种分层的配置方式既保证了服务的完整性，也便于问题排查和维护管理。