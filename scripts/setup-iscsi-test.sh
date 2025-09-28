#!/bin/bash

# iSCSI 集成测试环境设置脚本
# 在 Linux 环境中运行此脚本来准备 iSCSI 测试环境

set -e

echo "=== iSCSI 集成测试环境设置 ==="

# 检查是否为 root 用户
if [[ $EUID -eq 0 ]]; then
   echo "警告: 正在以 root 用户运行"
else
   echo "注意: 某些操作可能需要 sudo 权限"
fi

# 检查 Linux 发行版
if [[ -f /etc/os-release ]]; then
    . /etc/os-release
    echo "检测到 Linux 发行版: $NAME $VERSION"
else
    echo "警告: 无法检测 Linux 发行版"
fi

# 安装必需的包
echo "1. 安装必需的软件包..."

if command -v apt-get &> /dev/null; then
    # Ubuntu/Debian
    echo "使用 apt 安装软件包..."
    sudo apt-get update
    sudo apt-get install -y targetcli-fb open-iscsi sqlite3
elif command -v yum &> /dev/null; then
    # CentOS/RHEL
    echo "使用 yum 安装软件包..."
    sudo yum install -y targetcli open-iscsi sqlite
elif command -v dnf &> /dev/null; then
    # Fedora
    echo "使用 dnf 安装软件包..."
    sudo dnf install -y targetcli open-iscsi sqlite
else
    echo "错误: 未检测到支持的包管理器"
    exit 1
fi

# 检查 targetcli 是否可用
echo "2. 检查 targetcli 安装..."
if command -v targetcli &> /dev/null; then
    echo "✓ targetcli 已安装: $(which targetcli)"
    targetcli --version || echo "版本信息不可用"
else
    echo "✗ targetcli 未找到"
    exit 1
fi

# 检查并加载必需的内核模块
echo "3. 检查内核模块..."
REQUIRED_MODULES=("target_core_mod" "iscsi_target_mod" "target_core_file" "target_core_pscsi")

for module in "${REQUIRED_MODULES[@]}"; do
    if lsmod | grep -q "^$module"; then
        echo "✓ $module 已加载"
    else
        echo "  正在加载 $module..."
        sudo modprobe "$module" || echo "警告: 无法加载 $module"
    fi
done

# 启动相关服务
echo "4. 启动 iSCSI 服务..."
if systemctl is-active --quiet target; then
    echo "✓ target 服务已运行"
else
    echo "  启动 target 服务..."
    sudo systemctl start target || echo "警告: 无法启动 target 服务"
fi

if systemctl is-active --quiet iscsid; then
    echo "✓ iscsid 服务已运行"
else
    echo "  启动 iscsid 服务..."
    sudo systemctl start iscsid || echo "警告: 无法启动 iscsid 服务"
fi

# 检查 configfs 挂载
echo "5. 检查 configfs..."
if mount | grep -q configfs; then
    echo "✓ configfs 已挂载"
else
    echo "  挂载 configfs..."
    sudo mount -t configfs configfs /sys/kernel/config || echo "警告: 无法挂载 configfs"
fi

# 测试 targetcli 基本功能
echo "6. 测试 targetcli 基本功能..."
if sudo targetcli ls &> /dev/null; then
    echo "✓ targetcli 功能正常"
    echo "当前 target 配置:"
    sudo targetcli ls | head -10
else
    echo "✗ targetcli 测试失败"
    exit 1
fi

# 检查权限
echo "7. 检查当前用户权限..."
if groups | grep -q sudo; then
    echo "✓ 当前用户在 sudo 组中"
else
    echo "警告: 当前用户不在 sudo 组中，可能需要 root 权限"
fi

echo ""
echo "=== 环境设置完成 ==="
echo "现在可以运行集成测试:"
echo "  cd /path/to/pnas"
echo "  go test -v ./tests/integration/iscsi_integration_test.go"
echo ""
echo "手动验证命令:"
echo "  sudo targetcli ls                    # 查看当前配置"
echo "  sudo targetcli clearconfig confirm=True  # 清理配置"
echo ""