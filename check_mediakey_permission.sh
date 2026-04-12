#!/bin/bash

# macOS 媒体键权限诊断和修复脚本

echo "=========================================="
echo "🎵 Haoyun Music Player - 媒体键权限诊断"
echo "=========================================="
echo ""

# 检查辅助功能是否启用
echo "1️⃣  检查辅助功能状态..."
ACCESSIBILITY_ENABLED=$(osascript -e 'tell application "System Events" to get UI elements enabled' 2>&1)
if [ "$ACCESSIBILITY_ENABLED" = "true" ]; then
    echo "   ✅ 辅助功能已启用"
else
    echo "   ❌ 辅助功能未启用"
    echo ""
    echo "   💡 需要授予辅助功能权限才能让媒体键工作"
fi
echo ""

# 查找应用路径
DEV_APP_PATH="bin/haoyun-music-player.dev.app"
PROD_APP_PATH="/Applications/haoyun-music-player.app"

echo "2️⃣  检查应用安装状态..."
if [ -d "$DEV_APP_PATH" ]; then
    DEV_APP_ABSOLUTE=$(cd "$(dirname "$DEV_APP_PATH")" && pwd)/$(basename "$DEV_APP_PATH")
    echo "   ✅ 开发版本应用: $DEV_APP_ABSOLUTE"
else
    echo "   ⚠️  开发版本应用未找到"
fi

if [ -d "$PROD_APP_PATH" ]; then
    echo "   ✅ 生产版本应用: $PROD_APP_PATH"
else
    echo "   ⚠️  生产版本应用未找到"
fi
echo ""

# 获取 Bundle ID
echo "3️⃣  应用 Bundle ID 信息..."
if [ -f "$DEV_APP_PATH/Contents/Info.plist" ]; then
    BUNDLE_ID_DEV=$(/usr/libexec/PlistBuddy -c "Print :CFBundleIdentifier" "$DEV_APP_PATH/Contents/Info.plist" 2>/dev/null || echo "未知")
    echo "   开发版本: $BUNDLE_ID_DEV"
fi

if [ -f "$PROD_APP_PATH/Contents/Info.plist" ]; then
    BUNDLE_ID_PROD=$(/usr/libexec/PlistBuddy -c "Print :CFBundleIdentifier" "$PROD_APP_PATH/Contents/Info.plist" 2>/dev/null || echo "未知")
    echo "   生产版本: $BUNDLE_ID_PROD"
fi
echo ""

# 提供详细的修复步骤
echo "=========================================="
echo "🔧 修复步骤（请按顺序执行）"
echo "=========================================="
echo ""
echo "第 1 步：打开系统偏好设置"
echo "   👉 点击下面的链接直接跳转："
echo "   x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility"
echo ""
read -p "   按回车键打开系统偏好设置..." 
open "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility"
echo ""

echo "第 2 步：添加应用到辅助功能列表"
echo "   1. 点击左下角的 🔒 锁图标解锁（需要输入密码）"
echo "   2. 点击 '+' 号添加应用"
echo "   3. 选择以下应用之一："
if [ ! -z "$DEV_APP_ABSOLUTE" ]; then
    echo "      • 开发版本: $DEV_APP_ABSOLUTE"
fi
if [ -d "$PROD_APP_PATH" ]; then
    echo "      • 生产版本: $PROD_APP_PATH"
fi
echo "   4. 确保应用前面的复选框已勾选 ✓"
echo ""
read -p "   完成后按回车键继续..."

echo ""
echo "第 3 步：完全退出当前应用"
echo "   • 按 Cmd+Q 完全退出应用"
echo "   • 或者运行: pkill -f haoyun-music-player"
echo ""
read -p "   应用已退出？按回车键继续..."

echo ""
echo "第 4 步：重新启动应用"
echo "   • 重新运行: wails3 dev -config ./build/config.yml"
echo ""
read -p "   应用已启动？按回车键继续测试..."

echo ""
echo "=========================================="
echo "🧪 测试媒体键"
echo "=========================================="
echo ""
echo "请在接下来 10 秒内按下以下按键之一："
echo "   • F7 (上一曲)"
echo "   • F8 (播放/暂停)"
echo "   • F9 (下一曲)"
echo ""

sleep 10

echo ""
echo "=========================================="
echo "📊 检查结果"
echo "=========================================="
echo ""

LOG_FILE="$HOME/.haoyun-music/runtime/logs/app-$(date +%Y%m%d).log"

if [ -f "$LOG_FILE" ]; then
    MEDIA_KEY_EVENTS=$(tail -50 "$LOG_FILE" | grep -E "CGO callback|media key|Play/Pause|Next Track|Previous Track" | wc -l)
    
    if [ "$MEDIA_KEY_EVENTS" -gt 0 ]; then
        echo "✅ 成功！检测到媒体键事件："
        tail -50 "$LOG_FILE" | grep -E "CGO callback|media key|Play/Pause|Next Track|Previous Track" | tail -5
        echo ""
        echo "🎉 媒体键已成功配置！"
    else
        echo "❌ 未检测到媒体键事件"
        echo ""
        echo "可能的原因："
        echo "   1. 权限尚未生效（尝试重启系统）"
        echo "   2. 添加了错误的应用路径"
        echo "   3. 应用未完全重启"
        echo ""
        echo "💡 建议："
        echo "   • 确认添加的是当前运行的应用（开发版或生产版）"
        echo "   • 完全退出应用后重新启动"
        echo "   • 查看完整日志: tail -f $LOG_FILE"
    fi
else
    echo "⚠️  日志文件未找到: $LOG_FILE"
    echo "   请确保应用正在运行"
fi

echo ""
echo "=========================================="
echo "📖 更多信息"
echo "=========================================="
echo ""
echo "实时查看日志: tail -f $LOG_FILE | grep -E 'CGO|media'"
echo "查看文档: cat MEDIAKEY_DEBUG_GUIDE.md"
echo ""
