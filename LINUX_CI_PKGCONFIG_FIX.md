# Linux CI pkg-config 问题完整诊断与修复指南

## 🔴 当前错误

```
Package gtk+-3.0 was not found in the pkg-config search path.
Perhaps you should add the directory containing `gtk+-3.0.pc'
to the PKG_CONFIG_PATH environment variable
Package 'gtk+-3.0', required by 'virtual:world', not found
Package 'webkit2gtk-4.1', required by 'virtual:world', not found
```

## 📋 已实施的修复措施

### 1. 使用稳定的 Ubuntu 版本
- ✅ 从 `ubuntu-latest` 改为 `ubuntu-22.04`
- **原因**: `ubuntu-latest` 可能升级到 24.04,包名称和可用性会变化

### 2. 增强的包安装验证
```yaml
# 每个包安装后立即验证
if sudo apt-get install -y --no-install-recommends libgtk-3-dev; then
  # 检查 dpkg 确认真正安装成功
  if dpkg -l | grep -q "libgtk-3-dev"; then
    echo "✓ Verified: libgtk-3-dev is installed"
  else
    echo "✗ ERROR: Package not found after installation!"
    exit 1
  fi
fi
```

### 3. 精确的包名称检测
```bash
# 使用 --names-only 进行精确匹配
WEBKIT_41_AVAILABLE=$(apt-cache search --names-only '^libwebkit2gtk-4.1-dev$' | wc -l)
WEBKIT_40_AVAILABLE=$(apt-cache search --names-only '^libwebkit2gtk-4.0-dev$' | wc -l)
```

### 4. 全面的 .pc 文件搜索
```bash
# 搜索所有 GTK/WebKit 相关的 .pc 文件
find /usr -name "*gtk*.pc" -o -name "*gdk*.pc"
find /usr -name "*webkit*.pc"

# 精确查找所需的文件
find /usr -name "gtk+-3.0.pc"
find /usr -name "webkit2gtk-4.1.pc"
find /usr -name "webkit2gtk-4.0.pc"
```

### 5. 安装后立即测试 pkg-config
```bash
# 在安装完成后立即测试
echo "Testing GTK3..."
pkg-config --modversion gtk+-3.0

echo "Testing WebKit2GTK..."
pkg-config --modversion webkit2gtk-4.1 || pkg-config --modversion webkit2gtk-4.0
```

## 🔍 诊断步骤

当遇到此错误时,按以下步骤排查:

### 步骤 1: 检查包是否真正安装

查看 GitHub Actions 日志中的 "Install dependencies (Linux)" 步骤,找到:

```
=== Verifying installed packages ===
```

应该看到类似输出:
```
ii  libgtk-3-dev:amd64          3.24.33-1ubuntu2      amd64  GTK development files
ii  libwebkit2gtk-4.0-dev       2.38.6-0ubuntu0.22.04.1  amd64  WebKitGTK development files
```

**如果没有这些行或显示 "No matching packages found"**,说明包安装失败。

### 步骤 2: 检查 .pc 文件是否存在

在日志中查找:
```
All GTK-related .pc files found:
/usr/lib/x86_64-linux-gnu/pkgconfig/gtk+-3.0.pc

All WebKit-related .pc files found:
/usr/lib/x86_64-linux-gnu/pkgconfig/webkit2gtk-4.0.pc
```

**如果显示 "(none found)"**,说明:
- 包没有正确安装
- 或者 Ubuntu 版本不提供这些开发包

### 步骤 3: 检查 PKG_CONFIG_PATH

查找日志中的:
```
Final PKG_CONFIG_PATH: /usr/lib/x86_64-linux-gnu/pkgconfig:/usr/share/pkgconfig
```

然后检查:
```
Directories in PKG_CONFIG_PATH:
   ✓ Directory exists: /usr/lib/x86_64-linux-gnu/pkgconfig
     → Contains gtk+-3.0.pc
```

**如果目录不存在或不包含 .pc 文件**,需要手动添加路径。

### 步骤 4: 检查 pkg-config 测试结果

查找:
```
=== Immediate pkg-config verification after installation ===
Testing GTK3...
3.24.33
✓ GTK3 is immediately available via pkg-config
```

**如果这里就失败了**,说明即使安装了包,pkg-config 也找不到。

## 🛠️ 常见解决方案

### 方案 1: 包名称不匹配

**症状**: `apt-cache search` 显示包不可用

**解决**: Ubuntu 22.04 可能只有 `libwebkit2gtk-4.0-dev`,没有 4.1 版本

Workflow 已经自动处理这种情况,会尝试:
1. libwebkit2gtk-4.1-dev
2. libwebkit2gtk-4.0-dev (回退)
3. libwebkitgtk-dev (旧版本)

### 方案 2: .pc 文件在非标准位置

**症状**: 包已安装但 pkg-config 找不到

**解决**: Workflow 会自动搜索并添加路径到 PKG_CONFIG_PATH

如果需要手动指定,可以在 workflow 中添加:
```yaml
env:
  PKG_CONFIG_PATH: /usr/lib/x86_64-linux-gnu/pkgconfig:/usr/local/lib/pkgconfig
