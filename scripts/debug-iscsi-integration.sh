#!/bin/bash

# iSCSI 集成测试调试脚本
# 用于在 Linux 环境中调试集成测试问题

set -e

echo "=== iSCSI 集成测试调试脚本 ==="

# 检查基本环境
echo "1. 检查基本环境..."
echo "   操作系统: $(uname -s)"
echo "   内核版本: $(uname -r)"
echo "   Go 版本: $(go version)"

# 检查 targetcli 安装
echo "2. 检查 targetcli 安装..."
if command -v targetcli &> /dev/null; then
    echo "   ✓ targetcli 已安装: $(which targetcli)"
    echo "   版本信息:"
    targetcli --version || echo "   无法获取版本信息"
else
    echo "   ✗ targetcli 未安装"
    echo "   请运行: sudo apt-get install targetcli-fb"
    exit 1
fi

# 检查权限
echo "3. 检查权限..."
if [[ $EUID -eq 0 ]]; then
    echo "   ✓ 正在以 root 用户运行"
elif groups | grep -q sudo; then
    echo "   ✓ 当前用户在 sudo 组中"
else
    echo "   ⚠ 当前用户不在 sudo 组中，可能需要 root 权限"
fi

# 检查内核模块
echo "4. 检查内核模块..."
REQUIRED_MODULES=("target_core_mod" "iscsi_target_mod")

for module in "${REQUIRED_MODULES[@]}"; do
    if lsmod | grep -q "^$module"; then
        echo "   ✓ $module 已加载"
    else
        echo "   - 正在加载 $module..."
        sudo modprobe "$module" 2>/dev/null || echo "   ⚠ 无法加载 $module"
    fi
done

# 检查服务状态
echo "5. 检查服务状态..."
if systemctl is-active --quiet target 2>/dev/null; then
    echo "   ✓ target 服务正在运行"
else
    echo "   - 启动 target 服务..."
    sudo systemctl start target 2>/dev/null || echo "   ⚠ 无法启动 target 服务"
fi

# 检查 configfs
echo "6. 检查 configfs..."
if mount | grep -q configfs; then
    echo "   ✓ configfs 已挂载"
    echo "   挂载点: $(mount | grep configfs | awk '{print $3}')"
else
    echo "   - 挂载 configfs..."
    sudo mount -t configfs configfs /sys/kernel/config 2>/dev/null || echo "   ⚠ 无法挂载 configfs"
fi

# 测试 targetcli 基本功能
echo "7. 测试 targetcli 基本功能..."
if sudo targetcli ls &> /dev/null; then
    echo "   ✓ targetcli 基本功能正常"
    echo "   当前配置概览:"
    sudo targetcli ls | head -10
else
    echo "   ✗ targetcli 基本功能测试失败"
    exit 1
fi

# 检查临时目录权限
echo "8. 检查临时目录权限..."
TEST_FILE="/tmp/iscsi_test_$(date +%s).img"
if touch "$TEST_FILE" 2>/dev/null; then
    echo "   ✓ 可以在 /tmp 创建文件"
    rm -f "$TEST_FILE"
else
    echo "   ✗ 无法在 /tmp 创建文件"
fi

# 运行最小化测试
echo "9. 运行最小化 targetcli 测试..."
TEST_IQN="iqn.2024-01.com.pnas:debug-test"
TEST_BACKSTORE="debug_backstore"
TEST_FILE="/tmp/debug_test.img"

echo "   创建测试文件 backstore..."
if sudo targetcli /backstores/fileio create name="$TEST_BACKSTORE" file_or_dev="$TEST_FILE" size=10M &> /dev/null; then
    echo "   ✓ Backstore 创建成功"

    echo "   创建测试 target..."
    if sudo targetcli /iscsi create "$TEST_IQN" &> /dev/null; then
        echo "   ✓ Target 创建成功"

        echo "   创建 LUN 映射..."
        if sudo targetcli "/iscsi/$TEST_IQN/tpg1/luns" create "/backstores/fileio/$TEST_BACKSTORE" 0 &> /dev/null; then
            echo "   ✓ LUN 映射成功"
        else
            echo "   ✗ LUN 映射失败"
        fi

        echo "   清理测试资源..."
        sudo targetcli "/iscsi/$TEST_IQN/tpg1/luns" delete 0 &> /dev/null || true
        sudo targetcli /iscsi delete "$TEST_IQN" &> /dev/null || true
    else
        echo "   ✗ Target 创建失败"
    fi

    sudo targetcli /backstores/fileio delete "$TEST_BACKSTORE" &> /dev/null || true
    rm -f "$TEST_FILE"
else
    echo "   ✗ Backstore 创建失败"
fi

echo ""
echo "=== 调试完成 ==="
echo "如果以上测试都通过，现在可以运行 Go 集成测试:"
echo "  go test -v ./tests/integration/iscsi_integration_test.go"
echo ""
echo "如果集成测试仍然失败，请查看具体错误消息并检查:"
echo "1. 是否有足够的磁盘空间"
echo "2. SELinux 或 AppArmor 是否阻止操作"
echo "3. 防火墙设置"
echo "4. 系统日志: journalctl -u target.service -f"