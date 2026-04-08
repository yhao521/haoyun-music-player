# Linux CI pkg-config 问题修复说明

## 问题现象

在 GitHub Actions Linux CI 环境中构建时出现以下错误:

```
Package gtk+-3.0 was not found in the pkg-config search path.
Perhaps you should add the directory containing `gtk+-3.0.pc'
to the PKG_CONFIG_PATH environment variable
Package 'gtk+-3.0', required by 'virtual:world', not found
Package 'webkit2gtk-4.1', required by 'virtual:world', not found
```

## 根本原因分析

### 1. PKG_CONFIG_PATH 配置不完整
- 虽然安装了 GTK3 和 WebKit2GTK 开发库,但 `.pc` 文件可能不在标准搜索路径中
- 原配置逻辑可能在某些情况下生成空的路径或带有前导冒号的路径

### 2. 环境变量传递问题
- GitHub Actions 中写入 `$GITHUB_ENV` 只对后续步骤有效
- `task` 命令及其子进程(Go 编译器、pkg-config)可能无法正确继承环境变量
- 需要使用 `env` 命令显式传递关键环境变量

## 解决方案

### 修改 1: 改进 PKG_CONFIG_PATH 配置逻辑

**位置**: `.github/workflows/release.yml` - "Install dependencies (Linux)" 步骤

**关键改进**:

1. **确保默认路径存在**:
   ```bash
   # Ensure PKG_CONFIG_DIRS is not empty
   if [ -z "$PKG_CONFIG_DIRS" ]; then
     echo "WARNING: No .pc files found in standard locations!"
     echo "Setting default PKG_CONFIG_PATH..."
     PKG_CONFIG_DIRS="/usr/lib/x86_64-linux-gnu/pkgconfig:/usr/share/pkgconfig"
   fi
   ```

2. **正确的字符串拼接**(避免前导冒号):
   ```bash
   # 之前(可能有前导冒号):
   export PKG_CONFIG_PATH="$PKG_CONFIG_DIRS:$PKG_CONFIG_PATH"
   
   # 现在(正确处理空值):
   export PKG_CONFIG_PATH="$PKG_CONFIG_DIRS${PKG_CONFIG_PATH:+:$PKG_CONFIG_PATH}"
   ```
   
   `${PKG_CONFIG_PATH:+:$PKG_CONFIG_PATH}` 的含义:
   - 如果 `PKG_CONFIG_PATH` 非空,则扩展为 `:$PKG_CONFIG_PATH`
   - 如果 `PKG_CONFIG_PATH` 为空,则扩展为空字符串
   - 避免了 `:/usr/...` 这样的前导冒号问题

3. **导出到 GITHUB_ENV**:
   ```bash
   echo "PKG_CONFIG_PATH=$PKG_CONFIG_PATH" >> $GITHUB_ENV
   ```

### 修改 2: 增强诊断信息

**位置**: `.github/workflows/release.yml` - "Build and Package (Linux)" 步骤

**新增诊断内容**:

1. **全局搜索 .pc 文件**:
   ```bash
   ALL_GTK_PC=$(find /usr -name "gtk+-3.0.pc" 2>/dev/null || true)
   ALL_WEBKIT_PC=$(find /usr -name "webkit*.pc" 2>/dev/null || true)
   ```

2. **显示已安装的包**:
   ```bash
   dpkg -l | grep -E "(libgtk-3-dev|libwebkit2gtk)" || echo "No GTK/WebKit dev packages found!"
   ```

3. **逐个检查 PKG_CONFIG_PATH 中的目录**:
   ```bash
   for dir in $(echo "$PKG_CONFIG_PATH" | tr ':' '\n'); do
     if [ -d "$dir" ]; then
       echo "   ✓ Directory exists: $dir"
       ls "$dir"/gtk+-3.0.pc 2>/dev/null && echo "     → Contains gtk+-3.0.pc" || echo "     → No gtk+-3.0.pc"
     else
       echo "   ✗ Directory does NOT exist: $dir"
     fi
   done
   ```

4. **提供明确的解决建议**:
   ```bash
   echo "SUGGESTION: The .pc files may not be installed or are in a non-standard location."
   echo "Try reinstalling libgtk-3-dev: sudo apt-get install --reinstall libgtk-3-dev"
   ```

### 修改 3: 显式传递环境变量

**关键代码**:
```bash
# Export to current shell
export PKG_CONFIG_PATH="$PKG_CONFIG_PATH"

