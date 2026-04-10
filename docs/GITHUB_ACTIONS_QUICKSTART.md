# GitHub Actions 快速发布指南

## 🚀 快速开始（3 步发布新版本）

```bash
# 1️⃣ 提交所有更改
git add .
git commit -m "准备发布新版本"
git push origin main

# 2️⃣ 创建并推送标签（自动触发构建）
git tag v0.5.0
git push origin v0.5.0

# 3️⃣ 等待 GitHub Actions 完成（约 10-20 分钟）
# 访问: https://github.com/yourusername/haoyun-music-player/actions
```

## 📦 构建产物

完成后，在 **Releases** 页面会自动创建新版本，包含：

### macOS
- ✅ `*_macOS_arm64.dmg` - Apple Silicon (M1/M2/M3)
- ✅ `*_macOS_amd64.dmg` - Intel Mac
- ✅ `*_macOS_universal.zip` - 通用版本

### Windows
- ✅ `*_Windows_arm64.zip` - ARM64 Windows
- ✅ `*_Windows_amd64.zip` - x64 Windows

## 🔍 查看构建进度

1. 打开 GitHub 仓库
2. 点击 **Actions** 标签
3. 查看 **Release Build** 工作流状态
   - 🟡 黄色 = 构建中
   - 🟢 绿色 = 成功
   - 🔴 红色 = 失败（点击查看详情）

## ⚠️ 常见问题

### Q: 如何手动触发构建？
A: 在 Actions 页面点击 **Run workflow** 按钮

### Q: 构建失败了怎么办？
A: 
1. 点击失败的运行记录查看详细日志
2. 检查是否是依赖问题
3. 本地测试打包：`task package-all`

### Q: 可以只构建某个平台吗？
A: 当前配置会同时构建所有平台。如需单独构建，需要修改工作流文件。

### Q: 版本号是如何生成的？
A: 
- 格式：`v0.0.X_时间戳`
- X = Git 提交次数
- 时间戳 = YYYYMMDDHHmmSS

### Q: 如何标记为预发布版本？
A: 编辑 `.github/workflows/release.yml`，将 `prerelease: false` 改为 `true`

## 🛠️ 首次使用检查清单

- [ ] 代码已推送到 GitHub
- [ ] 本地测试过 `task darwin:package`（macOS）
- [ ] 确认 `go.mod` 和 `package-lock.json` 已提交
- [ ] GitHub Actions 已启用（Settings → Actions → General）

## 📝 示例：发布 v0.5.0

```bash
# 确保在 main 分支
git checkout main
git pull

# 提交最后更改
git add .
git commit -m "Release v0.5.0"
git push

# 打标签并推送
git tag v0.5.0
git push origin v0.5.0

# 完成！等待 CI/CD 自动构建和发布
```

## 🔗 相关链接

- 查看构建状态: `https://github.com/用户名/haoyun-music-player/actions`
- 查看 Releases: `https://github.com/用户名/haoyun-music-player/releases`
- 详细文档: [GITHUB_ACTIONS_RELEASE.md](./GITHUB_ACTIONS_RELEASE.md)

---

**提示**: 第一次推送标签后，可能需要几分钟才能看到 Release 创建完成。
