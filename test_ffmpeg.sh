#!/bin/bash

echo "=========================================="
echo "  Haoyun Music Player - FFmpeg 测试工具"
echo "=========================================="
echo ""

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未找到 Go，请先安装 Go 1.25+"
    exit 1
fi

echo "✅ Go 版本: $(go version)"
echo ""

# 检查 FFmpeg 是否安装
if ! command -v ffmpeg &> /dev/null; then
    echo "⚠️  警告: 未找到 FFmpeg"
    echo ""
    echo "请安装 FFmpeg:"
    echo "  macOS:   brew install ffmpeg"
    echo "  Ubuntu:  sudo apt-get install ffmpeg"
    echo "  Fedora:  sudo dnf install ffmpeg"
    echo ""
    read -p "是否继续运行测试？(y/n) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    echo "✅ FFmpeg 版本:"
    ffmpeg -version | head -n 1
    echo ""
fi

# 进入 tests 目录
cd "$(dirname "$0")/tests" || exit 1

# 编译测试程序
echo "🔨 编译测试程序..."
go build -o test_ffmpeg_bin test_ffmpeg.go

if [ $? -ne 0 ]; then
    echo "❌ 编译失败"
    exit 1
fi

echo "✅ 编译成功"
echo ""

# 运行测试
echo "🚀 运行测试..."
echo ""
./test_ffmpeg_bin

# 清理
rm -f test_ffmpeg_bin

echo ""
echo "=========================================="
echo "  测试完成"
echo "=========================================="