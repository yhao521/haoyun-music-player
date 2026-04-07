# GitHub Actions 部署检查清单

## 📋 部署前检查

### ✅ 代码准备
- [ ] 所有代码更改已提交
- [ ] `go.mod` 和 `go.sum` 已是最新
- [ ] `frontend/package-lock.json` 已是最新
- [ ] 本地测试打包成功 (`task package-all`)
- [ ] README.md 已更新（如需要）

### ✅ 仓库配置
- [ ] 代码已推送到 GitHub
- [ ] GitHub Actions 已启用 (Settings → Actions → General)
- [ ] 工作流文件已添加: `.github/workflows/release.yml`
- [ ] 分支保护规则已配置（如需要）

### ✅ 依赖检查
- [ ] Go 版本兼容 (当前: 1.25)
- [ ] Node.js 版本兼容 (当前: 20)
- [ ] Wails CLI 可安装
- [ ] Task 工具可用

### ✅ 权限检查
- [ ] 有推送标签的权限
- [ ] 有创建 Release 的权限
- [ ] GITHUB_TOKEN 自动可用（无需配置）

## 🚀 首次部署步骤

### 第 1 步: 提交工作流文件
```bash
git add .github/workflows/release.yml
git add GITHUB_ACTIONS_*.md
git commit -m "Add GitHub Actions for automated releases"
git push origin main
```

### 第 2 步: 验证工作流已同步
1. 访问: `https://github.com/用户名/haoyun-music-player/actions`
2. 确认看到 "Release Build" 工作流
3. 如果没有，等待几分钟或刷新页面

### 第 3 步: 测试手动触发
1. 在 Actions 页面点击 "Release Build"
2. 点击 "Run workflow" 按钮
3. 选择 `main` 分支
4. 点击绿色 "Run workflow" 按钮
5. 观察构建过程（应该看到黄色运行中状态）

### 第 4 步: 验证手动构建
等待构建完成后检查：
- [ ] 构建状态为绿色（成功）
- [ ] 可以在 Artifacts 下载产物（如果不是 tag 触发）
- [ ] 没有错误日志

### 第 5 步: 正式发布
```bash
# 创建发布标签
git tag v0.5.0
git push origin v0.5.0
```

### 第 6 步: 监控自动构建
1. 访问 Actions 页面
2. 应该看到新的运行记录（由 tag 触发）
3. 点击查看详情
4. 等待两个平台都构建完成（约 15-20 分钟）

### 第 7 步: 验证 Release
构建完成后检查：
- [ ] Releases 页面出现新版本
- [ ] 所有平台产物已上传（5 个文件）
- [ ] Release Notes 已生成
- [ ] 可以下载所有文件

## 🔍 详细验证清单

### 工作流文件验证
```bash
# 检查工作流文件语法
cat .github/workflows/release.yml

# 确认包含以下关键配置:
# ✓ on.push.tags: 'v*'
# ✓ permissions.contents: write
# ✓ matrix.os: [macos-latest, windows-latest]
# ✓ softprops/action-gh-release@v2
```

### 本地环境验证
```bash
# 测试 Go 环境
go version
go mod tidy

# 测试 Node.js 环境
cd frontend
npm install
npm run build
cd ..

# 测试 Task
task --version

# 测试 Wails
wails3 doctor

# 测试打包（可选，耗时较长）
task darwin:package
```

### GitHub 配置验证
```bash
# 检查远程仓库
git remote -v

# 检查标签
git tag -l

# 检查分支
git branch -a
```

## ⚠️ 常见问题预检查

### 问题 1: Actions 未显示
**检查**:
- [ ] 文件路径正确: `.github/workflows/release.yml`
- [ ] YAML 语法正确（无缩进错误）
- [ ] 已推送到正确的分支

**解决**:
```bash
# 验证文件存在
ls -la .github/workflows/release.yml

# 验证 YAML 语法
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/release.yml'))"
```

### 问题 2: 构建失败 - 依赖问题
**预检查**:
- [ ] `go mod tidy` 无错误
- [ ] `npm install` 无错误
- [ ] 所有依赖已提交到 Git

**解决**:
```bash
# 清理并重新安装
rm -rf node_modules
rm package-lock.json
npm install
git add package-lock.json

go mod tidy
git add go.mod go.sum

git commit -m "Fix dependencies"
git push
```

### 问题 3: 标签推送未触发
**检查**:
- [ ] 标签格式正确 (v*)
- [ ] 标签已推送到远程
- [ ] 工作流已启用

