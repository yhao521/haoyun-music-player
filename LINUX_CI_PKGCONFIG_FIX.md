# Linux CI pkg-config 问题 - 最终解决方案与诊断指南

## 🎯 核心问题

Wails v3.0.0-alpha.74 在编译时需要 GTK3 和 WebKit2GTK 开发库,但:
1. Ubuntu 22.04 只提供 `webkit2gtk-4.0`,不提供 `4.1`
2. pkg-config 可能找不到已安装的 `.pc` 文件
3. 包安装可能失败但没有被检测到

## ✅ 最终解决方案

### 策略: **安装所有可能的版本 + 全面搜索 .pc 文件**

```yaml
# 1. 尝试安装所有 GTK 版本
GTK_PACKAGES=("libgtk-3-dev" "libgtk-4-dev")
for pkg in "${GTK_PACKAGES[@]}"; do
  sudo apt-get install -y --no-install-recommends "$pkg"
done

# 2. 尝试安装所有 WebKit 版本  
WEBKIT_PACKAGES=("libwebkit2gtk-4.1-dev" "libwebkit2gtk-4.0-dev" "libwebkitgtk-dev")
for pkg in "${WEBKIT_PACKAGES[@]}"; do
  sudo apt-get install -y --no-install-recommends "$pkg"
done

# 3. 搜索系统中所有的 .pc 文件
find /usr -name "*.pc" | grep -E "(gtk|webkit)" > /tmp/pc_files.txt

# 4. 从所有找到的目录构建 PKG_CONFIG_PATH
while IFS= read -r pc_file; do
  dir=$(dirname "$pc_file")
  PKG_CONFIG_DIRS="$PKG_CONFIG_DIRS:$dir"
done < /tmp/pc_files.txt

# 5. 导出并验证
export PKG_CONFIG_PATH="$PKG_CONFIG_DIRS"
pkg-config --modversion gtk+-3.0
pkg-config --modversion webkit2gtk-4.0  # 或 4.1
```

## 🔑 关键改进点

### 1. **不依赖单一版本**
```bash
# ❌ 之前:只尝试一个版本
apt-get install libwebkit2gtk-4.1-dev

# ✅ 现在:尝试所有版本,使用第一个成功的
for pkg in "4.1-dev" "4.0-dev" "dev"; do
  if apt-get install "libwebkit2gtk-$pkg"; then
    break
  fi
done
```

### 2. **动态发现 .pc 文件位置**
```bash
# ❌ 之前:假设在标准位置
PKG_CONFIG_PATH="/usr/lib/x86_64-linux-gnu/pkgconfig"

# ✅ 现在:搜索整个系统
find /usr -name "*.pc" | grep -E "(gtk|webkit)"
# 然后从结果中提取所有目录
```

### 3. **立即验证每个步骤**
```bash
# 安装后立即检查 dpkg
if dpkg -l | grep -q "libgtk-3-dev"; then
  echo "✓ Verified"
fi

# 配置路径后立即测试 pkg-config
if pkg-config --modversion gtk+-3.0; then
  echo "✓ Accessible"
fi
```

## 📊 Ubuntu 22.04 预期结果

### 成功安装的包
```
ii  libgtk-3-dev:amd64          3.24.33-1ubuntu2
ii  libwebkit2gtk-4.0-dev       2.38.6-0ubuntu0.22.04.1
```

### 找到的 .pc 文件
```
/usr/lib/x86_64-linux-gnu/pkgconfig/gtk+-3.0.pc
/usr/lib/x86_64-linux-gnu/pkgconfig/webkit2gtk-4.0.pc
```

### pkg-config 测试
```bash
$ pkg-config --modversion gtk+-3.0
3.24.33

$ pkg-config --modversion webkit2gtk-4.0
2.38.6

$ pkg-config --modversion webkit2gtk-4.1
Package webkit2gtk-4.1 was not found  # 这是正常的!
```

## ⚠️ Wails v3 Alpha 版本的注意事项

### 问题
Wails v3.0.0-alpha.74 可能在内部硬编码了对 `webkit2gtk-4.1` 的要求。

