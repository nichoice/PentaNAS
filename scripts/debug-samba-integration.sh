#!/bin/bash

# Samba 集成测试调试脚本
# 用于在 Linux 环境中调试 Samba 集成测试问题

set -e

echo "=== Samba 集成测试调试脚本 ==="

# 检查基本环境
echo "1. 检查基本环境..."
echo "   操作系统: $(uname -s)"
echo "   内核版本: $(uname -r)"
echo "   Go 版本: $(go version)"

# 检查 Samba 安装
echo "2. 检查 Samba 安装..."
SAMBA_COMMANDS=("smbd" "testparm" "smbcontrol" "smbpasswd" "pdbedit")

for cmd in "${SAMBA_COMMANDS[@]}"; do
    if command -v "$cmd" &> /dev/null; then
        echo "   ✓ $cmd 已安装: $(which $cmd)"
    else
        echo "   ✗ $cmd 未安装"
        echo "   请运行: sudo apt-get install samba samba-common-bin"
        exit 1
    fi
done

# 检查 Samba 版本
echo "3. 检查 Samba 版本..."
if smbd --version &> /dev/null; then
    VERSION=$(smbd --version)
    echo "   版本: $VERSION"
else
    echo "   ✗ 无法获取版本信息"
fi

# 检查权限
echo "4. 检查权限..."
if [[ $EUID -eq 0 ]]; then
    echo "   ✓ 正在以 root 用户运行"
elif groups | grep -q sudo; then
    echo "   ✓ 当前用户在 sudo 组中"
else
    echo "   ⚠ 当前用户不在 sudo 组中，可能需要 root 权限"
fi

# 检查服务状态
echo "5. 检查服务状态..."
SAMBA_SERVICES=("smbd" "nmbd")

for service in "${SAMBA_SERVICES[@]}"; do
    if systemctl is-active --quiet "$service" 2>/dev/null; then
        echo "   ✓ $service 服务正在运行"
        PID=$(systemctl show "$service" --property=MainPID --value)
        echo "     进程 ID: $PID"
    else
        echo "   - $service 服务未运行，正在启动..."
        sudo systemctl start "$service" 2>/dev/null || echo "   ⚠ 无法启动 $service 服务"
    fi
done

# 检查配置文件
echo "6. 检查配置文件..."
SMB_CONF="/etc/samba/smb.conf"

if [[ -f "$SMB_CONF" ]]; then
    echo "   ✓ 配置文件存在: $SMB_CONF"
    echo "   文件大小: $(wc -c < "$SMB_CONF") 字节"
    echo "   最后修改: $(stat -c %y "$SMB_CONF")"
else
    echo "   ✗ 配置文件不存在: $SMB_CONF"
    exit 1
fi

# 验证配置
echo "7. 验证配置..."
if sudo testparm -s &> /dev/null; then
    echo "   ✓ 配置验证通过"
    echo "   配置概览:"
    sudo testparm -s | head -10
else
    echo "   ✗ 配置验证失败"
    sudo testparm -s
    exit 1
fi

# 检查网络端口
echo "8. 检查网络端口..."
SAMBA_PORTS=(139 445)

for port in "${SAMBA_PORTS[@]}"; do
    if netstat -tln 2>/dev/null | grep -q ":$port "; then
        echo "   ✓ 端口 $port 正在监听"
    else
        echo "   ⚠ 端口 $port 未监听"
    fi
done

# 检查防火墙
echo "9. 检查防火墙..."
if command -v ufw &> /dev/null && ufw status | grep -q "Status: active"; then
    echo "   检测到 UFW 防火墙"
    if ufw status | grep -q "139\|445\|Samba"; then
        echo "   ✓ Samba 端口已开放"
    else
        echo "   ⚠ Samba 端口可能被阻止"
        echo "   建议运行: sudo ufw allow samba"
    fi
elif command -v firewall-cmd &> /dev/null && firewall-cmd --state 2>/dev/null | grep -q "running"; then
    echo "   检测到 firewalld"
    if firewall-cmd --list-services | grep -q "samba"; then
        echo "   ✓ Samba 服务已开放"
    else
        echo "   ⚠ Samba 服务可能被阻止"
        echo "   建议运行: sudo firewall-cmd --add-service=samba --permanent && sudo firewall-cmd --reload"
    fi
else
    echo "   未检测到活动的防火墙"
fi

# 检查用户数据库
echo "10. 检查用户数据库..."
if sudo pdbedit -L &> /dev/null; then
    echo "   ✓ 用户数据库可访问"
    USER_COUNT=$(sudo pdbedit -L | wc -l)
    echo "   当前用户数量: $USER_COUNT"
    if [[ $USER_COUNT -gt 0 ]]; then
        echo "   用户列表:"
        sudo pdbedit -L | head -5 | sed 's/^/     /'
    fi
else
    echo "   ✗ 用户数据库不可访问"
fi

# 检查共享
echo "11. 检查共享..."
if testparm -s --section-name &> /dev/null; then
    echo "   ✓ 共享配置可读取"
    SHARES=$(testparm -s --section-name | grep -v global)
    SHARE_COUNT=$(echo "$SHARES" | wc -l)
    echo "   当前共享数量: $SHARE_COUNT"
    if [[ $SHARE_COUNT -gt 0 ]]; then
        echo "   共享列表:"
        echo "$SHARES" | head -5 | sed 's/^/     /'
    fi
