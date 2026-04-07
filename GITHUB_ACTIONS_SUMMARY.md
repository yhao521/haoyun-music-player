# GitHub Actions 自动化打包发布 - 实施总结

## ✅ 已完成的工作

### 1. 创建的文件

#### `.github/workflows/release.yml`
GitHub Actions 工作流配置文件，包含：
- **触发条件**: 推送 `v*` 标签或手动触发
- **构建平台**: macOS (Apple Silicon) 和 Windows
- **构建步骤**: 
  - 环境设置（Go、Node.js、Task、Wails CLI）
  - 依赖安装
  - 执行 `task package-all` 打包
  - 自动上传到 GitHub Release
- **权限配置**: `contents: write` 用于创建 Release

#### `GITHUB_ACTIONS_RELEASE.md`
详细的使用指南文档，包含：
- 触发方式说明
- 构建产物清单
- 工作流程详解
- 故障排查指南
- 自定义配置方法
- 最佳实践建议

#### `GITHUB_ACTIONS_QUICKSTART.md`
快速参考文档，包含：
- 3 步快速发布流程
- 常见问题解答
- 首次使用检查清单
- 实际示例演示

### 2. 工作流特性

#### 🎯 自动化程度
- ✅ 自动检测标签推送
- ✅ 自动安装所有依赖
- ✅ 自动多平台并行构建
- ✅ 自动创建 GitHub Release
- ✅ 自动上传构建产物

#### 📦 支持的 platform
| 平台 | 架构 | 格式 | 状态 |
|------|------|------|------|
| macOS | arm64 | .dmg | ✅ |
| macOS | amd64 | .dmg | ✅ |
| macOS | universal | .zip | ✅ |
| Windows | arm64 | .zip | ✅ |
| Windows | amd64 | .zip | ✅ |
| Linux | arm64/amd64 | .zip | ⏸️ (已注释) |

#### 🚀 性能优化
- Go 模块缓存
- npm 依赖缓存
- 多平台并行构建
- 增量构建支持

## 📋 使用步骤

### 首次使用

1. **提交工作流文件**
   ```bash
   git add .github/workflows/release.yml
   git add GITHUB_ACTIONS_RELEASE.md
   git add GITHUB_ACTIONS_QUICKSTART.md
   git commit -m "Add GitHub Actions for automated releases"
   git push origin main
   ```

2. **启用 GitHub Actions**
   - 访问仓库 Settings → Actions → General
   - 确保 Actions 已启用

3. **测试手动触发**
   - 进入 Actions 标签页
   - 选择 "Release Build"
   - 点击 "Run workflow"

### 正式发布

```bash
# 1. 准备发布
git add .
git commit -m "Prepare release vX.Y.Z"
git push origin main

# 2. 创建标签（自动触发构建）
git tag v0.5.0
git push origin v0.5.0

# 3. 等待构建完成（10-20 分钟）
# 查看进度: https://github.com/用户名/haoyun-music-player/actions

# 4. 检查 Release
# 访问: https://github.com/用户名/haoyun-music-player/releases
```

## 🔧 技术细节

### 版本号生成
```yaml
VERSION_1.VERSION_2 = 0.0 (来自 Taskfile.yml)
VERSION_3 = git rev-list HEAD --no-merges --count (提交次数)
TIMESTAMP = date +"%Y%m%d%H%M%S" (构建时间)

最终格式: haoyun-music-player_v0.0.X_YYYYMMDDHHmmSS_platform_arch.ext
```

### 构建环境
- **macOS**: `macos-latest` (Apple Silicon M1/M2)
- **Windows**: `windows-latest` (x64)
- **Go**: 从 `go.mod` 自动读取（当前 1.25）
- **Node.js**: v20
- **Wails**: 最新 alpha 版本

### 依赖管理
- Task 工具：通过 `arduino/setup-task@v2` 安装
- Wails CLI：通过 `go install` 安装
- 系统依赖：
  - macOS: `create-dmg`, `zip` (通过 Homebrew)
  - Windows: 内置 PowerShell 和 zip

## ⚠️ 注意事项

### 1. 标签格式
必须使用 `v` 前缀：
- ✅ `v0.5.0`
- ✅ `v1.0.0-beta.1`
- ❌ `0.5.0`
- ❌ `release-0.5.0`

### 2. 构建时间
- 首次构建：约 15-20 分钟（需要下载依赖）
- 后续构建：约 10-15 分钟（有缓存）

### 3. 并发限制
- GitHub 免费账户：最多 20 个并行 job
- 本项目使用 2 个 job（macOS + Windows），不会超限

### 4. 存储空间
- 每个 Release 的产物约 200-500 MB
- 建议定期清理旧的 Draft Releases

### 5. Linux 支持
当前 Linux 构建已在 `Taskfile.yml` 中注释，如需启用：
1. 取消 `package-all` 中 Linux 相关行的注释
2. 在工作流中添加 `ubuntu-latest` runner
3. 安装 GTK/WebKit 等系统依赖

## 🐛 故障排查

### 常见问题及解决方案

#### 问题 1: 工作流未触发
**原因**: 标签格式不正确或未推送到远程
**解决**:
```bash
# 检查标签是否推送
git tag -l
git ls-remote --tags origin

# 重新推送
git push origin v0.5.0
```

#### 问题 2: 构建失败 - 依赖问题
**原因**: Go 或 Node.js 依赖不完整
**解决**:
```bash
# 本地测试
go mod tidy
cd frontend && npm install
task package-all
```

#### 问题 3: macOS DMG 创建失败
**原因**: `hdiutil` 命令不可用
**解决**: 确保使用 `macos-latest` runner（已配置）

#### 问题 4: Release 未创建
**原因**: 不是从 tag 触发
**解决**: 检查工作流中的条件判断 `if: startsWith(github.ref, 'refs/tags/')`

#### 问题 5: 权限不足
**原因**: GITHUB_TOKEN 权限不够
**解决**: 确认工作流中有 `permissions: contents: write`

## 📊 监控和维护

### 查看构建日志
1. 进入 Actions 标签页
2. 点击具体的运行记录
3. 展开各个步骤查看详细输出

### 清理旧产物
```bash
# 删除本地的旧构建
rm -rf bin/releases/*

# GitHub 上的旧 Release 可以手动删除
# 或通过 GitHub API 批量清理
```

### 更新工作流
修改 `.github/workflows/release.yml` 后：
```bash
git add .github/workflows/release.yml
git commit -m "Update CI/CD workflow"
git push origin main
```

## 🎉 成功标志

构建成功后，你应该看到：

1. ✅ Actions 页面显示绿色对勾
2. ✅ Releases 页面出现新版本
3. ✅ 所有平台的构建产物已上传
4. ✅ Release Notes 自动生成

## 📚 相关资源

- [GitHub Actions 官方文档](https://docs.github.com/en/actions)
- [Wails v3 文档](https://v3.wails.io/)
- [Task 工具文档](https://taskfile.dev/)
- [softprops/action-gh-release](https://github.com/softprops/action-gh-release)

## 🔄 后续优化建议

1. **添加测试阶段**: 在打包前运行单元测试
2. **代码质量检查**: 集成 golangci-lint
3. **通知机制**: 构建完成后发送 Slack/Discord 通知
4. **Changelog 生成**: 自动生成变更日志
5. **Docker 镜像**: 同时推送 Docker Hub
6. **Homebrew Tap**: 自动更新 Homebrew formula

---

**创建时间**: 2026-04-07  
**工作流版本**: 1.0  
**维护者**: yhao521
