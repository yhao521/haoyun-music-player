#!/bin/bash

# Haoyun Music Player - 快速启动脚本

set -e

echo "🎵 Haoyun Music Player - 快速启动"
echo "=================================="

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    echo "❌ 错误：未找到 Go，请确保已安装 Go 1.25+"
    exit 1
fi

echo "✅ Go 版本：$(go version)"

# 检查 Node.js 是否安装
if ! command -v node &> /dev/null; then
    echo "❌ 错误：未找到 Node.js，请确保已安装 Node.js 18+"
    exit 1
fi

echo "✅ Node.js 版本：$(node --version)"

# 检查 Wails 是否安装
if ! command -v wails3 &> /dev/null && ! command -v wails &> /dev/null; then
    echo "⚠️  警告：未找到 Wails CLI"
    echo "📦 请先安装 Wails v3:"
    echo "   go install github.com/wailsapp/wails/v3/cmd/wails@latest"
    echo ""
fi

echo ""
echo "📦 安装依赖..."

# 安装 Go 依赖
echo "   → 安装 Go 依赖..."
go mod tidy

# 安装前端依赖
echo "   → 安装前端依赖..."
cd frontend
npm install
cd ..

echo ""
echo "✅ 依赖安装完成!"
echo ""
echo "🚀 启动开发模式..."
echo ""
echo "   如果已安装 Wails v3，请运行:"
echo "   wails3 dev -config ./build/config.yml"
echo ""
echo "   或者查看 QUICKSTART.md 了解更多选项"
echo ""
echo "🎵 Happy Coding!"
