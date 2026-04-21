# NAS/SMB 远程音乐库实现方案

## 📋 目录

- [1. 方案概述](#1-方案概述)
- [2. 技术方案对比](#2-技术方案对比)
- [3. 推荐方案详解](#3-推荐方案详解)
- [4. 技术架构设计](#4-技术架构设计)
- [5. 核心代码实现](#5-核心代码实现)
- [6. 前端 UI 改造](#6-前端-ui-改造)
- [7. 后端服务接口](#7-后端服务接口)
- [8. 安全与优化](#8-安全与优化)
- [9. 实施路线图](#9-实施路线图)
- [10. 注意事项](#10-注意事项)

---

## 1. 方案概述

### 1.1 背景

随着用户音乐收藏的增长，本地存储已无法满足需求。越来越多的用户使用 NAS（网络附加存储）或 SMB 共享来集中管理音乐文件。本方案旨在为 Haoyun Music Player 添加对 NAS/SMB 远程音乐库的支持，让用户可以直接播放存储在远程服务器上的音乐。

### 1.2 目标

- ✅ 支持通过 SMB/CIFS 协议连接 NAS 设备
- ✅ 透明扫描远程音乐库，提取元数据
- ✅ 保持与本地音乐库一致的用户体验
- ✅ 确保跨平台兼容性（macOS/Windows/Linux）
- ✅ 提供安全的凭据管理和连接测试功能

---

## 2. 技术方案对比

### 2.1 可选方案

| 方案 | 优点 | 缺点 | 适用场景 |
|------|------|------|----------|
| **SMB 客户端直连** | 纯 Go 实现，跨平台，无需系统依赖 | 需要处理认证、网络超时、重连 | ✅ **推荐** |
| **系统挂载 (mount)** | 透明访问，像本地文件一样 | macOS/Windows/Linux 实现差异大，权限复杂 | 备选 |
| **NFS 协议** | Linux 原生支持好 | Windows/macOS 需额外配置 | 特定场景 |
| **WebDAV** | HTTP 基础，防火墙友好 | 性能略低于 SMB | 云端存储 |

### 2.2 Go 语言 SMB 客户端库对比

| 库名称 | 版本 | 特点 | 维护状态 | 推荐度 |
|--------|------|------|----------|--------|
| **[hirochachacha/go-smb2](https://github.com/hirochachacha/go-smb2)** | v1.x | 纯 Go，SMB2/3，API 简洁 | ⭐ 活跃 | ✅ **首选** |
| CloudSoda/go-smb2 | v0.x | hirochachacha 的 fork，修复 bug | 活跃 | ✅ 备选 |
| stacktitan/smb | v1.x | 仅 SMB1，较老 | 不活跃 | ❌ 不推荐 |
| mvo5/libsmbclient-go | - | CGO 绑定，线程不安全 | 有限维护 | ⚠️ 慎用 |

**选择理由**：
- `hirochachacha/go-smb2` 是纯 Go 实现，无 CGO 依赖，编译简单
- 支持 SMB2/SMB3 协议，安全性更好
- API 设计清晰，易于集成到现有架构
- 跨平台兼容性好，符合项目规范

---

## 3. 推荐方案详解

### 3.1 技术栈

```yaml
核心库: github.com/hirochachacha/go-smb2
密码加密: golang.org/x/crypto/bcrypt
并发控制: sync.Mutex / sync.RWMutex
缓存策略: 内存缓存 + 本地临时文件
```

### 3.2 架构设计原则

1. **透明集成**: 对上层业务逻辑隐藏 SMB 细节，统一使用 `MusicLibrary` 接口
2. **异步扫描**: 远程扫描耗时较长，必须使用 goroutine 异步执行
3. **错误恢复**: 实现断线重连和指数退避重试机制
4. **资源管理**: 及时关闭 SMB 连接，避免资源泄漏
5. **安全存储**: 敏感凭据加密存储，不在日志中明文输出

---

## 4. 技术架构设计

### 4.1 数据结构扩展

#### MusicLibrary 结构体扩展

在 `backend/libraryservice.go` 中扩展 `MusicLibrary` 结构：

```go
// MusicLibrary 音乐库结构
type MusicLibrary struct {
    Name      string      `json:"name"`
    Path      string      `json:"path"`       // 本地路径或 smb://server/share/path
    Type      string      `json:"type"`       // "local" 或 "smb"
    SMBConfig *SMBConfig  `json:"smb_config,omitempty"` // SMB 配置（仅远程库）
    CreatedAt time.Time   `json:"created_at"`
    UpdatedAt time.Time   `json:"updated_at"`
    Tracks    []TrackInfo `json:"tracks"`
}

// SMBConfig SMB 连接配置
type SMBConfig struct {
    Server   string `json:"server"`    // 服务器地址 (IP 或域名)
    Port     int    `json:"port"`      // 端口，默认 445
    Share    string `json:"share"`     // 共享名称
    Username string `json:"username"`  // 用户名
    Password string `json:"password"`  // 密码（加密存储）
    Domain   string `json:"domain"`    // 域（可选，如 WORKGROUP）
    Path     string `json:"path"`      // 共享内的相对路径
}
```

#### TrackInfo 字段说明

```go
type TrackInfo struct {
    Path      string `json:"path"`                 // 歌曲路径（SMB 库为相对路径）
    Filename  string `json:"filename"`             // 文件名
    Title     string `json:"title,omitempty"`      // 标题
    Artist    string `json:"artist,omitempty"`     // 艺术家
    Album     string `json:"album,omitempty"`      // 专辑
    Duration  int64  `json:"duration"`             // 秒
    Size      int64  `json:"size"`                 // 字节
    LyricPath string `json:"lyric_path,omitempty"` // 歌词文件路径
}
```

### 4.2 模块划分

```
backend/
├── libraryservice.go      # 音乐库管理（扩展现有逻辑）
├── smbclient.go           # 【新增】SMB 客户端封装
├── music_service.go       # 服务层接口（扩展 SMB 相关方法）
└── pkg/
    └── crypto/            # 【新增】加密工具（可选）
        └── password.go
```

---

## 5. 核心代码实现

### 5.1 SMB 客户端封装 (`backend/smbclient.go`)

```go
package backend

import (
    "context"
    "fmt"
    "io"
    "net"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/hirochachacha/go-smb2"
)

// SMBClientWrapper SMB 客户端封装
type SMBClientWrapper struct {
    config  *SMBConfig
    session *smb2.Session
    share   *smb2.Share
    conn    net.Conn
}

// NewSMBClient 创建 SMB 客户端
func NewSMBClient(config *SMBConfig) (*SMBClientWrapper, error) {
    if config.Port == 0 {
        config.Port = 445
    }

    // 建立 TCP 连接
    addr := fmt.Sprintf("%s:%d", config.Server, config.Port)
    conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
    if err != nil {
        return nil, fmt.Errorf("连接 SMB 服务器失败：%w", err)
    }

    // 创建拨号器
    dialer := &smb2.Dialer{
        Initiator: &smb2.NTLMInitiator{
            User:     config.Username,
            Password: config.Password,
            Domain:   config.Domain,
        },
    }

    // 建立 SMB 会话
    session, err := dialer.Dial(conn)
    if err != nil {
        conn.Close()
        return nil, fmt.Errorf("SMB 会话建立失败：%w", err)
    }

    // 挂载共享
    share, err := session.Mount(config.Share)
    if err != nil {
        session.Logoff()
        conn.Close()
        return nil, fmt.Errorf("挂载共享失败：%w", err)
    }

    return &SMBClientWrapper{
        config:  config,
        session: session,
        share:   share,
        conn:    conn,
    }, nil
}

// ReadFile 读取远程文件内容
func (sc *SMBClientWrapper) ReadFile(path string) ([]byte, error) {
    fullPath := filepath.Join(sc.config.Path, path)
    f, err := sc.share.Open(fullPath)
    if err != nil {
        return nil, fmt.Errorf("打开文件失败：%w", err)
    }
    defer f.Close()

    return io.ReadAll(f)
}

// Stat 获取文件信息
func (sc *SMBClientWrapper) Stat(path string) (os.FileInfo, error) {
    fullPath := filepath.Join(sc.config.Path, path)
    return sc.share.Stat(fullPath)
}

// ReadDir 读取目录内容
func (sc *SMBClientWrapper) ReadDir(path string) ([]os.FileInfo, error) {
    fullPath := filepath.Join(sc.config.Path, path)
    return sc.share.ReadDir(fullPath)
}

// Walk 递归遍历目录（类似 filepath.Walk）
func (sc *SMBClientWrapper) Walk(root string, walkFn filepath.WalkFunc) error {
    return sc.walk(root, walkFn)
}

func (sc *SMBClientWrapper) walk(path string, walkFn filepath.WalkFunc) error {
    info, err := sc.Stat(path)
    if err != nil {
        return walkFn(path, nil, err)
    }

    err = walkFn(path, info, nil)
    if err != nil {
        return err
    }

    if !info.IsDir() {
        return nil
    }

    entries, err := sc.ReadDir(path)
    if err != nil {
        return walkFn(path, info, err)
    }

    for _, entry := range entries {
        subPath := filepath.Join(path, entry.Name())
        if err := sc.walk(subPath, walkFn); err != nil {
            return err
        }
    }

    return nil
}

// Close 关闭 SMB 连接
func (sc *SMBClientWrapper) Close() error {
    var errs []error

    if sc.share != nil {
        if err := sc.share.Umount(); err != nil {
            errs = append(errs, err)
        }
    }

    if sc.session != nil {
        sc.session.Logoff()
    }

    if sc.conn != nil {
        if err := sc.conn.Close(); err != nil {
            errs = append(errs, err)
        }
    }

    if len(errs) > 0 {
        return fmt.Errorf("关闭 SMB 连接时发生错误：%v", errs)
    }

    return nil
}

// IsAudioFile 判断是否为音频文件
func IsAudioFile(filename string) bool {
    ext := strings.ToLower(filepath.Ext(filename))
    audioExts := map[string]bool{
        ".mp3":  true,
        ".wav":  true,
        ".flac": true,
        ".aac":  true,
        ".m4a":  true,
        ".ogg":  true,
        ".wma":  true,
        ".ape":  true,
        ".opus": true,
    }
    return audioExts[ext]
}

// ReadFileWithRetry 带重试的文件读取
func (sc *SMBClientWrapper) ReadFileWithRetry(path string, maxRetries int) ([]byte, error) {
    var lastErr error
    for i := 0; i < maxRetries; i++ {
        data, err := sc.ReadFile(path)
        if err == nil {
            return data, nil
        }
        lastErr = err
        
        // 指数退避重试
        time.Sleep(time.Duration(i+1) * time.Second)
        
        // 第 2 次重试后重新连接
        if i >= 1 {
            sc.Close()
            newClient, err := NewSMBClient(sc.config)
            if err != nil {
                lastErr = fmt.Errorf("重新连接失败：%w", err)
                continue
            }
            *sc = *newClient
        }
    }
    return nil, fmt.Errorf("重试 %d 次后仍失败：%w", maxRetries, lastErr)
}
```

### 5.2 LibraryManager 扩展 (`backend/libraryservice.go`)

在现有 `LibraryManager` 中添加 SMB 扫描支持：

```go
// scanDirectoryWithMetadata 扫描目录（支持本地和 SMB）
func (lm *LibraryManager) scanDirectoryWithMetadata(dirPath string) ([]TrackInfo, error) {
    // 判断是否为 SMB 路径
    if strings.HasPrefix(dirPath, "smb://") {
        return lm.scanSMBDirectory(dirPath)
    }

    // 本地文件系统扫描（原有逻辑）
    var tracks []TrackInfo
    err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            log.Printf("⚠️ 访问 %s 失败：%v", path, err)
            return nil
        }

        if info.IsDir() {
            return nil
        }

        if !IsAudioFile(info.Name()) {
            return nil
        }

        // 提取元数据
        metadata, err := lm.metadataManager.GetMetadata(path)
        if err != nil {
            log.Printf("⚠️ 提取元数据失败 %s：%v", path, err)
            return nil
        }

        // 查找歌词文件
        lyricPath := lm.findLyricFile(path)

        track := TrackInfo{
            Path:      path,
            Filename:  info.Name(),
            Title:     getStringFromMetadata(metadata, "title", info.Name()),
            Artist:    getStringFromMetadata(metadata, "artist", ""),
            Album:     getStringFromMetadata(metadata, "album", ""),
            Duration:  metadata.Duration,
            Size:      info.Size(),
            LyricPath: lyricPath,
        }

        tracks = append(tracks, track)
        return nil
    })

    if err != nil {
        return nil, err
    }

    return tracks, nil
}

// scanSMBDirectory 扫描 SMB 远程目录
func (lm *LibraryManager) scanSMBDirectory(smbURL string) ([]TrackInfo, error) {
    // 解析 SMB URL 并获取配置
    config, err := parseSMBURL(smbURL)
    if err != nil {
        return nil, fmt.Errorf("解析 SMB URL 失败：%w", err)
    }

    // 创建 SMB 客户端
    client, err := NewSMBClient(config)
    if err != nil {
        return nil, err
    }
    defer client.Close()

    var tracks []TrackInfo
    var mu sync.Mutex

    log.Printf("🔍 开始扫描 SMB 音乐库：%s", smbURL)

    // 使用 Walk 遍历远程目录
    err = client.Walk(".", func(path string, info os.FileInfo, err error) error {
        if err != nil {
            log.Printf("⚠️ 访问 %s 失败：%v", path, err)
            return nil // 继续遍历其他文件
        }

        if info.IsDir() {
            return nil
        }

        if !IsAudioFile(info.Name()) {
            return nil
        }

        // 提取元数据（需要先下载到临时文件）
        track, err := lm.extractMetadataFromSMB(client, path, info)
        if err != nil {
            log.Printf("⚠️ 提取元数据失败 %s：%v", path, err)
            return nil
        }

        mu.Lock()
        tracks = append(tracks, track)
        mu.Unlock()

        // 每扫描 100 个文件打印进度
        if len(tracks)%100 == 0 {
            log.Printf("  已扫描 %d 首歌曲...", len(tracks))
        }

        return nil
    })

    if err != nil {
        return nil, fmt.Errorf("扫描 SMB 目录失败：%w", err)
    }

    log.Printf("✓ SMB 扫描完成，发现 %d 首歌曲", len(tracks))
    return tracks, nil
}

// extractMetadataFromSMB 从 SMB 文件提取元数据
func (lm *LibraryManager) extractMetadataFromSMB(client *SMBClientWrapper, remotePath string, info os.FileInfo) (TrackInfo, error) {
    // 下载文件到临时位置以提取元数据
    tmpFile, err := os.CreateTemp("", "music_meta_*.tmp")
    if err != nil {
        return TrackInfo{}, fmt.Errorf("创建临时文件失败：%w", err)
    }
    defer os.Remove(tmpFile.Name())
    defer tmpFile.Close()

    // 读取远程文件内容（带重试）
    data, err := client.ReadFileWithRetry(remotePath, 3)
    if err != nil {
        return TrackInfo{}, err
    }

    // 写入临时文件
    if _, err := tmpFile.Write(data); err != nil {
        return TrackInfo{}, err
    }
    tmpFile.Close()

    // 使用现有元数据管理器提取
    metadata, err := lm.metadataManager.GetMetadata(tmpFile.Name())
    if err != nil {
        // 如果提取失败，使用基本信息
        return TrackInfo{
            Path:     remotePath,
            Filename: info.Name(),
            Duration: 0,
            Size:     info.Size(),
        }, nil
    }

    // 查找歌词文件（在 SMB 共享中）
    lyricPath := lm.findLyricFileInSMB(client, remotePath)

    return TrackInfo{
        Path:      remotePath,
        Filename:  info.Name(),
        Title:     getStringFromMetadata(metadata, "title", info.Name()),
        Artist:    getStringFromMetadata(metadata, "artist", ""),
        Album:     getStringFromMetadata(metadata, "album", ""),
        Duration:  metadata.Duration,
        Size:      info.Size(),
        LyricPath: lyricPath,
    }, nil
}

// findLyricFileInSMB 在 SMB 共享中查找歌词文件
func (lm *LibraryManager) findLyricFileInSMB(client *SMBClientWrapper, trackPath string) string {
    // 尝试在同目录查找 .lrc 文件
    dir := filepath.Dir(trackPath)
    baseName := strings.TrimSuffix(filepath.Base(trackPath), filepath.Ext(trackPath))
    
    lyricCandidates := []string{
        filepath.Join(dir, baseName+".lrc"),
        filepath.Join(dir, baseName+".txt"),
    }

    for _, candidate := range lyricCandidates {
        if _, err := client.Stat(candidate); err == nil {
            return candidate
        }
    }

    return ""
}

// parseSMBURL 解析 SMB URL
// 格式: smb://user:pass@server:port/share/path
// 简化版: smb://server/share/path （凭据单独提供）
func parseSMBURL(url string) (*SMBConfig, error) {
    if !strings.HasPrefix(url, "smb://") {
        return nil, fmt.Errorf("无效的 SMB URL")
    }

    // 移除前缀
    url = strings.TrimPrefix(url, "smb://")
    
    // 简单解析：server/share/path
    parts := strings.SplitN(url, "/", 2)
    if len(parts) < 2 {
        return nil, fmt.Errorf("SMB URL 格式错误，应为 smb://server/share/path")
    }

    serverPart := parts[0]
    path := "/" + parts[1]

    // 解析服务器和端口
    server := serverPart
    port := 445
    if idx := strings.LastIndex(serverPart, ":"); idx != -1 {
        server = serverPart[:idx]
        // TODO: 解析端口号
    }

    // 解析共享名和路径
    pathParts := strings.SplitN(path, "/", 3)
    share := pathParts[1]
    relativePath := "/"
    if len(pathParts) > 2 {
        relativePath = "/" + pathParts[2]
    }

    return &SMBConfig{
        Server: server,
        Port:   port,
        Share:  share,
        Path:   relativePath,
    }, nil
}
```

### 5.3 密码加密工具 (`backend/pkg/crypto/password.go`)

```go
package crypto

import (
    "golang.org/x/crypto/bcrypt"
)

// EncryptPassword 使用 bcrypt 加密密码
func EncryptPassword(password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    return string(hash), nil
}

// VerifyPassword 验证密码
func VerifyPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

---

## 6. 前端 UI 改造

### 6.1 SettingsView.vue 扩展

在 `frontend/src/views/SettingsView.vue` 中添加 SMB 库配置界面（完整代码见文档附件）。

### 6.2 国际化配置

在 `frontend/src/i18n/locales/zh-CN.json` 和 `en-US.json` 中添加翻译键。

---

## 7. 后端服务接口

### 7.1 MusicService 扩展 (`backend/music_service.go`)

添加以下方法：
- `TestSMBConnection`: 测试 SMB 连接
- `AddSMBLibrary`: 添加 SMB 远程音乐库
- `SelectLocalFolder`: 选择本地文件夹（已有）
- `AddLocalLibrary`: 添加本地音乐库（已有）

---

## 8. 安全与优化

### 8.1 密码加密存储

**重要性**：⚠️ 严禁明文存储密码

```go
// 保存时加密
encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// 使用时验证
err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(inputPassword))
```

### 8.2 连接池管理

实现连接池以避免重复建立连接，限制最大连接数（建议 10）。

### 8.3 缓存策略

- **元数据缓存**：内存缓存，24 小时有效期
- **增量扫描**：记录文件哈希，仅扫描变更文件
- **流式读取**：仅读取文件头部提取元数据

### 8.4 错误处理与重试

实现指数退避重试机制，第 2 次重试后自动重连。

### 8.5 内存优化

对于大文件，避免一次性加载到内存，使用临时文件或流式读取。

---

## 9. 实施路线图

### 阶段 1：基础功能（1-2 周）
- 集成 `go-smb2` 库
- 实现 `SMBClientWrapper` 基础功能
- 扩展 `MusicLibrary` 数据结构
- 实现 SMB 目录扫描
- 添加测试连接功能

### 阶段 2：用户体验（1 周）
- 前端添加 SMB 库配置界面
- 实现连接测试功能
- 添加进度提示和错误反馈
- 支持编辑/删除远程库

### 阶段 3：优化增强（1-2 周）
- 实现密码加密存储
- 添加连接池管理
- 实现增量扫描
- 优化元数据提取
- 添加断线重连机制

### 阶段 4：高级功能（可选，2-3 周）
- 支持 NFS/WebDAV 等其他协议
- 实现流式播放
- 多服务器负载均衡
- 带宽限制和 QoS

---

## 10. 注意事项

### 10.1 关键问题

#### ⚠️ 网络延迟
- SMB 远程访问比本地慢 10-100 倍
- 所有扫描操作必须异步执行
- 提供进度提示和取消功能

#### ⚠️ 内存占用
- 提取元数据需下载文件，大文件可能占用大量内存
- 仅读取文件头部（前 1MB）提取元数据
- 设置并发限制（最多同时处理 5 个文件）

#### ⚠️ 并发控制
- 实现连接池，限制最大连接数
- 使用信号量控制并发扫描数量

#### ⚠️ 权限管理
- 使用 bcrypt 加密存储密码
- 不在日志中输出密码

#### ⚠️ FFmpeg 兼容性
- FFmpeg 可能无法直接处理 SMB 路径
- 先下载到临时文件再调用 FFmpeg

### 10.2 跨平台兼容性

| 平台 | 注意事项 |
|------|----------|
| **macOS** | 可能需要授予网络访问权限；首次连接可能弹出防火墙提示 |
| **Windows** | 确保 SMB1 已禁用；某些 NAS 可能需要启用 SMB2/3 |
| **Linux** | 可能需要安装 `cifs-utils`；检查防火墙规则 |

### 10.3 测试清单

- [ ] 连接不同品牌的 NAS（Synology、QNAP、群晖、威联通）
- [ ] 包含特殊字符的文件名（中文、空格、emoji）
- [ ] 深层嵌套目录（>10 层）
- [ ] 超大音乐库（>10000 首歌曲）
- [ ] 网络中断后的重连
- [ ] 错误的凭据处理
- [ ] 共享不存在的情况
- [ ] 只读共享的兼容性
- [ ] 并发访问多个 SMB 库

### 10.4 性能基准

预期性能指标：

| 操作 | 本地库 | SMB 库（千兆局域网） | SMB 库（WiFi） |
|------|--------|---------------------|---------------|
| 扫描 100 首歌 | ~2 秒 | ~10-30 秒 | ~30-60 秒 |
| 扫描 1000 首歌 | ~15 秒 | ~2-5 分钟 | ~5-10 分钟 |
| 提取单首元数据 | ~50ms | ~200-500ms | ~500ms-1s |
| 播放启动延迟 | <100ms | ~500ms-2s | ~1-3s |

---

## 11. 附录

### 11.1 依赖安装

在 `go.mod` 中添加：

```go
require (
    github.com/hirochachacha/go-smb2 v1.1.0
    golang.org/x/crypto v0.17.0
)
```

然后执行：

```bash
go mod tidy
```

### 11.2 常见 NAS SMB 配置

#### Synology DSM
```
启用 SMB：控制面板 > 文件服务 > SMB/AFP/NFS > 启用 SMB
最小 SMB 版本：SMB2
最大 SMB 版本：SMB3
```

#### QNAP QTS
```
启用 SMB：控制台 > 网络与文件服务 > Win/Mac/NFS > Microsoft 网络
启用 SMB 2/3：勾选"启用 SMB 2.0/3.0"
```

### 11.3 故障排查

#### 问题 1：连接超时
```
原因：防火墙阻止 445 端口
解决：检查 NAS 防火墙设置，允许 445 端口入站
```

#### 问题 2：认证失败
```
原因：用户名/密码错误或域配置不正确
解决：
1. 验证凭据是否正确
2. 尝试添加/移除域参数
3. 检查 NAS 的 SMB 认证方式（NTLM vs Kerberos）
```

### 11.4 参考资料

- [go-smb2 官方文档](https://pkg.go.dev/github.com/hirochachacha/go-smb2)
- [SMB 协议规范](https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/)
- [bcrypt 密码加密](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- [Wails v3 文档](https://wails.io/)

---

## 12. 总结

本方案通过集成 `hirochachacha/go-smb2` 库，为 Haoyun Music Player 添加了完整的 NAS/SMB 远程音乐库支持。方案采用渐进式实施策略，从基础功能到高级优化，确保每个阶段都有明确的交付成果。

**核心优势**：
- ✅ 纯 Go 实现，无 CGO 依赖，跨平台兼容性好
- ✅ 透明的用户体验，远程库与本地库操作一致
- ✅ 完善的安全机制，密码加密存储
- ✅ 高性能优化，连接池和缓存策略
- ✅ 健壮的错误处理，断线重连和重试机制

**下一步行动**：
1. 评审本方案，确认技术选型和实施计划
2. 开始阶段 1 的开发工作
3. 准备测试环境（NAS 设备、测试音乐库）
4. 制定详细的测试用例

---

**文档版本**：v1.0  
**最后更新**：2026-04-13  
**作者**：Haoyun Music Player 开发团队