# Use env command to explicitly pass to task subprocess
env PKG_CONFIG_PATH="$PKG_CONFIG_PATH" CGO_ENABLED=1 task package-linux
```

这确保了:
- 当前 shell 环境有正确的 `PKG_CONFIG_PATH`
- `task` 命令及其所有子进程都能访问该变量
- 避免环境变量在进程间传递时丢失

## 验证步骤

提交更改后,GitHub Actions 将:

1. **安装依赖阶段**:
   - 安装 `libgtk-3-dev`, `libwebkit2gtk-4.1-dev`(或 4.0), `libasound2-dev`
   - 查找所有 `.pc` 文件位置
   - 构建完整的 `PKG_CONFIG_PATH` 并导出

2. **构建阶段**:
   - 显示详细的诊断信息(找到的 .pc 文件、已安装的包等)
   - 验证 `pkg-config` 能否找到 GTK3 和 WebKit2GTK
   - 如果失败,提供详细的调试信息和解决建议
   - 使用 `env` 命令显式传递环境变量执行 `task package-linux`

## 常见问题排查

### 问题 1: 仍然找不到 .pc 文件

**检查点**:
1. 查看日志中的 "Searching for GTK and WebKit .pc files" 部分
2. 确认 `dpkg -l` 输出中包含 `libgtk-3-dev` 和 `libwebkit2gtk-*-dev`
3. 如果 `.pc` 文件存在但 pkg-config 找不到,检查文件权限

**解决方案**:
```bash
# 重新安装包
sudo apt-get install --reinstall libgtk-3-dev libwebkit2gtk-4.1-dev

# 手动查找 .pc 文件
find /usr -name "*.pc" | grep -E "(gtk|webkit)"

# 手动测试 pkg-config
PKG_CONFIG_PATH=/usr/lib/x86_64-linux-gnu/pkgconfig pkg-config --modversion gtk+-3.0
```

### 问题 2: PKG_CONFIG_PATH 为空

**原因**: 没有找到任何 `.pc` 文件

**解决方案**:
- 检查 "Install dependencies (Linux)" 步骤的日志
- 确认 `apt-get install` 成功执行
- 查看是否有包冲突或安装失败的错误信息

### 问题 3: 包版本不匹配

**症状**: `webkit2gtk-4.1` 不可用,只有 `webkit2gtk-4.0`

**解决方案**:
- Workflow 已经支持自动检测并回退到 4.0 版本
- 查看日志中的 "Will install: libwebkit2gtk-4.0-dev (fallback to 4.0)" 提示

## 技术要点总结

1. **GitHub Actions 环境变量传递机制**:
   - `$GITHUB_ENV`: 对后续步骤有效
   - `export`: 对当前脚本内的子进程有效
   - `env VAR=value command`: 显式传递给特定命令

2. **Bash 字符串处理技巧**:
   - `${VAR:+value}`: 条件扩展,避免空值问题
   - `tr ':' '\n'`: 将冒号分隔的路径转换为多行

3. **pkg-config 工作原理**:
   - 搜索 `.pc` 文件获取编译和链接参数
   - 搜索路径由 `PKG_CONFIG_PATH` 环境变量控制
   - 默认搜索路径可通过 `pkg-config --variable pc_path pkg-config` 查看

4. **防御性编程最佳实践**:
   - 在关键操作前进行预检
   - 失败时提供充分的诊断信息
   - 给出明确的解决建议

## 相关文件

- `.github/workflows/release.yml`: CI/CD 工作流配置
- `Taskfile.yml`: 构建任务定义
- `build/linux/Taskfile.yml`: Linux 平台特定的构建任务

## 更新历史

- **2026-04-08**: 初始修复,改进 PKG_CONFIG_PATH 配置和环境变量传递
- **2026-04-08**: 增强诊断信息,添加默认路径保护,修复字符串拼接问题
