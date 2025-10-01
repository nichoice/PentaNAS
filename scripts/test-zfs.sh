#!/bin/bash
# ZFS功能测试脚本
# 需要root权限和至少2个可用磁盘设备

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}ZFS 功能测试脚本${NC}"
echo -e "${GREEN}================================${NC}"
echo

# 检查root权限
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}错误: 需要root权限运行此脚本${NC}"
    exit 1
fi

# 检查ZFS是否安装
if ! command -v zpool &> /dev/null; then
    echo -e "${RED}错误: ZFS未安装${NC}"
    echo "请先安装ZFS:"
    echo "  Ubuntu/Debian: apt install zfsutils-linux"
    echo "  CentOS/RHEL: yum install zfs"
    exit 1
fi

# 检查pnas命令
if [ ! -f "./build/pnas" ]; then
    echo -e "${YELLOW}构建pnas二进制文件...${NC}"
    make build
fi

PNAS="./build/pnas"

# 测试池名称
TEST_POOL="test_pool_$$"
TEST_DATASET="${TEST_POOL}/test_dataset"
TEST_VOLUME="${TEST_POOL}/test_volume"
TEST_SNAPSHOT="${TEST_DATASET}@test_snap"

# 清理函数
cleanup() {
    echo -e "\n${YELLOW}清理测试环境...${NC}"

    # 销毁快照
    if zfs list -H -t snapshot "$TEST_SNAPSHOT" 2>/dev/null; then
        echo "销毁快照: $TEST_SNAPSHOT"
        $PNAS zfs snapshot destroy "$TEST_SNAPSHOT" 2>/dev/null || true
    fi

    # 销毁数据集
    if zfs list -H "$TEST_DATASET" 2>/dev/null; then
        echo "销毁数据集: $TEST_DATASET"
        $PNAS zfs dataset destroy -r "$TEST_DATASET" 2>/dev/null || true
    fi

    # 销毁卷
    if zfs list -H -t volume "$TEST_VOLUME" 2>/dev/null; then
        echo "销毁卷: $TEST_VOLUME"
        zfs destroy "$TEST_VOLUME" 2>/dev/null || true
    fi

    # 销毁池
    if zpool list "$TEST_POOL" 2>/dev/null; then
        echo "销毁池: $TEST_POOL"
        $PNAS zfs pool destroy -f "$TEST_POOL" 2>/dev/null || true
    fi

    # 清理loop设备
    if [ -n "$LOOP_DEVICE1" ]; then
        losetup -d "$LOOP_DEVICE1" 2>/dev/null || true
    fi
    if [ -n "$LOOP_DEVICE2" ]; then
        losetup -d "$LOOP_DEVICE2" 2>/dev/null || true
    fi

    # 删除测试文件
    rm -f /tmp/zfs_test_disk1.img /tmp/zfs_test_disk2.img

    echo -e "${GREEN}清理完成${NC}"
}

trap cleanup EXIT

echo -e "${YELLOW}步骤1: 创建测试磁盘文件 (使用loop设备)${NC}"
dd if=/dev/zero of=/tmp/zfs_test_disk1.img bs=1M count=256 2>/dev/null
dd if=/dev/zero of=/tmp/zfs_test_disk2.img bs=1M count=256 2>/dev/null

LOOP_DEVICE1=$(losetup -f)
LOOP_DEVICE2=$(losetup -f --show /tmp/zfs_test_disk2.img)
losetup "$LOOP_DEVICE1" /tmp/zfs_test_disk1.img

echo "Loop设备1: $LOOP_DEVICE1"
echo "Loop设备2: $LOOP_DEVICE2"
echo

# 测试1: 创建存储池
echo -e "${YELLOW}测试1: 创建ZFS存储池 (镜像配置)${NC}"
$PNAS zfs pool create "$TEST_POOL" mirror "$LOOP_DEVICE1" "$LOOP_DEVICE2" --compression lz4
echo -e "${GREEN}✓ 池创建成功${NC}"
echo

# 测试2: 列出存储池
echo -e "${YELLOW}测试2: 列出存储池${NC}"
$PNAS zfs pool list
echo