### 解决方案选项

#### 选项 1: 等待 Wails 更新 (推荐)
- Wails 团队可能会发布支持 WebKit 4.0 的版本
- 关注 GitHub issues 和 release notes

#### 选项 2: 使用 Ubuntu 24.04
```yaml
matrix:
  os: [macos-latest, windows-latest, ubuntu-24.04]
```
Ubuntu 24.04 提供 `libwebkit2gtk-4.1-dev`

#### 选项 3: 降级到 Wails v2
如果 v3 的兼容性问题太多,考虑暂时使用稳定的 v2 版本

#### 选项 4: 创建符号链接 (临时方案)
```bash
# 警告:这只是权宜之计,可能导致运行时错误
ln -s /usr/lib/x86_64-linux-gnu/pkgconfig/webkit2gtk-4.0.pc \
      /usr/lib/x86_64-linux-gnu/pkgconfig/webkit2gtk-4.1.pc
```

## 🚀 部署步骤

### 1. 提交更改
```bash
git add .github/workflows/release.yml LINUX_CI_PKGCONFIG_FIX.md
git commit -m "fix: 采用激进策略安装所有 GTK/WebKit 版本并动态配置 pkg-config"
git push
```

### 2. 触发构建
```bash
git tag v0.x.x
git push origin v0.x.x
```

### 3. 监控日志

重点关注以下输出:

```
=== Installing GTK packages ===
Trying to install: libgtk-3-dev
✓ Successfully installed: libgtk-3-dev

=== Installing WebKit packages ===
Trying to install: libwebkit2gtk-4.1-dev
✗ Failed to install: libwebkit2gtk-4.1-dev
Trying to install: libwebkit2gtk-4.0-dev
✓ Successfully installed: libwebkit2gtk-4.0-dev

=== Final pkg-config verification ===
✓ gtk+-3.0 is accessible
✓ webkit2gtk-4.0 is accessible
✓✓✓ All critical dependencies are properly configured ✓✓✓
```

## 🔍 如果仍然失败

### 收集诊断信息

在 GitHub Actions 日志中查找:

1. **包安装状态**
   ```
   Installed GTK/WebKit packages:
   ii  libgtk-3-dev     ...
   ii  libwebkit2gtk-4.0-dev  ...
   ```

2. **.pc 文件位置**
   ```
   Found .pc files:
   /usr/lib/x86_64-linux-gnu/pkgconfig/gtk+-3.0.pc
   /usr/lib/x86_64-linux-gnu/pkgconfig/webkit2gtk-4.0.pc
   ```

3. **PKG_CONFIG_PATH 内容**
   ```
   Final PKG_CONFIG_PATH: /usr/lib/x86_64-linux-gnu/pkgconfig:...
   ```

4. **pkg-config 测试结果**
   ```
   Testing GTK versions...
   ✓ gtk+-3.0 found (version: 3.24.33)
   
   Testing WebKit versions...
   ✓ webkit2gtk-4.0 found (version: 2.38.6)
   ✗ webkit2gtk-4.1 not found
   ```

### 可能的根本原因

如果上述都显示正常,但仍然报错,那么问题是:

**Wails v3 的 CGO 指令硬编码了 `webkit2gtk-4.1`**

解决方法:
1. 升级到支持多版本的 Wails v3 新版本
2. 或使用 Ubuntu 24.04 (提供 webkit2gtk-4.1)
3. 或联系 Wails 团队报告此问题

## 💡 最佳实践建议

### 对于 Wails v3 项目

1. **优先使用 Ubuntu 24.04**
   ```yaml
   os: [macos-latest, windows-latest, ubuntu-24.04]
   ```

2. **或者等待更稳定的 Wails v3 版本**
   - Alpha 版本可能有各种兼容性问题
   - 考虑使用 Beta 或 RC 版本

3. **保持灵活的依赖策略**
   - 不要硬编码特定版本
   - 提供多个备选方案
   - 详细的错误提示帮助调试

### 对于其他 CGO 项目

