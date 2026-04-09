#!/bin/bash

# Haoyun Music Player - 发版脚本
# 用法: 
#   ./tag.sh              # 自动小版本号+1
#   ./tag.sh v0.6.0       # 指定版本号

set -e  # 遇到错误立即退出

# 确保在项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 检查 build/config.yml 是否存在
if [ ! -f "build/config.yml" ]; then
    echo "❌ 错误: 找不到 build/config.yml 文件"
    echo "当前目录: $(pwd)"
    exit 1
fi

version=$1

# 如果没有提供版本号，自动获取当前版本并小版本号+1
if [ -z "$version" ]; then
    echo "📖 从 build/config.yml 读取当前版本..."
    
    # 从 config.yml 中提取当前版本号
    current_version=$(grep -E '^\s*version:\s*"[0-9]+\.[0-9]+\.[0-9]+"' build/config.yml | head -1 | sed 's/.*version: *"\([0-9]*\.[0-9]*\.[0-9]*\)".*/\1/')
    
    if [ -z "$current_version" ]; then
        echo "❌ 错误: 无法从 build/config.yml 中解析版本号"
        exit 1
    fi
    
    echo "✅ 当前版本: $current_version"
    
    # 解析版本号的各个部分
    IFS='.' read -r major minor patch <<< "$current_version"
    
    # 小版本号 +1
    new_patch=$((patch + 1))
    new_version="${major}.${minor}.${new_patch}"
    version="v${new_version}"
    
    echo "🔢 新版本号: $version (小版本号自动+1)"
else
    # 验证提供的版本号格式
    if [[ ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "❌ 错误: 版本号格式不正确，应该是 v0.0.0 格式"
        echo "例如: v0.6.0, v1.0.0"
        exit 1
    fi
fi

echo ""
echo "🚀 开始发布版本: $version"

# 提取纯数字版本号 (去掉 v 前缀)
version_number=${version#v}
echo "📝 更新 build/config.yml 中的版本号为: $version_number"

# 使用 sed 更新 build/config.yml 中的版本号
# 注意：macOS 下的 sed -i 需要提供一个空字符串作为备份后缀
sed -i "s/version: \".*\" # The application version/version: \"$version_number\" # The application version/" build/config.yml
sed -i "s/const AppVersion = \".*\"/const AppVersion = \"$version_number\"/" main.go

echo "✅ 版本号已更新"

# 同步构建资源
echo "🔄 同步构建资源配置..."
wails3 task common:update:build-assets

echo "✅ 构建资源已同步"

# 提交更改
echo "📦 提交版本更新..."
git add build/config.yml build/ main.go
git commit -m "chore: bump version to $version_number"

# 推送代码到远程仓库
echo "🚀 推送代码到远程仓库..."
if ! git push; then
    echo ""
    echo "❌ 代码推送失败！"
    echo ""
    echo "可能的原因："
    echo "   1. 网络连接问题"
    echo "   2. 远程仓库有未合并的更改"
    echo "   3. 权限不足"
    echo ""
    echo "💡 解决步骤："
    echo "   1. 检查网络连接和 Git 配置"
    echo "   2. 如果有冲突，先拉取并合并：git pull --rebase origin main"
    echo "   3. 重新执行此脚本：./tag.sh ${version}"
    echo ""
    echo "或者手动执行："
    echo "   git push && git tag ${version} && git push origin ${version}"
    exit 1
fi

# 创建并推送 tag
echo "🏷️  创建 Git tag: $version"
git tag ${version}

echo "🚀 推送 tag 到远程仓库..."
if ! git push origin ${version}; then
    echo ""
    echo "❌ Tag 推送失败！"
    echo ""
    echo "可能的原因："
    echo "   1. 网络连接问题"
    echo "   2. Tag 已存在"
    echo "   3. 权限不足"
    echo ""
    echo "💡 解决步骤："
    echo "   1. 检查网络连接"
    echo "   2. 如果 tag 已存在且需要覆盖："
    echo "      git tag -d ${version} && git tag ${version} && git push --force origin ${version}"
    echo "   3. 或者删除远程 tag 后重新推送："
    echo "      git push --delete origin ${version} && git push origin ${version}"
    echo ""
    echo "⚠️  注意：代码已成功推送，但 tag 未推送"
    echo "   可以稍后手动推送：git push origin ${version}"
    exit 1
fi

echo ""
echo "🎉 版本 $version 发布成功！"
echo ""
echo "💡 提示:"
echo "   - GitHub Actions 将自动触发构建流程"
echo "   - 可以在 https://github.com/yhao521/haoyun-music-player/releases 查看发布状态"
echo ""
echo "如需撤销本次操作，可执行:"
echo "   git tag -d ${version} && git push --delete origin ${version} && git reset --hard HEAD~1"