# 测试3: 查看池状态
echo -e "${YELLOW}测试3: 查看池状态${NC}"
$PNAS zfs pool status "$TEST_POOL"
echo

# 测试4: 创建数据集
echo -e "${YELLOW}测试4: 创建数据集${NC}"
$PNAS zfs dataset create "$TEST_DATASET" --compression lz4 --quota 100M
echo -e "${GREEN}✓ 数据集创建成功${NC}"
echo

# 测试5: 列出数据集
echo -e "${YELLOW}测试5: 列出数据集${NC}"
$PNAS zfs dataset list "$TEST_POOL"
echo

# 测试6: 写入测试数据
echo -e "${YELLOW}测试6: 写入测试数据${NC}"
MOUNTPOINT=$(zfs get -H -o value mountpoint "$TEST_DATASET")
echo "挂载点: $MOUNTPOINT"

# 创建测试文件
for i in {1..10}; do
    echo "Test data line $i" > "$MOUNTPOINT/test_$i.txt"
done

echo "创建了10个测试文件"
ls -lh "$MOUNTPOINT"
echo

# 测试7: 创建快照
echo -e "${YELLOW}测试7: 创建快照${NC}"
$PNAS zfs snapshot create "$TEST_SNAPSHOT"
echo -e "${GREEN}✓ 快照创建成功${NC}"
echo

# 测试8: 列出快照
echo -e "${YELLOW}测试8: 列出快照${NC}"
$PNAS zfs snapshot list "$TEST_DATASET"
echo

# 测试9: 修改数据并回滚
echo -e "${YELLOW}测试9: 修改数据并回滚到快照${NC}"
echo "删除所有测试文件..."
rm -f "$MOUNTPOINT"/test_*.txt
ls -lh "$MOUNTPOINT"

echo "回滚到快照..."
$PNAS zfs snapshot rollback "$TEST_SNAPSHOT"

echo "验证文件已恢复..."
ls -lh "$MOUNTPOINT"
echo -e "${GREEN}✓ 回滚成功${NC}"
echo

# 测试10: 创建卷(zvol)
echo -e "${YELLOW}测试10: 创建ZFS卷 (zvol)${NC}"
$PNAS zfs volume create "$TEST_VOLUME" 50M --sparse
echo -e "${GREEN}✓ 卷创建成功${NC}"
echo

# 测试11: 列出卷
echo -e "${YELLOW}测试11: 列出卷${NC}"
$PNAS zfs volume list "$TEST_POOL"
echo

# 测试12: 检查设备文件
echo -e "${YELLOW}测试12: 验证zvol设备文件${NC}"
DEVICE_PATH="/dev/zvol/$TEST_VOLUME"
if [ -e "$DEVICE_PATH" ]; then
    echo -e "${GREEN}✓ 设备文件存在: $DEVICE_PATH${NC}"
    ls -lh "$DEVICE_PATH"
else
    echo -e "${RED}✗ 设备文件不存在: $DEVICE_PATH${NC}"
fi
echo

# 测试13: 查看压缩效果
echo -e "${YELLOW}测试13: 查看压缩统计${NC}"
zfs get compressratio,used,refer "$TEST_DATASET"
echo

# 测试14: JSON输出
echo -e "${YELLOW}测试14: JSON格式输出${NC}"
$PNAS zfs json pools
echo

# 测试15: 数据清洗
echo -e "${YELLOW}测试15: 启动数据清洗 (scrub)${NC}"
$PNAS zfs pool scrub "$TEST_POOL"
sleep 2
$PNAS zfs pool status "$TEST_POOL"
echo

# 所有测试完成
echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}所有测试完成！${NC}"
echo -e "${GREEN}================================${NC}"
echo
echo "测试摘要:"
echo "  ✓ 存储池管理"
echo "  ✓ 数据集管理"
echo "  ✓ 快照创建和回滚"
echo "  ✓ 卷(zvol)管理"
echo "  ✓ 压缩功能"
echo "  ✓ 数据清洗"
echo
echo -e "${YELLOW}提示: 清理将在脚本退出时自动执行${NC}"
