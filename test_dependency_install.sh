#!/bin/bash

# ============================================
# 依赖工具自动安装功能 - 快速测试脚本
# ============================================

echo "🚀 Haoyun Music Player - 依赖工具功能测试"
echo "=========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 检查项目目录
if [ ! -f "main.go" ]; then
    echo -e "${RED}❌ 错误: 请在项目根目录运行此脚本${NC}"
    exit 1
fi

echo "📁 项目目录: $(pwd)"
echo ""

# 步骤 1: 编译项目
echo "📦 步骤 1/4: 编译项目..."
echo "-------------------------------------------"
go build -o haoyun-music-player . 2>&1 | grep -E "error" 
if [ $? -eq 0 ]; then
    echo -e "${RED}❌ 编译失败${NC}"
    exit 1
else
    echo -e "${GREEN}✅ 编译成功${NC}"
fi
echo ""

# 步骤 2: 检查 FFmpeg 状态
echo "🔍 步骤 2/4: 检查 FFmpeg 状态..."
echo "-------------------------------------------"
if command -v ffmpeg &> /dev/null; then
    FFMPEG_VERSION=$(ffmpeg -version | head -n 1)
    echo -e "${GREEN}✅ FFmpeg 已安装${NC}"
    echo "   版本: $FFMPEG_VERSION"
else
    echo -e "${YELLOW}⚠️  FFmpeg 未安装${NC}"
    echo "   建议: brew install ffmpeg"
fi
echo ""

# 步骤 3: 显示托盘菜单预览
echo "📋 步骤 3/4: 托盘菜单结构预览..."
echo "-------------------------------------------"
cat << 'EOF'
托盘菜单将包含以下项：

🎵 [当前播放的歌曲名称]
──────────────
▶️ 播放/暂停 (Space)
⏮️ 上一曲
⏭️ 下一曲
──────────────
🎶 浏览音乐
❤️ 喜爱音乐
🔄 播放模式
📚 音乐库
🛠️ 依赖工具  ← 新增功能
   ├── ✅ FFmpeg (version x.x...)  [如果已安装]
   └── ❌ FFmpeg                    [如果未安装]
       └── 📦 安装 FFmpeg
   └── 🔄 重新检查所有工具
💾 下载音乐
💤 保持系统唤醒
🚀 开机启动
⚙️ 设置
🖥️ 显示主窗口
──────────────
ℹ️ 版本信息
🚪 退出
EOF
echo ""

# 步骤 4: 运行应用（可选）
echo "🎯 步骤 4/4: 准备启动应用..."
echo "-------------------------------------------"
read -p "是否现在启动应用进行测试？(y/n): " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo -e "${GREEN}🚀 启动应用中...${NC}"
    echo "提示: 查看系统托盘图标，点击可看到依赖工具菜单"
    echo ""
    echo "观察要点："
    echo "  1. 控制台输出依赖检测结果"
    echo "  2. 托盘菜单中的 '🛠️ 依赖工具' 项"
    echo "  3. 如果 FFmpeg 缺失，尝试点击安装"
    echo "  4. 安装完成后菜单应自动刷新"
    echo ""
    
    ./haoyun-music-player
else
    echo ""
    echo -e "${YELLOW}⏭️  跳过启动${NC}"
    echo ""
    echo "手动启动命令:"
    echo "  ./haoyun-music-player"
    echo ""
    echo "或使用开发模式:"
    echo "  wails3 dev"
fi

echo ""
echo "=========================================="
echo -e "${GREEN}✅ 测试准备完成${NC}"
echo ""
echo "📚 相关文档:"
echo "  - DEPENDENCY_AUTO_INSTALL.md     : 完整功能说明"
echo "  - DEPENDENCY_INSTALL_VERIFICATION.md : 详细测试清单"
echo "  - FFMPEG_GUIDE.md                : FFmpeg 安装指南"
echo ""
echo "🐛 遇到问题?"
echo "  1. 查看控制台日志获取详细信息"
echo "  2. 检查 FFmpeg 是否在 PATH 中: which ffmpeg"
echo "  3. 阅读故障排除文档"
echo ""