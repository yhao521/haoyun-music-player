# GitHub Actions 自动化发布指南

## 概述

本项目已配置 GitHub Actions 工作流，可以自动打包多平台应用并发布到 GitHub Release。

## 触发方式

### 1. 自动触发（推荐）

推送标签时自动触发构建和发布：

```bash
# 创建并推送标签
git tag v0.5.0
git push origin v0.5.0
```

标签格式必须为 `v*`（如 v0.5.0、v1.0.0 等）。

### 2. 手动触发

在 GitHub 仓库页面：
1. 进入 **Actions** 标签页
2. 选择 **Release Build** 工作流
3. 点击 **Run workflow** 按钮
4. 选择分支后运行

## 构建产物

### macOS
- `haoyun-music-player_vX.Y.Z_TIMESTAMP_macOS_arm64.dmg` - Apple Silicon 版本
- `haoyun-music-player_vX.Y.Z_TIMESTAMP_macOS_amd64.dmg` - Intel 版本
- `haoyun-music-player_vX.Y.Z_TIMESTAMP_macOS_universal.zip` - 通用版本（包含 README 和 LICENSE）

### Windows
- `haoyun-music-player_vX.Y.Z_TIMESTAMP_Windows_arm64.zip` - ARM64 版本
- `haoyun-music-player_vX.Y.Z_TIMESTAMP_Windows_amd64.zip` - x64 版本

### Linux（当前已禁用）
如需启用 Linux 构建，请取消 `Taskfile.yml` 中 `package-all` 任务里 Linux 相关行的注释。

## 工作流程说明

### 构建环境
- **macOS**: 使用 `macos-latest` runner（Apple Silicon M1/M2）
- **Windows**: 使用 `windows-latest` runner

### 构建步骤
1. 检出代码（获取完整 Git 历史以计算版本号）
2. 设置 Go 环境（从 go.mod 读取版本）
3. 设置 Node.js 环境（v20）
4. 安装 Task 工具
5. 安装 Wails CLI
6. 更新构建资源
7. 安装平台依赖（macOS: create-dmg, zip）
8. 执行 `task package-all` 进行多平台打包
9. 上传产物到 GitHub Release

### 版本号生成
版本号由三部分组成：
- `VERSION_1.VERSION_2`: 在 Taskfile.yml 中定义（当前为 0.0）
- `VERSION_3`: 通过 `git rev-list HEAD --no-merges --count` 自动计算
- `TIMESTAMP`: 构建时间戳（格式：YYYYMMDDHHmmSS）

最终版本格式：`v0.0.X_YYYYMMDDHHmmSS`

## 首次使用准备

### 1. 确保仓库已推送到 GitHub

```bash
git remote add origin https://github.com/yourusername/haoyun-music-player.git
git push -u origin main
```

### 2. 检查 Taskfile.yml 配置

确认以下任务存在且正常工作：
- `task package-all` - 主打包任务
- `task darwin:package` - macOS 打包
- `task windows:package` - Windows 打包

### 3. 测试本地打包

在推送标签前，建议先在本地测试打包：

```bash
# 测试 macOS 打包
task darwin:package

# 测试 Windows 打包（需要交叉编译环境）
ARCH=amd64 task windows:package
```

## 发布流程示例

### 发布新版本 v0.5.0

```bash
# 1. 确保所有更改已提交
git add .
git commit -m "Prepare for v0.5.0 release"

# 2. 推送代码
git push origin main

# 3. 创建标签
git tag v0.5.0

# 4. 推送标签（触发 CI/CD）
git push origin v0.5.0
```

推送标签后，GitHub Actions 会自动：
1. 启动构建工作流
2. 在 macOS 和 Windows runner 上并行构建
3. 打包所有平台的产物
4. 创建 GitHub Release 并上传文件

### 查看构建进度

1. 进入 GitHub 仓库页面
2. 点击 **Actions** 标签页
3. 查看 **Release Build** 工作流的运行状态
4. 点击具体运行记录查看详细日志

## 故障排查

### 构建失败常见原因

#### 1. Go 模块依赖问题
**症状**: `go mod tidy` 或编译失败
**解决**: 
```bash
go mod tidy
git add go.mod go.sum
git commit -m "Update dependencies"
```

#### 2. Node.js 依赖问题
**症状**: `npm install` 失败
**解决**:
```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
git add package-lock.json
git commit -m "Rebuild frontend dependencies"
```

#### 3. Wails CLI 版本问题
**症状**: `wails3` 命令找不到或版本不兼容
**解决**: 检查工作流中的 Wails 版本是否与项目匹配

#### 4. macOS DMG 创建失败
**症状**: `hdiutil` 命令失败
**解决**: 确保使用 macOS runner，`hdiutil` 是 macOS 内置工具

#### 5. 权限问题
**症状**: 无法写入 Release
**解决**: 检查工作流的 `permissions` 配置，确保有 `contents: write` 权限

### 手动调试

如果自动构建失败，可以手动触发工作流进行测试：

1. 进入 GitHub Actions 页面
2. 选择 **Release Build**
3. 点击 **Run workflow**
4. 查看日志定位问题

## 自定义配置

### 修改构建平台

编辑 `.github/workflows/release.yml` 中的 `matrix` 配置：

```yaml
strategy:
  matrix:
    os: [macos-latest, windows-latest, ubuntu-latest]  # 添加 Linux
```

### 修改 Node.js 版本

```yaml
- name: Set up Node.js
  uses: actions/setup-node@v4
  with:
    node-version: '18'  # 修改为你需要的版本
```

### 启用预发布版本

如果需要将某些版本标记为预发布：

```yaml
- name: Upload artifacts to GitHub Release
  uses: softprops/action-gh-release@v2
  with:
    prerelease: true  # 改为 true
```

### 添加构建缓存加速

工作流已启用 Go 和 npm 缓存，后续构建会更快。

## 安全注意事项

1. **GITHUB_TOKEN**: 工作流使用自动生成的 `GITHUB_TOKEN`，无需手动配置
2. **依赖安全**: 定期更新 Go 和 Node.js 依赖以修复安全漏洞
3. **Runner 安全**: 使用 GitHub 托管的 runner，无需担心环境问题

## 最佳实践

1. **语义化版本**: 遵循语义化版本规范（Major.Minor.Patch）
2. **测试后再发布**: 在本地充分测试后再推送标签
3. **发布说明**: 工作流会自动生成 Release Notes，也可以手动编辑
4. **草稿模式**: 如需审核，可将 `draft: false` 改为 `draft: true`
5. **清理旧产物**: 手动触发的构建产物保留 30 天后自动删除

## 相关文档

- [GitHub Actions 文档](https://docs.github.com/en/actions)
- [Wails v3 文档](https://v3.wails.io/)
- [Task 工具文档](https://taskfile.dev/)
- [softprops/action-gh-release](https://github.com/softprops/action-gh-release)

## 技术支持

如有问题，请：
1. 查看 GitHub Actions 日志
2. 检查本地是否能成功打包
3. 提交 Issue 并附上错误日志
