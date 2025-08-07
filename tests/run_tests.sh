#!/bin/bash

# 测试运行脚本
set -e

echo "🧪 开始运行单元测试..."

# 设置测试环境变量
export GO_ENV=test

# 进入项目根目录
cd "$(dirname "$0")/.."

echo "📋 运行所有测试..."
go test -v ./tests/... -count=1

echo ""
echo "📊 运行测试覆盖率分析..."
go test -v ./tests/... -coverprofile=coverage.out -covermode=atomic

echo ""
echo "📈 生成覆盖率报告..."
go tool cover -html=coverage.out -o coverage.html

echo ""
echo "⚡ 运行性能测试..."
go test -v ./tests/... -bench=. -benchmem -run=^$ -count=1

echo ""
echo "🎯 运行竞态条件检测..."
go test -v ./tests/... -race -count=1

echo ""
echo "✅ 所有测试完成！"
echo "📄 覆盖率报告已生成: coverage.html"