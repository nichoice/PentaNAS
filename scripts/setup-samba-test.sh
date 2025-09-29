#!/bin/bash

# Samba 集成测试环境设置脚本
# 在 Linux 环境中运行此脚本来准备 Samba 测试环境

set -e

echo "=== Samba 集成测试环境设置 ==="

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
    sudo apt-get install -y samba samba-common-bin smbclient cifs-utils
elif command -v yum &> /dev/null; then
    # CentOS/RHEL
    echo "使用 yum 安装软件包..."
    sudo yum install -y samba samba-common samba-client cifs-utils
elif command -v dnf &> /dev/null; then
    # Fedora
    echo "使用 dnf 安装软件包..."
    sudo dnf install -y samba samba-common samba-client cifs-utils
else
    echo "错误: 未检测到支持的包管理器"
    exit 1
fi

# 检查 Samba 安装
echo "2. 检查 Samba 安装..."
SAMBA_COMMANDS=("smbd" "testparm" "smbcontrol" "smbpasswd" "pdbedit")

for cmd in "${SAMBA_COMMANDS[@]}"; do
    if command -v "$cmd" &> /dev/null; then
        echo "✓ $cmd 已安装: $(which $cmd)"
    else
        echo "✗ $cmd 未找到"
        exit 1
    fi
done

# 检查 Samba 版本
echo "3. 检查 Samba 版本..."
if smbd --version &> /dev/null; then
    SAMBA_VERSION=$(smbd --version)
    echo "✓ Samba 版本: $SAMBA_VERSION"
else
    echo "✗ 无法获取 Samba 版本"
    exit 1
fi

# 备份原始配置
echo "4. 备份原始配置..."
SMB_CONF="/etc/samba/smb.conf"
BACKUP_DIR="/var/backups/samba"

if [[ -f "$SMB_CONF" ]]; then
    sudo mkdir -p "$BACKUP_DIR"
    BACKUP_FILE="$BACKUP_DIR/smb.conf.backup.$(date +%Y%m%d-%H%M%S)"
    sudo cp "$SMB_CONF" "$BACKUP_FILE"
    echo "✓ 配置已备份到: $BACKUP_FILE"
else
    echo "警告: Samba 配置文件不存在，将创建基本配置"
fi

# 创建基本的 Samba 配置（如果不存在）
echo "5. 检查/创建基本配置..."
if [[ ! -f "$SMB_CONF" ]] || [[ ! -s "$SMB_CONF" ]]; then
    echo "创建基本的 Samba 配置..."
    sudo tee "$SMB_CONF" > /dev/null <<'EOF'
[global]
   workgroup = WORKGROUP
   server string = PNAS Samba Server
   netbios name = PNAS
   security = user
   map to guest = Bad User
   dns proxy = no
   log file = /var/log/samba/log.%m
   max log size = 1000
   logging = file

   # 性能优化
   socket options = TCP_NODELAY IPTOS_LOWDELAY SO_RCVBUF=65536 SO_SNDBUF=65536
   read raw = yes
   write raw = yes
   oplocks = yes
   max xmit = 65535
   dead time = 15
   getwd cache = yes

   # 多通道支持
   server multi channel support = yes

# 示例共享（用于测试）
[test]
   path = /tmp/samba-test
   browseable = yes
   writable = yes
   guest ok = no
   create mask = 0664
   directory mask = 0775
   comment = Test Share for PNAS
EOF
    echo "✓ 基本配置已创建"
else
    echo "✓ 配置文件已存在"
fi

# 验证配置
echo "6. 验证 Samba 配置..."
if sudo testparm -s &> /dev/null; then
    echo "✓ Samba 配置验证通过"
    echo "配置概览:"
    sudo testparm -s | head -20
else
    echo "✗ Samba 配置验证失败"
    sudo testparm -s
    exit 1
fi

# 创建测试目录
echo "7. 创建测试目录..."
TEST_DIRS=("/tmp/samba-test" "/tmp/pnas-test-share" "/tmp/pnas-service-test")

for dir in "${TEST_DIRS[@]}"; do
    if [[ ! -d "$dir" ]]; then
        mkdir -p "$dir"
        chmod 755 "$dir"
        echo "✓ 创建目录: $dir"
    else
        echo "✓ 目录已存在: $dir"
    fi
done

# 启动 Samba 服务
echo "8. 启动 Samba 服务..."
SAMBA_SERVICES=("smbd" "nmbd")

for service in "${SAMBA_SERVICES[@]}"; do
    if systemctl is-active --quiet "$service"; then
        echo "✓ $service 服务已运行"
    else
        echo "启动 $service 服务..."
        sudo systemctl start "$service" || echo "警告: 无法启动 $service 服务"
        sudo systemctl enable "$service" || echo "警告: 无法启用 $service 服务"
    fi
done

