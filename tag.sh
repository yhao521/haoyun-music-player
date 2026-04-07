#!/bin/bash

# Haoyun Music Player - 发版脚本
# 用法: ./tag.sh v0.6.0

set -e  # 遇到错误立即退出

version=$1

if [ -z "$version" ]; then
    echo "❌ 错误: 请提供版本号"
    echo "用法: ./tag.sh v0.6.0"
    exit 1
fi

# 检查版本号格式 (应该以 v 开头)
if [[ ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "❌ 错误: 版本号格式不正确，应该是 v0.0.0 格式"
    echo "例如: v0.6.0, v1.0.0"
    exit 1
fi

echo "🚀 开始发布版本: $version"

# 提取纯数字版本号 (去掉 v 前缀)
version_number=${version#v}
echo "📝 更新 build/config.yml 中的版本号为: $version_number"

# 使用 sed 更新 build/config.yml 中的版本号
# 注意：macOS 下的 sed -i 需要提供一个空字符串作为备份后缀
sed -i "s/version: \".*\" # The application version/version: \"$version_number\" # The application version/" build/config.yml

echo "✅ 版本号已更新"

# 同步构建资源
echo "🔄 同步构建资源配置..."
wails3 task common:update:build-assets

echo "✅ 构建资源已同步"

# 提交更改
echo "📦 提交版本更新..."
git add build/config.yml build/
git commit -m "chore: bump version to $version_number"

# 创建并推送 tag
echo "🏷️  创建 Git tag: $version"
git tag ${version}
git push origin ${version}

echo ""
echo "🎉 版本 $version 发布成功！"
echo ""
echo "💡 提示:"
echo "   - GitHub Actions 将自动触发构建流程"
echo "   - 可以在 https://github.com/yhao521/haoyun-music-player/releases 查看发布状态"
echo ""
echo "如需撤销本次操作，可执行:"
echo "   git tag -d ${version}"
echo "   git push --delete origin ${version}"
echo "   git reset --hard HEAD~1"