else
    echo "   ✗ 无法读取共享配置"
fi

# 检查日志
echo "12. 检查日志..."
LOG_DIR="/var/log/samba"
if [[ -d "$LOG_DIR" ]]; then
    echo "   ✓ 日志目录存在: $LOG_DIR"
    echo "   日志文件:"
    ls -la "$LOG_DIR" | head -5 | sed 's/^/     /'

    # 检查最近的错误
    if [[ -f "$LOG_DIR/log.smbd" ]]; then
        echo "   最近的 smbd 日志 (最后10行):"
        tail -10 "$LOG_DIR/log.smbd" 2>/dev/null | sed 's/^/     /' || echo "     无法读取日志"
    fi
else
    echo "   ⚠ 日志目录不存在: $LOG_DIR"
fi

# 检查磁盘空间
echo "13. 检查磁盘空间..."
BACKUP_DIR="/var/backups/samba"
for dir in "$BACKUP_DIR" "/tmp" "/var/log"; do
    if [[ -d "$dir" ]]; then
        SPACE=$(df -h "$dir" | tail -1 | awk '{print $4}')
        echo "   $dir: $SPACE 可用"
    fi
done

# 运行最小化测试
echo "14. 运行最小化测试..."

# 测试配置重载
echo "   测试配置重载..."
if sudo smbcontrol all reload-config &> /dev/null; then
    echo "   ✓ 配置重载成功"
else
    echo "   ✗ 配置重载失败"
fi

# 测试用户管理
echo "   测试用户管理..."
TEST_USER="debug-test-user"
if sudo pdbedit -L | grep -q "$TEST_USER"; then
    echo "   - 清理已存在的测试用户..."
    sudo smbpasswd -x "$TEST_USER" &> /dev/null || true
fi

# 创建测试用户
if id "$TEST_USER" &>/dev/null; then
    echo "   - 系统用户已存在"
else
    echo "   - 创建系统用户..."
    sudo useradd -M -s /sbin/nologin "$TEST_USER" &> /dev/null || echo "   ⚠ 无法创建系统用户"
fi

echo "   - 创建 Samba 用户..."
if echo -e "TestPass123!\nTestPass123!" | sudo smbpasswd -a -s "$TEST_USER" &> /dev/null; then
    echo "   ✓ Samba 用户创建成功"

    # 测试用户删除
    echo "   - 删除测试用户..."
    sudo smbpasswd -x "$TEST_USER" &> /dev/null || echo "   ⚠ 删除 Samba 用户失败"
    sudo userdel "$TEST_USER" &> /dev/null || echo "   ⚠ 删除系统用户失败"
    echo "   ✓ 用户管理测试完成"
else
    echo "   ✗ Samba 用户创建失败"
fi

# 测试共享管理
echo "   测试共享管理..."
TEST_SHARE="debug-test-share"
TEST_PATH="/tmp/debug-test-share"

# 创建测试目录
mkdir -p "$TEST_PATH"

# 备份当前配置
BACKUP_FILE="/tmp/smb.conf.debug.backup.$(date +%s)"
sudo cp "$SMB_CONF" "$BACKUP_FILE"

# 添加测试共享
echo "   - 添加测试共享..."
cat << EOF | sudo tee -a "$SMB_CONF" > /dev/null

[$TEST_SHARE]
   path = $TEST_PATH
   comment = Debug Test Share
   browseable = yes
   writable = yes
   guest ok = no
EOF

# 验证配置
if sudo testparm -s &> /dev/null; then
    echo "   ✓ 测试共享添加成功"

    # 测试共享列表
    if testparm -s --section-name | grep -q "$TEST_SHARE"; then
        echo "   ✓ 共享在列表中可见"
    else
        echo "   ⚠ 共享在列表中不可见"
    fi

    # 重载配置
    if sudo smbcontrol all reload-config &> /dev/null; then
        echo "   ✓ 配置重载成功"
    else
        echo "   ⚠ 配置重载失败"
    fi
else
    echo "   ✗ 测试共享配置无效"
fi

# 恢复配置
echo "   - 恢复原始配置..."
sudo cp "$BACKUP_FILE" "$SMB_CONF"
sudo smbcontrol all reload-config &> /dev/null || true

# 清理
rm -f "$BACKUP_FILE"
rm -rf "$TEST_PATH"

echo ""
echo "=== 调试完成 ==="
echo ""
echo "如果以上测试都通过，现在可以运行 Go 测试:"
echo "  # 单元测试"
echo "  go test -v ./tests/samba_services_test.go"
echo ""
echo "  # 集成测试"
echo "  go test -v ./tests/integration/samba_integration_test.go"
echo ""
echo "如果测试仍然失败，请检查:"
echo "1. SELinux 设置 (如果启用): sudo setsebool -P samba_enable_home_dirs on"
echo "2. AppArmor 设置 (如果启用): 检查 Samba 相关策略"
echo "3. 磁盘空间是否充足"
echo "4. 系统日志: journalctl -u smbd -f"
echo "5. Samba 日志: tail -f /var/log/samba/log.smbd"
echo ""