# 检查服务状态
echo "9. 检查服务状态..."
for service in "${SAMBA_SERVICES[@]}"; do
    if systemctl is-active --quiet "$service"; then
        echo "✓ $service: $(systemctl is-active $service)"
        # 获取进程信息
        PID=$(systemctl show "$service" --property=MainPID --value)
        if [[ "$PID" != "0" ]]; then
            echo "  进程 ID: $PID"
        fi
    else
        echo "⚠ $service: $(systemctl is-active $service)"
    fi
done

# 创建测试用户
echo "10. 创建测试用户..."
TEST_USER="pnas-test"
TEST_PASS="TestPass123!"

if id "$TEST_USER" &>/dev/null; then
    echo "✓ 测试用户 $TEST_USER 已存在"
else
    echo "创建测试用户 $TEST_USER..."
    sudo useradd -M -s /sbin/nologin "$TEST_USER" || echo "警告: 无法创建系统用户"
fi

# 添加 Samba 用户
if sudo pdbedit -L | grep -q "$TEST_USER"; then
    echo "✓ Samba 用户 $TEST_USER 已存在"
else
    echo "添加 Samba 用户 $TEST_USER..."
    echo -e "$TEST_PASS\n$TEST_PASS" | sudo smbpasswd -a -s "$TEST_USER" || echo "警告: 无法添加 Samba 用户"
fi

# 检查防火墙
echo "11. 检查防火墙设置..."
if command -v ufw &> /dev/null && ufw status | grep -q "Status: active"; then
    echo "检测到 UFW 防火墙，检查 Samba 端口..."
    if ! ufw status | grep -q "139\|445"; then
        echo "建议开放 Samba 端口："
        echo "  sudo ufw allow samba"
    else
        echo "✓ Samba 端口已开放"
    fi
elif command -v firewall-cmd &> /dev/null && firewall-cmd --state 2>/dev/null | grep -q "running"; then
    echo "检测到 firewalld，检查 Samba 服务..."
    if ! firewall-cmd --list-services | grep -q "samba"; then
        echo "建议开放 Samba 服务："
        echo "  sudo firewall-cmd --add-service=samba --permanent"
        echo "  sudo firewall-cmd --reload"
    else
        echo "✓ Samba 服务已开放"
    fi
else
    echo "未检测到活动的防火墙或防火墙未启用"
fi

# 测试基本功能
echo "12. 测试基本功能..."

# 测试配置重载
if sudo smbcontrol all reload-config &> /dev/null; then
    echo "✓ 配置重载测试通过"
else
    echo "⚠ 配置重载测试失败"
fi

# 测试用户列表
if sudo pdbedit -L &> /dev/null; then
    echo "✓ 用户列表功能正常"
    USER_COUNT=$(sudo pdbedit -L | wc -l)
    echo "  当前 Samba 用户数量: $USER_COUNT"
else
    echo "⚠ 用户列表功能异常"
fi

# 测试共享列表
if testparm -s --section-name &> /dev/null; then
    echo "✓ 共享列表功能正常"
    SHARE_COUNT=$(testparm -s --section-name | grep -v global | wc -l)
    echo "  当前共享数量: $SHARE_COUNT"
else
    echo "⚠ 共享列表功能异常"
fi

# 检查权限
echo "13. 检查权限..."
if groups | grep -q sudo; then
    echo "✓ 当前用户在 sudo 组中"
else
    echo "警告: 当前用户不在 sudo 组中，可能需要 root 权限"
fi

# 检查 SELinux（如果存在）
if command -v sestatus &> /dev/null; then
    echo "14. 检查 SELinux..."
    SELINUX_STATUS=$(sestatus | grep "SELinux status" | awk '{print $3}')
    if [[ "$SELINUX_STATUS" == "enabled" ]]; then
        echo "⚠ SELinux 已启用，可能需要配置 Samba 相关策略"
        echo "建议运行: sudo setsebool -P samba_enable_home_dirs on"
    else
        echo "✓ SELinux 未启用或已禁用"
    fi
fi

echo ""
echo "=== 环境设置完成 ==="
echo ""
echo "现在可以运行 Samba 测试:"
echo "  # 单元测试"
echo "  go test -v ./tests/samba_services_test.go"
echo ""
echo "  # 集成测试"
echo "  go test -v ./tests/integration/samba_integration_test.go"
echo ""
echo "手动验证命令:"
echo "  sudo testparm -s                    # 验证配置"
echo "  sudo pdbedit -L                     # 列出用户"
echo "  sudo smbcontrol all reload-config   # 重载配置"
echo "  systemctl status smbd nmbd          # 检查服务状态"
echo ""
echo "清理测试用户和目录:"
echo "  sudo smbpasswd -x pnas-test         # 删除测试用户"
echo "  sudo userdel pnas-test              # 删除系统用户"
echo "  rm -rf /tmp/samba-test /tmp/pnas-*  # 删除测试目录"
echo ""
echo "恢复配置:"
echo "  sudo cp $BACKUP_FILE $SMB_CONF      # 恢复原始配置"
echo ""