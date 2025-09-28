# iSCSI 功能测试指南

本指南说明如何在 Linux 环境中测试 PNAS 的 iSCSI 功能。

## 环境要求

- **Linux 系统** (Ubuntu 18.04+, CentOS 7+, RHEL 7+)
- root 或 sudo 权限
- Go 1.19+
- targetcli-fb 和 open-iscsi 软件包

**注意**: 集成测试仅在 Linux 环境中运行，因为需要真实的 targetcli 工具和内核模块。

## 快速开始

### 1. 复制项目到 Linux 环境

```bash
# 将整个项目目录复制到 Linux 机器
scp -r pnas/ user@linux-server:/path/to/
# 或使用 git clone
git clone <repository-url>
cd pnas
```

### 2. 运行环境设置脚本

```bash
chmod +x scripts/setup-iscsi-test.sh
./scripts/setup-iscsi-test.sh
```

此脚本将：
- 安装必需的软件包 (targetcli-fb, open-iscsi)
- 加载必需的内核模块
- 启动相关服务
- 验证环境配置

### 3. 运行集成测试

```bash
# 运行完整的集成测试（需要在 Linux 环境中运行）
go test -v ./tests/integration/iscsi_integration_test.go

# 运行特定的集成测试
go test -v ./tests/integration/iscsi_integration_test.go -run TestISCSITargetService_Integration
go test -v ./tests/integration/iscsi_integration_test.go -run TestISCSILUNService_Integration

# 跳过集成测试（只运行单元测试）
go test -short ./tests/iscsi_services_test.go
```

## 详细步骤

### 手动环境设置

如果自动脚本失败，可以手动执行以下步骤：

#### 1. 安装软件包

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install -y targetcli-fb open-iscsi sqlite3
```

**CentOS/RHEL:**
```bash
sudo yum install -y targetcli open-iscsi sqlite
```

**Fedora:**
```bash
sudo dnf install -y targetcli open-iscsi sqlite
```

#### 2. 加载内核模块

```bash
sudo modprobe target_core_mod
sudo modprobe iscsi_target_mod
sudo modprobe target_core_file
sudo modprobe target_core_pscsi
```

#### 3. 启动服务

```bash
sudo systemctl start target
sudo systemctl start iscsid
sudo systemctl enable target
sudo systemctl enable iscsid
```

#### 4. 挂载 configfs

```bash
sudo mount -t configfs configfs /sys/kernel/config
```

### 验证环境

```bash
# 检查 targetcli 是否工作
sudo targetcli ls

# 检查服务状态
systemctl status target
systemctl status iscsid

# 检查内核模块
lsmod | grep target
lsmod | grep iscsi
```

## 测试说明

### 单元测试 vs 集成测试

- **单元测试** (`tests/iscsi_services_test.go`):
  - 使用模拟的 targetcli，测试业务逻辑
  - 可在任意平台运行（包括 macOS）
  - 验证代码逻辑和参数传递

- **集成测试** (`tests/integration/iscsi_integration_test.go`):
  - 使用真实的 targetcli，测试系统集成
  - **仅在 Linux 环境运行**
  - 验证与实际 Linux 内核的交互

### 集成测试内容

集成测试将验证：

1. **Target 创建**: 调用真实的 `targetcli /iscsi create <iqn>`
2. **Portal 设置**: 配置网络门户
3. **Target 启用/禁用**: 测试 TPG 的启用和禁用
4. **Target 删除**: 清理创建的资源
5. **Backstore 创建**: 创建文件和块设备后端存储
6. **LUN 创建**: 创建逻辑单元号
7. **LUN 映射**: 将 LUN 映射到 Target

### 手动验证

测试运行后，可以手动验证：

```bash
# 查看所有 iSCSI targets
sudo targetcli ls /iscsi

# 查看特定 target 的详细信息
sudo targetcli ls /iscsi/iqn.2024-01.com.pnas:integration-test

# 查看 target 的 LUN 映射
sudo targetcli ls /iscsi/iqn.2024-01.com.pnas:lun-test/tpg1/luns

# 查看所有 backstores
sudo targetcli ls /backstores

# 查看文件类型的 backstores
sudo targetcli ls /backstores/fileio

# 查看块设备类型的 backstores
sudo targetcli ls /backstores/block

# 查看系统中的 target 配置文件
ls -la /sys/kernel/config/target/iscsi/

# 清理所有配置（测试完成后）
sudo targetcli clearconfig confirm=True
```

## 故障排除

### 常见问题

1. **权限错误**
   ```
   ERROR: configfs not mounted
   ```
   解决：`sudo mount -t configfs configfs /sys/kernel/config`

2. **模块加载失败**
   ```
   modprobe: FATAL: Module target_core_mod not found
   ```
   解决：检查内核版本和配置，可能需要安装内核模块包

3. **服务启动失败**
   ```
   Failed to start target.service
   ```
   解决：检查系统日志 `journalctl -u target.service`

4. **targetcli 命令失败**
   ```
   targetcli: command not found
   ```
   解决：安装 `targetcli-fb` 包

### 日志查看

```bash
# 查看 target 服务日志
journalctl -u target.service -f

# 查看 iscsid 服务日志
journalctl -u iscsid.service -f

# 查看内核消息
dmesg | grep -i iscsi
dmesg | grep -i target
```

## 测试清理

测试完成后，清理环境：

```bash
# 清理 targetcli 配置
sudo targetcli clearconfig confirm=True

# 停止服务（如果不需要）
sudo systemctl stop target
sudo systemctl stop iscsid

# 卸载模块（如果不需要）
sudo modprobe -r iscsi_target_mod
sudo modprobe -r target_core_mod
```

## 安全注意事项

- 测试环境应该是隔离的，不要在生产环境运行
- 测试会创建和删除 iSCSI targets，确保不会影响现有配置
- 部分操作需要 root 权限，请谨慎执行