#!/bin/bash

# 测试系统媒体键和全局快捷键

echo "======================================"
echo "  Haoyun Music Player - 快捷键测试"
echo "======================================"
echo ""

echo "📋 已注册的快捷键:"
echo ""
echo "【系统媒体键】(需要辅助功能权限)"
echo "  F7 (⏮️)  - 上一曲"
echo "  F8 (⏯️)  - 播放/暂停"
echo "  F9 (⏭️)  - 下一曲"
echo ""
echo "【自定义全局快捷键】"
echo "  Ctrl+Shift+P - 播放/暂停"
echo "  Ctrl+Shift+N - 下一曲"
echo "  Ctrl+Shift+B - 上一曲"
echo "  Ctrl+Shift+↑ - 音量增加"
echo "  Ctrl+Shift+↓ - 音量减少"
echo ""

echo "🔧 检查辅助功能权限..."
# 使用正确的 tccutil 命令检查权限
if tccutil reset Accessibility 2>&1 | grep -q "Usage"; then
    echo "⚠️  无法自动检测权限状态"
    echo ""
    echo "💡 如何检查和授予权限:"
    echo "  1. 打开 系统偏好设置 > 安全性与隐私 > 隐私"
    echo "  2. 选择左侧的 '辅助功能'"
    echo "  3. 查看列表中是否有 haoyun-music-player"
    echo "  4. 如果没有,点击 '+' 添加应用并勾选"
    echo "  5. 重启应用使权限生效"
else
    echo "✅ 辅助功能权限已授予"
fi

echo ""
echo "📝 日志文件位置:"
echo "  ~/.haoyun-music/runtime/logs/app-YYYYMMDD.log"
echo ""
echo "🎵 开始测试吧!"
echo "  1. 启动应用: ./haoyun-music-player"
echo "  2. 添加一些音乐到播放列表"
echo "  3. 尝试按下上述快捷键"
echo "  4. 查看日志确认按键是否被正确捕获"
echo ""
echo "======================================"

# macOS 媒体键功能测试脚本
# 用法: ./test_media_keys.sh

echo "=========================================="
echo "  Haoyun Music Player - 媒体键功能测试"
echo "=========================================="
echo ""

# 检查是否在 macOS 上
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo "❌ 错误: 此脚本仅适用于 macOS 系统"
    exit 1
fi

echo "✅ 检测到 macOS 系统"
echo ""

# 检查 Go 环境
echo "📦 检查 Go 环境..."
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未找到 Go,请先安装 Go"
    exit 1
fi
echo "   Go 版本: $(go version)"
echo ""

# 检查 Wails
echo "🔧 检查 Wails..."
if ! command -v wails3 &> /dev/null; then
    echo "⚠️  警告: 未找到 wails3 命令"
    echo "   请确保已安装 Wails v3"
else
    echo "   Wails 已安装"
fi
echo ""

# 检查 CGO 支持
echo "🔨 检查 CGO 支持..."
CGO_ENABLED=$(go env CGO_ENABLED)
if [ "$CGO_ENABLED" != "1" ]; then
    echo "⚠️  警告: CGO 未启用 (CGO_ENABLED=$CGO_ENABLED)"
    echo "   尝试启用 CGO..."
    export CGO_ENABLED=1
else
    echo "   ✅ CGO 已启用"
fi
echo ""

# 检查 Xcode Command Line Tools
echo "🛠️  检查 Xcode Command Line Tools..."
if ! xcode-select -p &> /dev/null; then
    echo "⚠️  警告: 未安装 Xcode Command Line Tools"
    echo "   运行以下命令安装:"
    echo "   xcode-select --install"
else
    echo "   ✅ Xcode Command Line Tools 已安装"
fi
echo ""

# 清理旧的构建缓存
echo "🧹 清理构建缓存..."
go clean -cache
echo "   ✅ 缓存已清理"
echo ""

# 编译应用
echo "🏗️  编译应用..."
cd "$(dirname "$0")"
if wails3 build -config ./build/config.yml 2>&1 | tee /tmp/wails_build.log; then
    echo "   ✅ 编译成功"
else
    echo "   ❌ 编译失败,查看日志: /tmp/wails_build.log"
    exit 1
fi
echo ""

# 检查二进制文件
BINARY_PATH="./build/bin/Haoyun Music Player"
if [ ! -f "$BINARY_PATH" ]; then
    echo "❌ 错误: 未找到编译后的二进制文件"
    exit 1
fi
echo "✅ 二进制文件已生成: $BINARY_PATH"
echo ""

# 提示用户测试
echo "=========================================="
echo "  🎯 测试步骤"
echo "=========================================="
echo ""
echo "1. 启动应用:"
echo "   $BINARY_PATH"
echo ""
echo "2. 观察终端日志,应看到:"
echo "   🍎 正在注册 macOS 媒体键..."
echo "   ✅ macOS 媒体键注册成功"
echo "   📝 支持的按键: 播放/暂停(F8), 下一曲(F9), 上一曲(F7)"
echo ""
echo "3. 测试媒体键:"
echo "   • 按 F8 (或 Fn+F8): 播放/暂停"
echo "   • 按 F9 (或 Fn+F9): 下一曲"
echo "   • 按 F7 (或 Fn+F7): 上一曲"
echo ""
echo "4. 验证响应,日志应显示:"
echo "   ▶️⏸️  收到媒体键:播放/暂停"
echo "   ⏭️  收到媒体键:下一曲"
echo "   ⏮️  收到媒体键:上一曲"
echo ""
echo "5. 如果媒体键无响应:"
echo "   • 检查辅助功能权限:"
echo "     系统偏好设置 > 安全性与隐私 > 隐私 > 辅助功能"
echo "   • 确保 'Haoyun Music Player' 已勾选"
echo ""
echo "=========================================="
echo "  💡 提示"
echo "=========================================="
echo ""
echo "• 媒体键无需应用获得焦点即可响应"
echo "• 支持后台运行时的媒体控制"
echo "• 与托盘菜单、快捷键协同工作"
echo ""
echo "祝测试顺利! 🎉"
echo ""