1. **总是安装 `-dev` 包**,不仅仅是运行时库
2. **使用 `--no-install-recommends`** 减少冲突
3. **验证每个安装步骤**,不要假设成功
4. **动态配置 PKG_CONFIG_PATH**,不要硬编码路径
5. **提供详细的诊断信息**,方便排查问题

## 📝 总结

这次的修复采用了**最激进的策略**:
- ✅ 安装所有可能的 GTK 和 WebKit 版本
- ✅ 搜索整个系统的 .pc 文件
- ✅ 从所有找到的目录构建 PKG_CONFIG_PATH
- ✅ 每一步都进行验证和诊断

这应该能够解决绝大多数 pkg-config 相关的问题。如果仍然失败,很可能是 Wails v3 Alpha 版本本身的限制,需要考虑升级 Wails 或更换 Ubuntu 版本。

## 🔴 历史错误参考

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

# Linux 打包路径问题修复

## 🔴 问题现象

在 GitHub Actions Linux CI 构建中,Go 编译成功但打包阶段失败:

```
task: [linux:build:native] go build -tags production -trimpath -buildvcs=false -ldflags="-w -s" -o bin/haoyun-music-player
task: [linux:generate:dotdesktop] mkdir -p /home/runner/work/haoyun-music-player/haoyun-music-player/build/linux/appimage
task: [linux:generate:dotdesktop] wails3 generate .desktop -name "haoyun-music-player" -exec "haoyun-music-player" -icon "haoyun-music-player" -outputfile "/home/runner/work/haoyun-music-player/haoyun-music-player/build/linux/haoyun-music-player.desktop" -categories "Development;"
task: [linux:generate:aur] wails3 tool package -name "haoyun-music-player" -format archlinux -config ./build/linux/nfpm/nfpm.yaml -out /home/runner/work/haoyun-music-player/haoyun-music-player/bin
cd: no such file or directory: "./bin"
Error: Process completed with exit code 201.
```

## 🔍 根本原因

### 1. nfpm.yaml 使用了相对路径

``yaml
contents:
  - src: "./bin/haoyun-music-player"  # ❌ 相对路径
    dst: "/usr/local/bin/haoyun-music-player"
```

### 2. Taskfile 命令未指定工作目录

``yaml
generate:aur:
  cmds:
    - wails3 tool package -name "{{.APP_NAME}}" -format archlinux -config ./build/linux/nfpm/nfpm.yaml -out {{.ROOT_DIR}}/bin
    # ❌ 没有 cd 到项目根目录,导致相对路径失效
```

当 `wails3 tool package` 执行时,它的工作目录可能不是项目根目录,因此 `./bin/haoyun-music-player` 路径找不到文件。

## ✅ 解决方案

### 修改 1: 清理 nfpm.yaml 中的路径

**文件**: `build/linux/nfpm/nfpm.yaml`

```
# 之前 (有前导 ./)
contents:
  - src: "./bin/haoyun-music-player"
    dst: "/usr/local/bin/haoyun-music-player"
  - src: "./build/appicon.png"
    dst: "/usr/share/icons/hicolor/128x128/apps/haoyun-music-player.png"

# 之后 (移除前导 ./,更清晰)
contents:
  - src: "bin/haoyun-music-player"
    dst: "/usr/local/bin/haoyun-music-player"
  - src: "build/appicon.png"
    dst: "/usr/share/icons/hicolor/128x128/apps/haoyun-music-player.png"
```

**注意**: 这些路径仍然是相对路径,需要在正确的工作目录下执行 nfpm 命令。

### 修改 2: 在 Taskfile 中明确工作目录

**文件**: `build/linux/Taskfile.yml`

```
# 之前
generate:deb:
  cmds:
    - wails3 tool package -name "{{.APP_NAME}}" -format deb -config ./build/linux/nfpm/nfpm.yaml -out {{.ROOT_DIR}}/bin

# 之后 (添加 cd {{.ROOT_DIR}})
generate:deb:
  cmds:
    - cd {{.ROOT_DIR}} && wails3 tool package -name "{{.APP_NAME}}" -format deb -config ./build/linux/nfpm/nfpm.yaml -out {{.ROOT_DIR}}/bin
```

