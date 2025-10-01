#!/bin/bash
# PostgreSQL 数据库测试脚本

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}PostgreSQL 数据库测试脚本${NC}"
echo -e "${GREEN}================================${NC}"
echo

# 检查是否提供了密码
if [ -z "$PGPASSWORD" ]; then
    echo -e "${YELLOW}请设置 PGPASSWORD 环境变量${NC}"
    echo "例如: export PGPASSWORD='your_password'"
    exit 1
fi

# 配置变量
PGHOST=${PGHOST:-localhost}
PGPORT=${PGPORT:-5432}
PGUSER=${PGUSER:-pnas}
PGDATABASE=${PGDATABASE:-pnas}

echo -e "${YELLOW}测试 PostgreSQL 连接${NC}"
echo "主机: $PGHOST"
echo "端口: $PGPORT"
echo "用户: $PGUSER"
echo "数据库: $PGDATABASE"
echo

# 测试 1: 检查 PostgreSQL 是否运行
echo -e "${YELLOW}测试 1: 检查 PostgreSQL 服务${NC}"
if command -v systemctl &> /dev/null; then
    sudo systemctl status postgresql --no-pager | head -5
elif command -v docker &> /dev/null; then
    docker ps | grep postgres || echo "No PostgreSQL container found"
fi
echo

# 测试 2: 测试连接
echo -e "${YELLOW}测试 2: 测试数据库连接${NC}"
if psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -c "SELECT version();" &> /dev/null; then
    echo -e "${GREEN}✓ 连接成功${NC}"
    psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -c "SELECT version();" -t | head -1
else
    echo -e "${RED}✗ 连接失败${NC}"
    exit 1
fi
echo

# 测试 3: 检查数据库是否存在
echo -e "${YELLOW}测试 3: 检查数据库${NC}"
if psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -c "\l" &> /dev/null; then
    echo -e "${GREEN}✓ 数据库存在${NC}"
else
    echo -e "${RED}✗ 数据库不存在${NC}"
    echo "请先创建数据库："
    echo "  CREATE DATABASE $PGDATABASE OWNER $PGUSER;"
    exit 1
fi
echo

# 测试 4: 使用 PNAS 配置文件测试
echo -e "${YELLOW}测试 4: 生成 PNAS 配置文件${NC}"
cat > /tmp/pnas-test-config.yaml <<EOF
server:
  port: "8080"
  debug: false

database:
  type: "postgres"
  host: "$PGHOST"
  port: $PGPORT
  user: "$PGUSER"
  password: "$PGPASSWORD"
  dbname: "$PGDATABASE"
  sslmode: "disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300

jwt:
  secret: "test-secret-key-change-this"
  expire: 24

audit:
  enabled: false
EOF

echo -e "${GREEN}✓ 配置文件已生成: /tmp/pnas-test-config.yaml${NC}"
cat /tmp/pnas-test-config.yaml
echo

# 测试 5: 运行迁移（如果 PNAS 可执行文件存在）
if [ -f "./build/pnas" ]; then
    echo -e "${YELLOW}测试 5: 运行数据库迁移${NC}"
    export DATABASE_TYPE=postgres
    export DATABASE_HOST=$PGHOST
    export DATABASE_PORT=$PGPORT
    export DATABASE_USER=$PGUSER
    export DATABASE_PASSWORD=$PGPASSWORD
    export DATABASE_DBNAME=$PGDATABASE

    if ./build/pnas migrate; then
        echo -e "${GREEN}✓ 迁移成功${NC}"
    else
        echo -e "${RED}✗ 迁移失败${NC}"
        exit 1
    fi
    echo

    # 测试 6: 查看创建的表
    echo -e "${YELLOW}测试 6: 检查数据库表${NC}"
    psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -c "\dt"
    echo
else
    echo -e "${YELLOW}跳过迁移测试（pnas 可执行文件不存在）${NC}"
    echo "请先编译: make build"
    echo
fi

# 测试 7: 性能测试
echo -e "${YELLOW}测试 7: 简单性能测试${NC}"
psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" <<EOF
-- 创建测试表
CREATE TABLE IF NOT EXISTS test_performance (
    id SERIAL PRIMARY KEY,
    data TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 插入测试数据
INSERT INTO test_performance (data)
SELECT 'test data ' || generate_series(1, 1000);

-- 查询测试
\timing on
SELECT COUNT(*) FROM test_performance;

-- 清理
DROP TABLE test_performance;
EOF
echo

echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}所有测试完成！${NC}"
echo -e "${GREEN}================================${NC}"
echo
echo "下一步："
echo "1. 复制配置文件: cp /tmp/pnas-test-config.yaml config.yaml"
echo "2. 启动 PNAS: ./build/pnas server"
echo