```

### 方案 3: 包安装被跳过或失败

**症状**: 日志中没有 "Successfully installed" 消息

**可能原因**:
1. APT 缓存过期
2. 网络问题
3. 包依赖冲突

**解决**: Workflow 已经添加了重试逻辑和详细错误输出

### 方案 4: Ubuntu 版本太新或太旧

**症状**: 所有包都找不到

**解决**: 
- Ubuntu 24.04: 可能有更新的包名(如 GTK4)
- Ubuntu 20.04: 可能需要更旧的包名

当前使用 **Ubuntu 22.04**,这是最稳定的选择。

## 📊 Ubuntu 22.04 预期行为

在 Ubuntu 22.04 上,你应该看到:

### 可用的包
```bash
libgtk-3-dev              # GTK 3.24.x
libwebkit2gtk-4.0-dev     # WebKit2GTK 4.0 (不是 4.1!)
libasound2-dev            # ALSA
```

### .pc 文件位置
```bash
/usr/lib/x86_64-linux-gnu/pkgconfig/gtk+-3.0.pc
/usr/lib/x86_64-linux-gnu/pkgconfig/webkit2gtk-4.0.pc  # 注意是 4.0
```

### pkg-config 测试
```bash
pkg-config --modversion gtk+-3.0        # 输出: 3.24.33
pkg-config --modversion webkit2gtk-4.0  # 输出: 2.38.x
pkg-config --modversion webkit2gtk-4.1  # 失败! 4.1 不存在
```

## ⚠️ 关键注意事项

### 1. WebKit 版本差异

**Ubuntu 22.04 只提供 WebKit2GTK 4.0**,不是 4.1!

如果你的代码中硬编码了 `webkit2gtk-4.1`:
```go
//go:build linux
// #cgo pkg-config: webkit2gtk-4.1  // ❌ 在 Ubuntu 22.04 上会失败
```

需要改为支持多个版本:
```go
//go:build linux
// #cgo pkg-config: webkit2gtk-4.0  // ✅ Ubuntu 22.04
```

或者使用构建标签区分:
```go
// ubuntu22.go
//go:build linux && ubuntu22
// #cgo pkg-config: webkit2gtk-4.0

// ubuntu24.go  
//go:build linux && ubuntu24
// #cgo pkg-config: webkit2gtk-4.1
```

### 2. Wails v3 的要求

Wails v3 可能要求特定版本的 WebKit。检查:
- Wails 文档中关于 Linux 依赖的说明
- `go.mod` 中 Wails 的版本要求

如果 Wails 严格要求 WebKit 4.1,你可能需要:
- 使用 Ubuntu 24.04
- 或等待 Wails 更新以支持 4.0

### 3. 检查你的 Go 代码

查看项目中是否有 CGO 指令指定了 WebKit 版本:

```bash
grep -r "webkit2gtk" --include="*.go" .
```

如果找到类似:
```go
// #cgo pkg-config: webkit2gtk-4.1
```

这就是问题所在!需要修改为 4.0 或支持多版本。

## 🎯 下一步行动

### 1. 提交当前修复
```bash
git add .github/workflows/release.yml
git commit -m "fix: 增强 Linux 依赖安装的诊断和验证"
git push
```

### 2. 触发新的构建
```bash
git tag v0.x.x
git push origin v0.x.x
```

### 3. 仔细检查日志

重点关注:
- ✅ 包是否真正安装 (`dpkg -l` 输出)
- ✅ `.pc` 文件是否存在 (`find` 输出)
- ✅ pkg-config 是否能找到库 (immediate verification)
- ❌ 任何错误消息或警告

### 4. 如果仍然失败

收集以下信息并提供:
1. 完整的 "Install dependencies (Linux)" 步骤日志
2. `dpkg -l | grep -E "(gtk|webkit)"` 的输出
3. `find /usr -name "*.pc" | grep -E "(gtk|webkit)"` 的输出
4. 你的 Go 代码中关于 WebKit 的 CGO 指令

## 🔧 临时解决方案

如果急需构建,可以:

### 选项 A: 使用 Docker 构建
```bash
# 在 macOS 上使用 Docker 交叉编译
task setup:docker
task build:docker
```

### 选项 B: 本地构建
```bash
# 如果你有 Ubuntu 22.04 机器
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.0-dev libasound2-dev
task linux:package
```

### 选项 C: 禁用 Linux 构建
暂时从 matrix 中移除 Linux:
```yaml
matrix:
  os: [macos-latest, windows-latest]  # 暂时移除 ubuntu-22.04
```

## 📝 总结

当前修复已经大幅增强了诊断能力,应该能够:
1. ✅ 准确检测可用的包版本
2. ✅ 验证包是否真正安装
3. ✅ 找到所有 .pc 文件的位置
4. ✅ 配置正确的 PKG_CONFIG_PATH
5. ✅ 提供详细的调试信息

如果仍然失败,日志会明确指出是哪个环节出了问题,从而可以快速定位和修复。