**解决**:
```bash
# 验证标签
git tag -l "v*"

# 查看远程标签
git ls-remote --tags origin

# 重新推送
git push origin v0.5.0 --force
```

### 问题 4: Release 未创建
**检查**:
- [ ] 是从 tag 触发的（不是手动触发）
- [ ] 构建成功（绿色状态）
- [ ] 工作流中有上传步骤

**解决**:
- 检查工作流中的条件: `if: startsWith(github.ref, 'refs/tags/')`
- 确认使用的是 tag 推送，不是手动触发

## 📊 构建时间参考

| 阶段 | 预计时间 | 说明 |
|------|---------|------|
| 队列等待 | 0-2 min | 取决于 GitHub 负载 |
| 环境准备 | 2-3 min | 安装 Go, Node.js, Task |
| 依赖安装 | 3-5 min | Go modules + npm packages |
| 前端构建 | 2-3 min | Vite build |
| Go 编译 | 3-5 min | CGO enabled |
| 打包 | 1-2 min | 创建 DMG/ZIP |
| 上传 | 1-2 min | 上传到 Release |
| **总计** | **12-20 min** | 两平台并行 |

## 🎯 成功标志

### 构建成功
```
✅ Actions 页面显示绿色对勾
✅ 所有步骤都显示完成
✅ 无错误或警告
✅ 日志末尾显示 "Upload artifacts to GitHub Release" 成功
```

### Release 创建成功
```
✅ Releases 页面出现新版本
✅ 版本号为推送的标签（如 v0.5.0）
✅ 包含 5 个构建产物
✅ Release Notes 自动生成
✅ 标记为 "Latest release"（如果是最新版本）
```

### 产物完整性
```
✅ haoyun-music-player_vX.X.X_TIMESTAMP_macOS_arm64.dmg
✅ haoyun-music-player_vX.X.X_TIMESTAMP_macOS_amd64.dmg
✅ haoyun-music-player_vX.X.X_TIMESTAMP_macOS_universal.zip
✅ haoyun-music-player_vX.X.X_TIMESTAMP_Windows_arm64.zip
✅ haoyun-music-player_vX.X.X_TIMESTAMP_Windows_amd64.zip
```

## 🔄 日常发布流程

### 标准发布
```bash
# 1. 准备代码
git add .
git commit -m "Prepare for release"
git push origin main

# 2. 打标签
git tag v0.5.0
git push origin v0.5.0

# 3. 等待构建（15-20 分钟）
# 访问: https://github.com/用户名/haoyun-music-player/actions

# 4. 验证 Release
# 访问: https://github.com/用户名/haoyun-music-player/releases

# 5. 通知用户（可选）
# 发送更新通知、更新文档等
```

### 紧急修复
```bash
# 1. 修复问题
git add .
git commit -m "Fix critical bug"
git push origin main

# 2. 创建补丁版本
git tag v0.5.1
git push origin v0.5.1

# 3. 监控构建
# 同上
```

### 预发布版本
修改工作流文件，将 `prerelease: false` 改为 `true`，然后：
```bash
git tag v0.6.0-beta.1
git push origin v0.6.0-beta.1
```

## 📝 维护任务

### 定期维护（每月）
- [ ] 检查 Go 版本是否需要更新
- [ ] 检查 Node.js 版本是否需要更新
- [ ] 清理旧的 Draft Releases
- [ ] 审查工作流依赖的版本

### 版本更新时
- [ ] 更新 `build/config.yml` 中的版本号
- [ ] 运行 `wails3 task common:update:build-assets`
- [ ] 提交所有更改
- [ ] 创建新标签

### 工作流更新时
- [ ] 在本地测试 YAML 语法
- [ ] 使用手动触发测试
- [ ] 确认向后兼容
- [ ] 更新相关文档

## 🆘 获取帮助

### 日志查看
```bash
# 通过 GitHub CLI
gh run list --workflow=release.yml
gh run view <run-id>

# 查看详细日志
gh run view <run-id> --log
```

### 调试技巧
1. 添加 `echo` 语句输出变量值
2. 使用 `set -x` 启用 bash 调试
3. 分步注释排查问题
4. 使用手动触发快速迭代

### 社区资源
- [GitHub Actions 文档](https://docs.github.com/en/actions)
- [Wails Discord](https://discord.gg/wails)
- [Stack Overflow - github-actions](https://stackoverflow.com/questions/tagged/github-actions)

---

**最后更新**: 2026-04-07  
**维护者**: yhao521  
**版本**: 1.0

**提示**: 打印此清单，每次发布前逐项检查！