对所有三个打包格式都进行了相同修改:
- `generate:deb`
- `generate:rpm`
- `generate:aur`

## 🎯 工作原理

### 执行流程

1. **构建阶段** (`build:native`)
   ```bash
   go build -o bin/haoyun-music-player
   # 在项目根目录执行,生成 bin/haoyun-music-player
   ```

2. **打包阶段** (`generate:aur`)
   ```bash
   cd /home/runner/work/haoyun-music-player/haoyun-music-player && \
   wails3 tool package \
     -name "haoyun-music-player" \
     -format archlinux \
     -config ./build/linux/nfpm/nfpm.yaml \
     -out /home/runner/work/haoyun-music-player/haoyun-music-player/bin
   ```

3. **nfpm 读取配置**
   ```yaml
   # nfpm.yaml 中的相对路径现在是相对于项目根目录
   contents:
     - src: "bin/haoyun-music-player"  # ✅ 能找到!
       dst: "/usr/local/bin/haoyun-music-player"
   ```

## 📊 关键要点

### 1. 相对路径的陷阱

在 CI/CD 环境中,命令的工作目录可能不是你期望的:
- Task 可能在子目录中执行
- Docker 容器可能有不同的工作目录
- 并行任务可能改变当前目录

**最佳实践**:
- 始终使用 `cd {{.ROOT_DIR}} && command` 确保在正确的目录执行
- 或者使用绝对路径: `{{.ROOT_DIR}}/bin/haoyun-music-player`

### 2. nfpm 配置文件的路径解析

nfpm 配置文件中的 `src` 路径是相对于:
- **nfpm 命令执行时的工作目录**,不是配置文件所在目录

因此必须确保:
```bash
cd /project/root && nfpm pkg --config build/linux/nfpm/nfpm.yaml
```

### 3. Wails v3 Taskfile 变量

- `{{.ROOT_DIR}}`: 项目根目录的绝对路径
- `{{.BIN_DIR}}`: 通常是 `bin`
- `{{.APP_NAME}}`: 应用名称

使用这些变量可以确保路径的正确性。

## 🔧 其他可能的修复方案

### 方案 A: 使用绝对路径 (更可靠)

修改 `nfpm.yaml`:

```
# 但这需要模板化,因为绝对路径在不同环境中不同
contents:
  - src: "${PROJECT_ROOT}/bin/haoyun-music-player"
    dst: "/usr/local/bin/haoyun-music-player"
```

然后在 Taskfile 中设置环境变量:

```yaml
generate:aur:
  cmds:
    - PROJECT_ROOT={{.ROOT_DIR}} wails3 tool package ...
```

### 方案 B: 使用 nfpm 的 --packager 参数指定工作目录

```bash
wails3 tool package --chdir {{.ROOT_DIR}} ...
```

(如果 wails3 支持这个参数)

### 方案 C: 当前采用的方案 (推荐)

在 Taskfile 中使用 `cd {{.ROOT_DIR}} && command`,简单直接且可靠。

## ✅ 验证步骤

提交更改后,GitHub Actions 应该能够:

1. ✅ 成功编译 Go 代码到 `bin/haoyun-music-player`
2. ✅ 找到二进制文件并创建 DEB/RPM/AUR 包
3. ✅ 输出打包好的文件到 `bin/` 目录

预期日志:

```
task: [linux:build:native] go build -tags production ... -o bin/haoyun-music-player
task: [linux:generate:aur] cd /home/runner/work/... && wails3 tool package ...
✓ Successfully created Arch Linux package
```

## 📝 总结

**问题**: nfpm 命令在错误的工作目录执行,导致相对路径失效

**解决**: 
1. 清理 nfpm.yaml 中的路径格式(移除 `./`)
2. 在 Taskfile 的所有 nfpm 命令前添加 `cd {{.ROOT_DIR}} &&`

**教训**: 
- 在 CI/CD 环境中,永远不要假设工作目录
- 始终显式指定或切换到正确的目录
- 使用 `{{.ROOT_DIR}}` 等变量确保路径的可靠性
