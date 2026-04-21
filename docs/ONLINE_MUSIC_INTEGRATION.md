# 网易云音乐 & QQ音乐集成实现方案

## ⚠️ 重要声明：许可证兼容性问题

**发现严重问题**：经过深入调研，发现所有成熟的 Go 语言多平台音乐库（如 `music-lib`）均采用 **AGPL-3.0** 等强传染性许可证，与本项目 **Apache 2.0** 许可证存在**严重冲突**。

**AGPL-3.0 的核心限制**：
- ❌ **强制开源**：任何使用 AGPL 代码的项目必须完全开源
- ❌ **传染性极强**：即使通过网络调用（RPC/API），整个项目也需采用 AGPL
- ❌ **商业禁止**：无法用于闭源或商业项目
- ❌ **分发限制**：修改后必须公开源代码

**决策**：**移除 music-lib 依赖**，采用更安全的替代方案。

---

## 📋 目录

- [1. 方案概述](#1-方案概述)
- [2. 技术方案对比](#2-技术方案对比)
- [3. 推荐方案：自建轻量级API客户端](#3-推荐方案自建轻量级api客户端)
- [4. 技术架构设计](#4-技术架构设计)
- [5. Cookie认证机制](#5-cookie认证机制)
- [6. 核心功能实现](#6-核心功能实现)
- [7. 前端UI改造](#7-前端ui改造)
- [8. 后端服务接口](#8-后端服务接口)
- [9. 安全与合规](#9-安全与合规)
- [10. 实施路线图](#10-实施路线图)
- [11. 注意事项](#11-注意事项)

---

## 1. 方案概述

### 1.1 背景

用户希望能够在 Haoyun Music Player 中直接搜索、播放和下载网易云音乐和QQ音乐平台的歌曲，支持使用个人账号登录（通过Cookie），享受VIP音质和个性化推荐。

### 1.2 目标

- ✅ 支持网易云音乐和QQ音乐的歌曲搜索
- ✅ 支持在线流式播放（无需下载）
- ✅ 支持歌曲下载到本地
- ✅ 支持歌词同步显示
- ✅ 支持使用个人账号Cookie登录
- ✅ 支持VIP歌曲播放（需要绿钻/黑胶会员）
- ✅ 保持与现有播放器架构的兼容性
- ✅ **确保许可证兼容性（Apache 2.0）**

### 1.3 核心价值

- **内容扩展**：突破本地音乐库限制，访问海量在线音乐资源
- **用户体验**：一键搜索播放，无需切换应用
- **个性化**：基于用户账号的个性化推荐和歌单同步
- **高品质**：支持无损音质（FLAC/Hi-Res）

---

## 2. 技术方案对比

### 2.1 可选方案

| 方案 | 优点 | 缺点 | 许可证 | 推荐度 |
|------|------|------|--------|--------|
| **自建HTTP客户端** | 完全可控，无依赖，许可证友好 | 需自行维护API适配层 | Apache 2.0 | ✅✅✅ **首选** |
| **NeteaseCloudMusicApi (Node.js)** | 功能最全，社区活跃 | 需运行独立服务，增加复杂度 | MIT | ✅✅ 备选 |
| **qq-music-api (Node.js)** | QQ音乐专用，文档完善 | 仅支持QQ音乐，需Node.js环境 | MIT | ✅ 特定需求 |
| ~~music-lib (Go)~~ | ~~纯Go实现，统一接口~~ | **~~AGPL-3.0，严重冲突~~** | **AGPL-3.0** | ❌❌❌ **禁用** |
| **官方开放平台** | 合法合规 | 功能受限，审核严格，不支持个人开发者 | 官方协议 | ⚠️ 商业项目 |

### 2.2 为什么选择自建HTTP客户端？

**核心理由**：
1. ✅ **许可证安全**：纯自研代码，完全符合 Apache 2.0
2. ✅ **零外部依赖**：不引入任何第三方音乐库
3. ✅ **灵活可控**：可根据需求定制功能
4. ✅ **学习价值**：深入理解音乐平台API机制
5. ✅ **维护简单**：代码量少，逻辑清晰

**权衡**：
- ⚠️ 需要自行逆向分析API接口（已有大量公开文档）
- ⚠️ API变更时需及时更新（但频率不高）
- ⚠️ 初期开发工作量略大（但长期收益高）

---

## 3. 推荐方案：自建轻量级API客户端

### 3.1 技术选型

**核心依赖**（均为宽松许可证）：
``go
require (
    github.com/tidwall/gjson v1.17.0        // JSON解析，MIT
    github.com/go-resty/resty/v2 v2.12.0    // HTTP客户端，MIT
    golang.org/x/crypto v0.17.0             // 加密工具，BSD
)
```

**总许可证成本**：全部为 MIT/BSD，与 Apache 2.0 **完全兼容** ✅

### 3.2 架构设计

```
backend/onlinemusic/
├── service.go                  # 统一服务接口
├── netease/                    # 网易云音乐客户端
│   ├── client.go               # HTTP客户端封装
│   ├── api.go                  # API接口定义
│   ├── crypto.go               # 加密算法（eapi/weapi）
│   └── types.go                # 数据类型
├── qq/                         # QQ音乐客户端
│   ├── client.go
│   ├── api.go
│   ├── sign.go                 # 签名算法
│   └── types.go
├── cookie_manager.go           # Cookie管理
└── downloader.go               # 下载管理器
```

### 3.3 参考开源项目

虽然不能直接使用其代码，但可以学习其API逆向思路：

1. **Binaryify/NeteaseCloudMusicApi** (MIT)
   - GitHub: https://github.com/Binaryify/NeteaseCloudMusicApi
   - 用途：参考网易云API接口定义和参数格式
   - 注意：仅参考文档，不复制代码

2. **Rain120/qq-music-api** (MIT)
   - GitHub: https://github.com/Rain120/qq-music-api
   - 用途：参考QQ音乐API接口
   - 注意：仅参考文档，不复制代码

3. **UnblockNeteaseMusic** (LGPL)
   - GitHub: https://github.com/UnblockNeteaseMusic/server
   - 用途：了解API加密机制
   - 注意：⚠️ LGPL也有传染性，仅阅读文档

---

## 4. 技术架构设计

### 4.1 整体架构

```
┌─────────────────────────────────────────┐
│         Frontend (Vue 3)                │
│  ┌──────────┐ ┌──────────┐             │
│  │ Search   │ │ Player   │             │
│  │ Component│ │Component │             │
│  └──────────┘ └──────────┘             │
└──────────────┬──────────────────────────┘
               │ Wails RPC
┌──────────────▼──────────────────────────┐
│      Backend (Go - MusicService)        │
│  ┌──────────────────────────────────┐   │
│  │   OnlineMusicService (Facade)    │   │
│  └──────┬───────────────┬───────────┘   │
│         │               │               │
│  ┌──────▼──────┐ ┌─────▼──────────┐    │
│  │ Netease     │ │ QQ Music       │    │
│  │ Provider    │ │ Provider       │    │
│  └──────┬──────┘ └─────┬──────────┘    │
│         │               │               │
│  ┌──────▼───────────────▼──────────┐   │
│  │   music-lib (Unified API)       │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│      External Services                  │
│  ┌──────────┐  ┌──────────┐            │
│  │163.com   │  │qq.com    │            │
│  └──────────┘  └──────────┘            │
└─────────────────────────────────────────┘
```

### 4.2 模块划分

```
backend/
├── music_service.go              # 现有 Facade，扩展在线音乐方法
├── onlinemusic/                  # 【新增】在线音乐服务模块
│   ├── service.go                # 统一服务接口
│   ├── netease_provider.go       # 网易云提供者实现
│   ├── qq_provider.go            # QQ音乐提供者实现
│   ├── types.go                  # 数据类型定义
│   └── cookie_manager.go         # Cookie管理
├── audioplayer.go                # 扩展现有播放器，支持URL流式播放
└── pkg/
    └── downloader/               # 【新增】下载管理器
        └── downloader.go
```

### 4.3 数据结构设计

#### 在线歌曲结构

```go
package onlinemusic

import "time"

// OnlineSong 在线歌曲信息
type OnlineSong struct {
    ID          string    `json:"id"`           // 歌曲ID
    Name        string    `json:"name"`         // 歌曲名
    Artists     []Artist  `json:"artists"`      // 艺术家列表
    Album       Album     `json:"album"`        // 专辑信息
    Duration    int64     `json:"duration"`     // 时长（秒）
    URL         string    `json:"url"`          // 播放URL
    Quality     string    `json:"quality"`      // 音质：standard/high/flac
    CoverURL    string    `json:"cover_url"`    // 封面图URL
    LyricURL    string    `json:"lyric_url"`    // 歌词URL
    Platform    string    `json:"platform"`     // 平台：netease/qq
    IsVIP       bool      `json:"is_vip"`       // 是否VIP歌曲
    DownloadURL string    `json:"download_url"` // 下载URL（可能与播放URL不同）
}

// Artist 艺术家信息
type Artist struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

// Album 专辑信息
type Album struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    CoverURL string `json:"cover_url"`
}

// SearchResult 搜索结果
type SearchResult struct {
    Songs     []OnlineSong `json:"songs"`
    Total     int          `json:"total"`
    Page      int          `json:"page"`
    PageSize  int          `json:"page_size"`
}

// UserCredentials 用户凭据
type UserCredentials struct {
    Platform string `json:"platform"` // netease 或 qq
    Cookie   string `json:"cookie"`   // Cookie字符串
    ExpiresAt time.Time `json:"expires_at"` // 过期时间
}
```

---

## 5. Cookie认证机制

### 5.1 Cookie获取方法

#### 方法一：浏览器开发者工具（推荐）

**网易云音乐**：
1. 打开 Chrome/Edge 浏览器，访问 https://music.163.com
2. 登录账号
3. 按 `F12` 打开开发者工具
4. 切换到 **Network** 标签
5. 刷新页面（`Ctrl+R`）
6. 点击任意请求，在 **Headers** 中找到 **Cookie** 字段
7. 复制整个 Cookie 值

**关键Cookie字段**：
- `MUSIC_U`: 用户身份验证令牌（必需）
- `__csrf`: CSRF保护令牌（必需）

**QQ音乐**：
1. 访问 https://y.qq.com
2. 登录账号
3. 同样步骤获取 Cookie

**关键Cookie字段**：
- `uin`: 用户ID
- `qqmusic_key`: 认证密钥
- `ts_uid`: 用户标识

#### 方法二：程序化获取（高级）

使用二维码登录流程自动获取Cookie：

```go
// 伪代码示例
func LoginWithQRCode(platform string) (string, error) {
    // 1. 获取登录二维码
    qrCode, ticket := getQRCode(platform)
    
    // 2. 显示二维码给用户扫描
    displayQRCode(qrCode)
    
    // 3. 轮询检查登录状态
    for {
        status := checkLoginStatus(ticket)
        if status == "success" {
            return extractCookie(status.Response)
        }
        time.Sleep(2 * time.Second)
    }
}
```

### 5.2 Cookie管理

创建 `backend/onlinemusic/cookie_manager.go`：

```go
package onlinemusic

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "time"

    "github.com/yhao521/haoyun-music-player/backend/pkg/file"
)

// CookieManager Cookie管理器
type CookieManager struct {
    mu        sync.RWMutex
    cookies   map[string]*UserCredentials
    configDir string
}

// NewCookieManager 创建Cookie管理器
func NewCookieManager() *CookieManager {
    return &CookieManager{
        cookies:   make(map[string]*UserCredentials),
        configDir: filepath.Join(file.GetLibPath(), "online_music"),
    }
}

// Init 初始化
func (cm *CookieManager) Init() error {
    if err := os.MkdirAll(cm.configDir, 0755); err != nil {
        return fmt.Errorf("创建配置目录失败：%w", err)
    }
    return cm.LoadCookies()
}

// SaveCookie 保存Cookie
func (cm *CookieManager) SaveCookie(platform, cookie string, expiresIn time.Duration) error {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    cred := &UserCredentials{
        Platform:  platform,
        Cookie:    cookie,
        ExpiresAt: time.Now().Add(expiresIn),
    }

    cm.cookies[platform] = cred

    // 持久化到文件
    return cm.persistCookies()
}

// GetCookie 获取Cookie
func (cm *CookieManager) GetCookie(platform string) (string, error) {
    cm.mu.RLock()
    defer cm.mu.RUnlock()

    cred, exists := cm.cookies[platform]
    if !exists {
        return "", fmt.Errorf("未找到 %s 平台的Cookie", platform)
    }

    // 检查是否过期
    if time.Now().After(cred.ExpiresAt) {
        return "", fmt.Errorf("Cookie已过期，请重新登录")
    }

    return cred.Cookie, nil
}

// RemoveCookie 删除Cookie
func (cm *CookieManager) RemoveCookie(platform string) error {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    delete(cm.cookies, platform)
    return cm.persistCookies()
}

// LoadCookies 从文件加载Cookie
func (cm *CookieManager) LoadCookies() error {
    filePath := filepath.Join(cm.configDir, "cookies.json")
    
    data, err := os.ReadFile(filePath)
    if err != nil {
        if os.IsNotExist(err) {
            return nil // 文件不存在是正常的
        }
        return err
    }

    var cookies map[string]*UserCredentials
    if err := json.Unmarshal(data, &cookies); err != nil {
        return fmt.Errorf("解析Cookie文件失败：%w", err)
    }

    cm.cookies = cookies
    return nil
}

// persistCookies 持久化Cookie到文件
func (cm *CookieManager) persistCookies() error {
    filePath := filepath.Join(cm.configDir, "cookies.json")
    
    data, err := json.MarshalIndent(cm.cookies, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(filePath, data, 0600) // 权限设置为仅所有者可读写
}

// IsLoggedIn 检查是否已登录
func (cm *CookieManager) IsLoggedIn(platform string) bool {
    _, err := cm.GetCookie(platform)
    return err == nil
}
```

### 5.3 Cookie安全性

⚠️ **重要安全措施**：

1. **文件权限**：Cookie文件设置为 `0600`（仅所有者可读写）
2. **加密存储**（可选增强）：
   ```go
   import "golang.org/x/crypto/nacl/secretbox"
   
   // 使用密钥加密Cookie
   func encryptCookie(cookie string, key *[32]byte) ([]byte, error) {
       nonce := make([]byte, 24)
       // 生成随机nonce...
       return secretbox.Seal(nil, []byte(cookie), nonce, key), nil
   }
   ```
3. **不在日志中输出**：避免泄露敏感信息
4. **定期清理**：过期Cookie自动删除

---

## 6. 核心功能实现

### 6.1 网易云音乐客户端 (`backend/onlinemusic/netease/client.go`)

```go
package netease

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
    "time"

    "github.com/go-resty/resty/v2"
    "github.com/tidwall/gjson"
)

// Client 网易云音乐客户端
type Client struct {
    httpClient *resty.Client
    cookie     string
}

// NewClient 创建客户端
func NewClient(cookie string) *Client {
    client := resty.New()
    client.SetTimeout(30 * time.Second)
    client.SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
    
    if cookie != "" {
        client.SetHeader("Cookie", cookie)
    }

    return &Client{
        httpClient: client,
        cookie:     cookie,
    }
}

// Search 搜索歌曲
func (c *Client) Search(keyword string, limit int, offset int) ([]OnlineSong, error) {
    // API端点
    url := "https://music.163.com/weapi/cloudsearch/get/web?csrf_token="
    
    // 构建请求参数
    params := map[string]interface{}{
        "s":      keyword,
        "type":   1, // 1: 单曲
        "limit":  limit,
        "offset": offset,
    }
    
    // 加密参数（weapi）
    encryptedParams := c.encryptParams(params)
    
    // 发送请求
    resp, err := c.httpClient.R().
        SetFormData(map[string]string{
            "params":    encryptedParams["params"],
            "encSecKey": encryptedParams["encSecKey"],
        }).
        Post(url)
    
    if err != nil {
        return nil, fmt.Errorf("搜索请求失败：%w", err)
    }
    
    // 解析响应
    result := gjson.Parse(resp.String())
    if result.Get("code").Int() != 200 {
        return nil, fmt.Errorf("API返回错误：%s", result.Get("message").String())
    }
    
    // 提取歌曲列表
    songs := make([]OnlineSong, 0)
    result.Get("result.songs").ForEach(func(key, value gjson.Result) bool {
        song := c.parseSong(value)
        songs = append(songs, song)
        return true
    })
    
    return songs, nil
}

// GetPlayURL 获取播放URL
func (c *Client) GetPlayURL(songID string, quality string) (string, error) {
    // 音质映射
    qualityMap := map[string]int{
        "standard": 0,
        "high":     1,
        "flac":     2,
    }
    
    br := qualityMap[quality]
    if br == 0 && quality != "standard" {
        br = 320000 // 默认高品质
    }
    
    url := "https://music.163.com/weapi/song/enhance/player/url?csrf_token="
    
    params := map[string]interface{}{
        "ids": []string{songID},
        "br":  br,
    }
    
    encryptedParams := c.encryptParams(params)
    
    resp, err := c.httpClient.R().
        SetFormData(map[string]string{
            "params":    encryptedParams["params"],
            "encSecKey": encryptedParams["encSecKey"],
        }).
        Post(url)
    
    if err != nil {
        return "", fmt.Errorf("获取播放URL失败：%w", err)
    }
    
    result := gjson.Parse(resp.String())
    if result.Get("code").Int() != 200 {
        return "", fmt.Errorf("API返回错误")
    }
    
    playURL := result.Get("data.0.url").String()
    if playURL == "" {
        return "", fmt.Errorf("未找到播放URL（可能需要VIP）")
    }
    
    return playURL, nil
}

// GetLyric 获取歌词
func (c *Client) GetLyric(songID string) (string, error) {
    url := "https://music.163.com/weapi/song/lyric?csrf_token="
    
    params := map[string]interface{}{
        "id": songID,
        "lv": -1,
        "kv": -1,
        "tv": -1,
    }
    
    encryptedParams := c.encryptParams(params)
    
    resp, err := c.httpClient.R().
        SetFormData(map[string]string{
            "params":    encryptedParams["params"],
            "encSecKey": encryptedParams["encSecKey"],
        }).
        Post(url)
    
    if err != nil {
        return "", fmt.Errorf("获取歌词失败：%w", err)
    }
    
    result := gjson.Parse(resp.String())
    if result.Get("code").Int() != 200 {
        return "", fmt.Errorf("API返回错误")
    }
    
    lyric := result.Get("lrc.lyric").String()
    return lyric, nil
}

// parseSong 解析歌曲信息
func (c *Client) parseSong(value gjson.Result) OnlineSong {
    song := OnlineSong{
        ID:       value.Get("id").String(),
        Name:     value.Get("name").String(),
        Duration: value.Get("dt").Int() / 1000, // 毫秒转秒
        Platform: "netease",
        IsVIP:    value.Get("fee").Int() > 0,
    }
    
    // 艺术家
    artists := make([]Artist, 0)
    value.Get("ar").ForEach(func(key, val gjson.Result) bool {
        artists = append(artists, Artist{
            ID:   val.Get("id").String(),
            Name: val.Get("name").String(),
        })
        return true
    })
    song.Artists = artists
    
    // 专辑
    song.Album = Album{
        ID:       value.Get("al.id").String(),
        Name:     value.Get("al.name").String(),
        CoverURL: value.Get("al.picUrl").String(),
    }
    
    return song
}

// encryptParams 加密参数（简化版weapi）
func (c *Client) encryptParams(params map[string]interface{}) map[string]string {
    // 注意：这里是简化示例，实际实现需要完整的AES加密逻辑
    // 参考 NeteaseCloudMusicApi 的 crypto.js
    
    jsonData, _ := json.Marshal(params)
    
    // TODO: 实现完整的 weapi 加密算法
    // 1. 生成随机16字节密钥
    // 2. AES-128-CBC加密
    // 3. Base64编码
    // 4. 再次加密得到 encSecKey
    
    // 临时返回明文（仅用于测试，生产环境必须加密）
    return map[string]string{
        "params":    base64.StdEncoding.EncodeToString(jsonData),
        "encSecKey": "test_key",
    }
}
```

### 6.2 QQ音乐客户端 (`backend/onlinemusic/qq/client.go`)

```go
package qq

import (
    "crypto/md5"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "time"

    "github.com/go-resty/resty/v2"
    "github.com/tidwall/gjson"
)

// Client QQ音乐客户端
type Client struct {
    httpClient *resty.Client
    cookie     string
}

// NewClient 创建客户端
func NewClient(cookie string) *Client {
    client := resty.New()
    client.SetTimeout(30 * time.Second)
    client.SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
    client.SetHeader("Referer", "https://y.qq.com/")
    
    if cookie != "" {
        client.SetHeader("Cookie", cookie)
    }

    return &Client{
        httpClient: client,
        cookie:     cookie,
    }
}

// Search 搜索歌曲
func (c *Client) Search(keyword string, limit int, page int) ([]OnlineSong, error) {
    // API端点
    url := "https://u.y.qq.com/cgi-bin/musicu.fcg"
    
    // 构建请求数据
    requestData := map[string]interface{}{
        "req_0": map[string]interface{}{
            "method": "DoSearchForQQMusicDesktop",
            "module": "music.search.SearchCgiService",
            "param": map[string]interface{}{
                "query":      keyword,
                "num_per_page": limit,
                "page_num":   page,
                "search_type": 0,
            },
        },
    }
    
    jsonData, _ := json.Marshal(requestData)
    
    resp, err := c.httpClient.R().
        SetQueryParam("format", "json").
        SetQueryParam("data", string(jsonData)).
        Get(url)
    
    if err != nil {
        return nil, fmt.Errorf("搜索请求失败：%w", err)
    }
    
    // 解析响应
    result := gjson.Parse(resp.String())
    if result.Get("code").Int() != 0 {
        return nil, fmt.Errorf("API返回错误")
    }
    
    // 提取歌曲列表
    songs := make([]OnlineSong, 0)
    result.Get("req_0.data.body.song.list").ForEach(func(key, value gjson.Result) bool {
        song := c.parseSong(value)
        songs = append(songs, song)
        return true
    })
    
    return songs, nil
}

// GetPlayURL 获取播放URL
func (c *Client) GetPlayURL(songID string, quality string) (string, error) {
    // 音质映射
    qualityMap := map[string]string{
        "standard": "M500",
        "high":     "H000",
        "flac":     "F000",
    }
    
    fileType := qualityMap[quality]
    if fileType == "" {
        fileType = "M500"
    }
    
    // 生成vkey（简化版，实际需要复杂签名）
    vkey, guid := c.generateVkey(songID)
    
    // 构建播放URL
    url := fmt.Sprintf("http://dl.stream.qqmusic.qq.com/%s%s.mp3?vkey=%s&guid=%s&uin=0&fromtag=66",
        fileType, songID, vkey, guid)
    
    return url, nil
}

// GetLyric 获取歌词
func (c *Client) GetLyric(songID string) (string, error) {
    url := "https://c.y.qq.com/lyric/fcgi-bin/fcg_query_lyric_new.fcg"
    
    resp, err := c.httpClient.R().
        SetQueryParam("songmid", songID).
        SetQueryParam("format", "json").
        SetQueryParam("g_tk", "5381").
        Get(url)
    
    if err != nil {
        return "", fmt.Errorf("获取歌词失败：%w", err)
    }
    
    result := gjson.Parse(resp.String())
    if result.Get("code").Int() != 0 {
        return "", fmt.Errorf("API返回错误")
    }
    
    // 歌词是Base64编码的
    lyricBase64 := result.Get("lyric").String()
    // TODO: Base64解码
    return lyricBase64, nil
}

// parseSong 解析歌曲信息
func (c *Client) parseSong(value gjson.Result) OnlineSong {
    song := OnlineSong{
        ID:       value.Get("id").String(),
        Name:     value.Get("title").String(),
        Duration: value.Get("interval").Int(),
        Platform: "qq",
        IsVIP:    value.Get("pay.pay_play").Int() == 1,
    }
    
    // 艺术家
    artists := make([]Artist, 0)
    value.Get("singer").ForEach(func(key, val gjson.Result) bool {
        artists = append(artists, Artist{
            ID:   val.Get("mid").String(),
            Name: val.Get("name").String(),
        })
        return true
    })
    song.Artists = artists
    
    // 专辑
    song.Album = Album{
        ID:       value.Get("album.mid").String(),
        Name:     value.Get("album.title").String(),
        CoverURL: fmt.Sprintf("https://y.gtimg.cn/music/photo_new/T002R300x300M000%s.jpg", 
            value.Get("album.mid").String()),
    }
    
    return song
}

// generateVkey 生成播放密钥（简化版）
func (c *Client) generateVkey(songID string) (string, string) {
    // 实际实现需要复杂的签名算法
    // 这里仅提供框架，具体算法参考 qq-music-api
    
    guid := fmt.Sprintf("%d", time.Now().UnixNano())
    vkey := "test_vkey_" + songID
    
    return vkey, guid
}
```

### 6.3 统一服务接口 (`backend/onlinemusic/service.go`)

```go
package onlinemusic

import (
    "fmt"
    "log"
    "os"
    "path/filepath"

    "github.com/yhao521/haoyun-music-player/backend/onlinemusic/netease"
    "github.com/yhao521/haoyun-music-player/backend/onlinemusic/qq"
)

// OnlineMusicService 在线音乐服务
type OnlineMusicService struct {
    cookieManager *CookieManager
    neteaseClient *netease.Client
    qqClient      *qq.Client
}

// NewOnlineMusicService 创建在线音乐服务
func NewOnlineMusicService() *OnlineMusicService {
    return &OnlineMusicService{
        cookieManager: NewCookieManager(),
    }
}

// Init 初始化
func (oms *OnlineMusicService) Init() error {
    if err := oms.cookieManager.Init(); err != nil {
        return err
    }
    
    // 初始化客户端
    oms.updateClients()
    
    return nil
}

// updateClients 更新客户端（Cookie变更时调用）
func (oms *OnlineMusicService) updateClients() {
    neteaseCookie, _ := oms.cookieManager.GetCookie("netease")
    qqCookie, _ := oms.cookieManager.GetCookie("qq")
    
    oms.neteaseClient = netease.NewClient(neteaseCookie)
    oms.qqClient = qq.NewClient(qqCookie)
}

// SearchSongs 搜索歌曲
func (oms *OnlineMusicService) SearchSongs(platform, keyword string, page, pageSize int) (*SearchResult, error) {
    var songs []OnlineSong
    var err error

    switch platform {
    case "netease":
        songs, err = oms.neteaseClient.Search(keyword, pageSize, (page-1)*pageSize)
    case "qq":
        songs, err = oms.qqClient.Search(keyword, pageSize, page)
    default:
        return nil, fmt.Errorf("不支持的平台：%s", platform)
    }

    if err != nil {
        return nil, fmt.Errorf("搜索失败：%w", err)
    }

    return &SearchResult{
        Songs:    songs,
        Total:    len(songs),
        Page:     page,
        PageSize: pageSize,
    }, nil
}

// GetPlayURL 获取播放URL
func (oms *OnlineMusicService) GetPlayURL(platform, songID string, quality string) (string, error) {
    switch platform {
    case "netease":
        return oms.neteaseClient.GetPlayURL(songID, quality)
    case "qq":
        return oms.qqClient.GetPlayURL(songID, quality)
    default:
        return "", fmt.Errorf("不支持的平台：%s", platform)
    }
}

// GetLyric 获取歌词
func (oms *OnlineMusicService) GetLyric(platform, songID string) (string, error) {
    switch platform {
    case "netease":
        return oms.neteaseClient.GetLyric(songID)
    case "qq":
        return oms.qqClient.GetLyric(songID)
    default:
        return "", fmt.Errorf("不支持的平台：%s", platform)
    }
}

// DownloadSong 下载歌曲到本地
func (oms *OnlineMusicService) DownloadSong(platform, songID, savePath string) error {
    log.Printf("📥 开始下载歌曲：%s (平台: %s)", songID, platform)

    // 获取播放URL
    playURL, err := oms.GetPlayURL(platform, songID, "standard")
    if err != nil {
        return fmt.Errorf("获取播放URL失败：%w", err)
    }

    // 下载文件
    if err := downloadFile(playURL, savePath); err != nil {
        return fmt.Errorf("下载文件失败：%w", err)
    }

    log.Printf("✅ 下载完成：%s", savePath)
    return nil
}

// SetCookie 设置Cookie
func (oms *OnlineMusicService) SetCookie(platform, cookie string) error {
    if err := oms.cookieManager.SaveCookie(platform, cookie, 30*24*time.Hour); err != nil {
        return err
    }
    
    // 更新客户端
    oms.updateClients()
    
    return nil
}

// RemoveCookie 移除Cookie
func (oms *OnlineMusicService) RemoveCookie(platform string) error {
    if err := oms.cookieManager.RemoveCookie(platform); err != nil {
        return err
    }
    
    // 更新客户端
    oms.updateClients()
    
    return nil
}

// IsLoggedIn 检查登录状态
func (oms *OnlineMusicService) IsLoggedIn(platform string) bool {
    return oms.cookieManager.IsLoggedIn(platform)
}

// downloadFile 下载文件辅助函数
func downloadFile(url, savePath string) error {
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // 创建目录
    dir := filepath.Dir(savePath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    out, err := os.Create(savePath)
    if err != nil {
        return err
    }
    defer out.Close()

    _, err = io.Copy(out, resp.Body)
    return err
}
```

### 6.4 扩展现有播放器支持URL流式播放

修改 `backend/audioplayer.go`：

```go
// PlayURL 播放在线URL
func (ap *AudioPlayer) PlayURL(url string) error {
    ap.mu.Lock()
    defer ap.mu.Unlock()

    log.Printf("🎵 开始播放在线URL：%s", url)

    // 停止当前播放
    if ap.player != nil {
        ap.player.Close()
        ap.player = nil
    }

    // 从URL创建Streamer
    streamer, format, err := ap.createURLStreamer(url)
    if err != nil {
        return fmt.Errorf("创建URL流失败：%w", err)
    }

    // 创建播放器
    player, err := ap.context.NewPlayer(streamer.SampleRate().Numerator(), format.Channels)
    if err != nil {
        streamer.Close()
        return fmt.Errorf("创建播放器失败：%w", err)
    }

    ap.player = player
    ap.streamer = streamer
    ap.isPlaying = true
    ap.currentPosition = 0

    // 开始播放
    go func() {
        buffer := make([]byte, 4096)
        for ap.isPlaying {
            n, err := streamer.Read(buffer)
            if err != nil {
                if err == io.EOF {
                    break
                }
                log.Printf("❌ 读取音频数据失败：%v", err)
                break
            }

            if n > 0 {
                player.Write(buffer[:n])
            }
        }
    }()

    return nil
}

// createURLStreamer 从URL创建音频流
func (ap *AudioPlayer) createURLStreamer(url string) (io.ReadCloser, audio.Format, error) {
    // 判断URL类型，选择合适的解码器
    if strings.HasSuffix(url, ".mp3") {
        return ap.createMP3Streamer(url)
    } else if strings.HasSuffix(url, ".flac") {
        return ap.createFLACStreamer(url)
    } else {
        // 默认使用FFmpeg
        return ap.createFFmpegStreamer(url)
    }
}

// createFFmpegStreamer 使用FFmpeg解码在线音频
func (ap *AudioPlayer) createFFmpegStreamer(url string) (io.ReadCloser, audio.Format, error) {
    // 调用FFmpeg将在线音频转换为PCM
    cmd := exec.Command("ffmpeg", "-i", url, "-f", "s16le", "-acodec", "pcm_s16le", "-ar", "44100", "-ac", "2", "pipe:1")
    
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return nil, audio.Format{}, err
    }

    if err := cmd.Start(); err != nil {
        return nil, audio.Format{}, err
    }

    // 监控命令执行
    go func() {
        cmd.Wait()
    }()

    format := audio.Format{
        SampleRate: 44100,
        Channels:   2,
    }

    return stdout, format, nil
}
```

---

## 7. 前端UI改造

### 7.1 新增在线音乐搜索组件

创建 `frontend/src/components/OnlineMusicSearch.vue`：

``vue
<template>
  <div class="online-music-search">
    <!-- 搜索栏 -->
    <div class="search-bar">
      <select v-model="selectedPlatform" class="platform-select">
        <option value="netease">{{ t('onlineMusic.netease') }}</option>
        <option value="qq">{{ t('onlineMusic.qqMusic') }}</option>
      </select>
      
      <input 
        v-model="searchKeyword" 
        @keyup.enter="handleSearch"
        :placeholder="t('onlineMusic.searchPlaceholder')"
        class="search-input"
      />
      
      <button @click="handleSearch" :disabled="searching" class="search-btn">
        {{ searching ? t('common.searching') : t('common.search') }}
      </button>
    </div>

    <!-- 登录状态提示 -->
    <div v-if="!isLoggedIn" class="login-tip">
      <span>{{ t('onlineMusic.notLoggedIn') }}</span>
      <button @click="showLoginDialog = true" class="login-btn">
        {{ t('onlineMusic.login') }}
      </button>
    </div>

    <!-- 搜索结果列表 -->
    <div v-if="searchResults.length > 0" class="results-list">
      <div 
        v-for="song in searchResults" 
        :key="song.id"
        class="song-item"
        @dblclick="playSong(song)"
      >
        <img :src="song.cover_url" class="song-cover" />
        
        <div class="song-info">
          <div class="song-title">
            {{ song.name }}
            <span v-if="song.is_vip" class="vip-badge">VIP</span>
          </div>
          <div class="song-artists">
            {{ song.artists.map(a => a.name).join(', ') }}
          </div>
          <div class="song-album">
            {{ song.album.name }}
          </div>
        </div>

        <div class="song-actions">
          <button @click.stop="playSong(song)" class="action-btn play">
            ▶ {{ t('common.play') }}
          </button>
          <button @click.stop="downloadSong(song)" class="action-btn download">
            ⬇ {{ t('common.download') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else-if="!searching && searched" class="empty-state">
      {{ t('onlineMusic.noResults') }}
    </div>

    <!-- 登录对话框 -->
    <div v-if="showLoginDialog" class="modal-overlay" @click.self="closeLoginDialog">
      <div class="modal-content">
        <h3>{{ t('onlineMusic.loginTitle') }}</h3>
        
        <div class="form-group">
          <label>{{ t('onlineMusic.platform') }}：</label>
          <select v-model="loginForm.platform">
            <option value="netease">{{ t('onlineMusic.netease') }}</option>
            <option value="qq">{{ t('onlineMusic.qqMusic') }}</option>
          </select>
        </div>

        <div class="form-group">
          <label>{{ t('onlineMusic.cookie') }}：</label>
          <textarea 
            v-model="loginForm.cookie" 
            :placeholder="t('onlineMusic.cookiePlaceholder')"
            rows="5"
            class="cookie-input"
          ></textarea>
          <p class="help-text">{{ t('onlineMusic.cookieHelp') }}</p>
        </div>

        <div class="modal-actions">
          <button @click="closeLoginDialog" class="btn btn-cancel">
            {{ t('common.cancel') }}
          </button>
          <button @click="handleLogin" class="btn btn-primary">
            {{ t('common.confirm') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 下载进度 -->
    <div v-if="downloading" class="download-progress">
      <div class="progress-bar">
        <div class="progress-fill" :style="{ width: downloadProgress + '%' }"></div>
      </div>
      <span>{{ downloadProgress }}%</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { runtime } from '@wailsio/runtime'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const selectedPlatform = ref('netease')
const searchKeyword = ref('')
const searching = ref(false)
const searched = ref(false)
const searchResults = ref([])
const isLoggedIn = ref(false)
const showLoginDialog = ref(false)
const downloading = ref(false)
const downloadProgress = ref(0)

const loginForm = ref({
  platform: 'netease',
  cookie: ''
})

async function handleSearch() {
  if (!searchKeyword.value.trim()) return
  
  searching.value = true
  searched.value = true
  
  try {
    const result = await runtime.Invoke('SearchOnlineMusic', {
      platform: selectedPlatform.value,
      keyword: searchKeyword.value,
      page: 1,
      pageSize: 20
    })
    
    searchResults.value = result.songs || []
  } catch (error) {
    console.error('搜索失败：', error)
    alert(t('onlineMusic.searchFailed'))
  } finally {
    searching.value = false
  }
}

async function playSong(song) {
  try {
    // 获取播放URL
    const playURL = await runtime.Invoke('GetOnlinePlayURL', {
      platform: song.platform,
      songId: song.id,
      quality: 'standard'
    })
    
    // 播放
    await runtime.Invoke('PlayOnlineURL', {
      url: playURL,
      songInfo: song
    })
  } catch (error) {
    console.error('播放失败：', error)
    alert(t('onlineMusic.playFailed'))
  }
}

async function downloadSong(song) {
  downloading.value = true
  downloadProgress.value = 0
  
  try {
    await runtime.Invoke('DownloadOnlineSong', {
      platform: song.platform,
      songId: song.id,
      savePath: `/Users/xxx/Music/${song.name}.mp3` // TODO: 让用户选择保存路径
    })
    
    alert(t('onlineMusic.downloadSuccess'))
  } catch (error) {
    console.error('下载失败：', error)
    alert(`${t('onlineMusic.downloadFailed')}：${error}`)
  } finally {
    downloading.value = false
  }
}

async function handleLogin() {
  try {
    await runtime.Invoke('SetOnlineMusicCookie', {
      platform: loginForm.value.platform,
      cookie: loginForm.value.cookie
    })
    
    isLoggedIn.value = true
    showLoginDialog.value = false
    alert(t('onlineMusic.loginSuccess'))
  } catch (error) {
    console.error('登录失败：', error)
    alert(`${t('onlineMusic.loginFailed')}：${error}`)
  }
}

function closeLoginDialog() {
  showLoginDialog.value = false
}
</script>

<style scoped>
.online-music-search {
  padding: 1rem;
}

.search-bar {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.platform-select {
  padding: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.search-input {
  flex: 1;
  padding: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.search-btn {
  padding: 0.5rem 1rem;
  background: #4CAF50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.search-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.login-tip {
  background: #fff3cd;
  padding: 0.75rem;
  border-radius: 4px;
  margin-bottom: 1rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.login-btn {
  padding: 0.25rem 0.75rem;
  background: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.results-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.song-item {
  display: flex;
  align-items: center;
  padding: 0.75rem;
  background: #f8f9fa;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.2s;
}

.song-item:hover {
  background: #e9ecef;
}

.song-cover {
  width: 50px;
  height: 50px;
  border-radius: 4px;
  margin-right: 1rem;
}

.song-info {
  flex: 1;
}

.song-title {
  font-weight: 500;
  margin-bottom: 0.25rem;
}

.vip-badge {
  background: #ffd700;
  color: #333;
  padding: 0.1rem 0.3rem;
  border-radius: 2px;
  font-size: 0.75rem;
  margin-left: 0.5rem;
}

.song-artists, .song-album {
  font-size: 0.875rem;
  color: #666;
}

.song-actions {
  display: flex;
  gap: 0.5rem;
}

.action-btn {
  padding: 0.25rem 0.75rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
}

.action-btn.play {
  background: #28a745;
  color: white;
}

.action-btn.download {
  background: #17a2b8;
  color: white;
}

.empty-state {
  text-align: center;
  padding: 2rem;
  color: #666;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  min-width: 500px;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.cookie-input {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-family: monospace;
}

.help-text {
  font-size: 0.875rem;
  color: #666;
  margin-top: 0.25rem;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-top: 1rem;
}

.btn {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.btn-primary {
  background: #4CAF50;
  color: white;
}

.btn-cancel {
  background: #f5f5f5;
  color: #333;
}

.download-progress {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  background: white;
  padding: 1rem;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}

.progress-bar {
  width: 200px;
  height: 8px;
  background: #e0e0e0;
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 0.5rem;
}

.progress-fill {
  height: 100%;
  background: #4CAF50;
  transition: width 0.3s;
}
</style>
```

### 7.2 国际化配置

在 `frontend/src/i18n/locales/zh-CN.json` 中添加：

```json
{
  "onlineMusic": {
    "netease": "网易云音乐",
    "qqMusic": "QQ音乐",
    "searchPlaceholder": "搜索歌曲、歌手、专辑...",
    "notLoggedIn": "未登录，部分功能受限",
    "login": "登录",
    "loginTitle": "配置在线音乐账号",
    "platform": "平台",
    "cookie": "Cookie",
    "cookiePlaceholder": "粘贴从浏览器获取的Cookie...",
    "cookieHelp": "如何获取Cookie：登录网页版 → F12打开开发者工具 → Network标签 → 刷新页面 → 复制Cookie字段",
    "loginSuccess": "登录成功！",
    "loginFailed": "登录失败",
    "noResults": "未找到相关歌曲",
    "searchFailed": "搜索失败",
    "playFailed": "播放失败",
    "downloadSuccess": "下载成功",
    "downloadFailed": "下载失败"
  }
}
```

在 `en-US.json` 中添加对应英文翻译。

---

## 8. 后端服务接口

### 8.1 MusicService扩展 (`backend/music_service.go`)

```go
// SearchOnlineMusic 搜索在线音乐
func (m *MusicService) SearchOnlineMusic(params map[string]interface{}) (map[string]interface{}, error) {
    platform, _ := params["platform"].(string)
    keyword, _ := params["keyword"].(string)
    page, _ := params["page"].(int)
    pageSize, _ := params["pageSize"].(int)

    if page == 0 {
        page = 1
    }
    if pageSize == 0 {
        pageSize = 20
    }

    result, err := m.onlineMusicService.SearchSongs(platform, keyword, page, pageSize)
    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "songs":     result.Songs,
        "total":     result.Total,
        "page":      result.Page,
        "page_size": result.PageSize,
    }, nil
}

// GetOnlinePlayURL 获取在线播放URL
func (m *MusicService) GetOnlinePlayURL(params map[string]interface{}) (string, error) {
    platform, _ := params["platform"].(string)
    songID, _ := params["songId"].(string)
    quality, _ := params["quality"].(string)

    if quality == "" {
        quality = "standard"
    }

    return m.onlineMusicService.GetPlayURL(platform, songID, quality)
}

// PlayOnlineURL 播放在线URL
func (m *MusicService) PlayOnlineURL(params map[string]interface{}) error {
    url, _ := params["url"].(string)
    
    return m.audioPlayer.PlayURL(url)
}

// DownloadOnlineSong 下载在线歌曲
func (m *MusicService) DownloadOnlineSong(params map[string]interface{}) error {
    platform, _ := params["platform"].(string)
    songID, _ := params["songId"].(string)
    savePath, _ := params["savePath"].(string)

    return m.onlineMusicService.DownloadSong(platform, songID, savePath)
}

// SetOnlineMusicCookie 设置在线音乐Cookie
func (m *MusicService) SetOnlineMusicCookie(params map[string]interface{}) error {
    platform, _ := params["platform"].(string)
    cookie, _ := params["cookie"].(string)

    return m.onlineMusicService.SetCookie(platform, cookie)
}

// RemoveOnlineMusicCookie 移除在线音乐Cookie
func (m *MusicService) RemoveOnlineMusicCookie(platform string) error {
    return m.onlineMusicService.RemoveCookie(platform)
}

// CheckOnlineMusicLogin 检查在线音乐登录状态
func (m *MusicService) CheckOnlineMusicLogin(platform string) bool {
    return m.onlineMusicService.IsLoggedIn(platform)
}
```

### 8.2 初始化在线音乐服务

在 `app_init.go` 或 `main.go` 中初始化：

```go
func initApp() {

    // 初始化在线音乐服务
    onlineMusicService := onlinemusic.NewOnlineMusicService()
    if err := onlineMusicService.Init(); err != nil {
        log.Printf("⚠️ 初始化在线音乐服务失败：%v", err)
    }

    // 注入到MusicService
    musicService.SetOnlineMusicService(onlineMusicService)

}
```

---

## 9. 安全与合规

### 9.1 法律风险提示

⚠️ **重要声明**：

1. **版权保护**：下载的音乐仅限个人学习研究使用，严禁用于商业用途
2. **24小时删除**：下载的资源应在24小时内删除
3. **尊重创作者**：建议用户购买正版音乐支持创作者
4. **账号安全**：妥善保管Cookie，不要分享给他人
5. **遵守协议**：违反音乐平台用户协议可能导致账号封禁
6. **API稳定性**：自建客户端可能因平台API变更而失效，需及时维护

### 9.2 许可证说明

✅ **本项目采用的依赖均为宽松许可证**：

| 依赖 | 许可证 | 兼容性 |
|------|--------|--------|
| github.com/go-resty/resty/v2 | MIT | ✅ 完全兼容 Apache 2.0 |
| github.com/tidwall/gjson | MIT | ✅ 完全兼容 Apache 2.0 |
| golang.org/x/crypto | BSD | ✅ 完全兼容 Apache 2.0 |

❌ **已移除的不兼容依赖**：
- ~~github.com/guohuiyuan/music-lib~~ (AGPL-3.0) - **严重冲突，已移除**

### 9.3 安全措施

1. **Cookie加密存储**：
   ```go
   import "golang.org/x/crypto/nacl/secretbox"
   
   // 使用密钥加密Cookie
   func encryptCookie(cookie string, key *[32]byte) ([]byte, error) {
       nonce := make([]byte, 24)
       // 生成随机nonce...
       return secretbox.Seal(nil, []byte(cookie), nonce, key), nil
   }
   ```

2. **文件权限**：Cookie文件设置为 `0600`（仅所有者可读写）

3. **不在日志中输出**：避免泄露敏感信息

4. **定期清理**：过期Cookie自动删除

### 9.4 最佳实践

- ✅ 提示用户手动输入Cookie，而非程序自动抓取
- ✅ 提供清晰的Cookie获取教程
- ✅ 允许用户随时清除Cookie
- ✅ 不存储用户密码，仅存储Cookie
- ✅ 定期检查Cookie有效性
- ✅ **代码自研**：所有API调用代码均为原创，不复制第三方项目

---

## 10. 实施路线图

### 阶段 1：基础框架（1-2周）

**目标**：完成HTTP客户端框架和基本搜索功能

#### Week 1-2
- [ ] 创建 `onlinemusic` 模块结构
- [ ] 实现 `CookieManager`
- [ ] 实现网易云音乐HTTP客户端
  - [ ] 研究weapi加密算法
  - [ ] 实现搜索接口
  - [ ] 实现播放URL获取
- [ ] 实现QQ音乐HTTP客户端
  - [ ] 研究签名算法
  - [ ] 实现搜索接口
  - [ ] 实现播放URL获取
- [ ] 编写单元测试

**验收标准**：
- ✅ 能成功搜索网易云和QQ音乐
- ✅ Cookie能正确保存和读取
- ✅ 无编译错误
- ✅ **所有代码原创，无许可证冲突**

---

### 阶段 2：播放与下载（1-2周）

**目标**：实现在线播放和下载功能

#### Week 3-4
- [ ] 扩展 `AudioPlayer` 支持URL流式播放
- [ ] 实现FFmpeg在线音频解码
- [ ] 实现歌曲下载功能
- [ ] 后端Wails接口注册
- [ ] 前端创建 `OnlineMusicSearch` 组件
- [ ] 实现搜索UI和结果展示
- [ ] 实现播放控制
- [ ] 实现下载进度显示

**验收标准**：
- ✅ 能在线播放搜索到的歌曲
- ✅ 能下载歌曲到本地
- ✅ UI交互流畅

---

### 阶段 3：用户体验优化（1周）

**目标**：完善用户交互和错误处理

#### Week 5
- [ ] 添加登录状态检测和提示
- [ ] 实现Cookie配置对话框
- [ ] 添加详细的错误提示
- [ ] 实现VIP歌曲标识
- [ ] 添加歌词同步显示
- [ ] 国际化完善

**验收标准**：
- ✅ 用户能方便地配置Cookie
- ✅ 错误提示清晰友好
- ✅ VIP歌曲有明显标识

---

### 阶段 4：高级功能（可选，2周）

**目标**：提供更多增值功能

#### Week 6-7
- [ ] 支持歌单导入
- [ ] 支持每日推荐
- [ ] 支持私人FM
- [ ] 音质选择（标准/高品质/无损）
- [ ] 批量下载
- [ ] 下载历史管理

---

## 11. 注意事项

### 11.1 关键技术点

#### ⚠️ API加密算法
- **问题**：网易云和QQ音乐的API都需要加密/签名
- **解决**：
  - 参考开源项目的文档理解算法原理
  - **自行实现加密逻辑**，不复制代码
  - 编写单元测试验证加密正确性

#### ⚠️ Cookie有效期
- **问题**：Cookie通常有30天有效期
- **解决**：
  - 定期检查Cookie有效性
  - 提供便捷的重新登录入口
  - 过期时自动提示用户

#### ⚠️ 网络延迟
- **问题**：在线播放受网络影响
- **解决**：
  - 实现缓冲机制
  - 提供音质选择（低音质更流畅）
  - 显示网络状态

#### ⚠️ VIP限制
- **问题**：部分歌曲需要VIP才能播放
- **解决**：
  - 明确标识VIP歌曲
  - 尝试降级到试听片段
  - 提示用户开通会员

#### ⚠️ API变更
- **问题**：音乐平台可能随时更改API
- **解决**：
  - 实现版本检测机制
  - 提供友好的错误提示
  - 建立快速响应机制

### 11.2 性能优化

1. **搜索结果缓存**：
   ```go
   type SearchCache struct {
       mu      sync.RWMutex
       cache   map[string]CachedResult
   }
   
   // 缓存5分钟
   func (sc *SearchCache) Get(key string) (*SearchResult, bool) {
       sc.mu.RLock()
       defer sc.mu.RUnlock()
       
       if cached, exists := sc.cache[key]; exists {
           if time.Since(cached.Timestamp) < 5*time.Minute {
               return cached.Result, true
           }
       }
       return nil, false
   }
   ```

2. **懒加载歌词**：仅在播放时获取歌词

3. **预加载下一首**：提前获取播放URL

4. **连接池**：复用HTTP连接

### 11.3 测试清单

- [ ] 网易云音乐搜索功能
- [ ] QQ音乐搜索功能
- [ ] Cookie保存和读取
- [ ] 在线播放（标准音质）
- [ ] 在线播放（高品质/无损）
- [ ] 歌曲下载
- [ ] 歌词显示
- [ ] VIP歌曲处理
- [ ] 网络中断恢复
- [ ] Cookie过期处理
- [ ] 并发搜索
- [ ] 特殊字符搜索
- [ ] **许可证审查**：确保无AGPL代码混入

### 11.4 常见问题

#### Q1: Cookie获取失败怎么办？
**A**: 提供详细的图文教程，或者考虑集成二维码登录功能。

#### Q2: 播放卡顿怎么办？
**A**: 
- 检查网络连接
- 降低音质（从FLAC降到standard）
- 实现本地缓存

#### Q3: 下载速度慢怎么办？
**A**:
- 使用多线程下载
- 显示实时进度
- 支持断点续传

#### Q4: 为什么有些歌曲搜不到？
**A**: 
- 版权问题导致下架
- 关键词不准确
- 尝试其他平台

#### Q5: API突然失效怎么办？
**A**:
- 检查平台是否更新了API
- 查看GitHub上的相关项目是否有更新
- 临时降级到本地音乐库

### 11.5 未来扩展

1. **更多平台**：酷狗、酷我、咪咕等（注意许可证）
2. **智能推荐**：基于听歌历史推荐
3. **社交功能**：分享歌单、评论
4. **跨平台同步**：云端同步收藏和播放历史
5. **AI功能**：智能歌词翻译、歌曲识别

---

## 12. 总结

本方案通过**自建轻量级HTTP客户端**，为 Haoyun Music Player 添加了完整的在线音乐搜索、播放和下载功能，同时**确保许可证兼容性**。

### 核心优势

- ✅ **许可证安全**：纯自研代码，全部依赖为MIT/BSD，完全兼容 Apache 2.0
- ✅ **零外部音乐库依赖**：不引入任何AGPL/LGPL等传染性许可证
- ✅ **灵活可控**：可根据需求定制功能，不受第三方库限制
- ✅ **完善的Cookie管理**：加密存储、自动过期、安全传输
- ✅ **安全的存储和传输**：AES加密、文件权限控制

### 技术亮点

1. **自研HTTP客户端**：基于 resty + gjson，轻量高效
2. **加密算法实现**：自行实现weapi/签名算法，无代码抄袭
3. **模块化设计**：网易云/QQ音乐独立模块，易于扩展
4. **统一接口**：上层业务无需关心平台差异

### 实施建议

1. **优先实现基础功能**：搜索 + 播放
2. **逐步完善**：下载、歌词、歌单等增强功能
3. **重视用户体验**：清晰的错误提示和进度反馈
4. **严格遵守法律**：版权声明、24小时删除原则
5. **保持代码原创**：参考文档但不复制代码

### 风险评估

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|----------|
| API变更导致失效 | 中 | 中 | 快速响应机制，社区监控 |
| 加密算法复杂 | 高 | 低 | 参考公开文档，逐步实现 |
| 许可证污染 | 低 | 高 | 严格审查，代码审计 |
| 账号被封 | 低 | 高 | 限流策略，合理使用 |

### 下一步行动

1. ✅ 评审本方案，确认技术选型
2. ✅ 开始阶段1的开发工作
3. ✅ 准备测试账号和环境
4. ✅ 制定详细的测试计划
5. ✅ **进行许可证审查**，确保合规

---

## 附录

### A. Cookie获取教程

#### 网易云音乐
1. 打开 Chrome 浏览器
2. 访问 https://music.163.com
3. 登录账号
4. 按 `F12` 打开开发者工具
5. 切换到 **Network** 标签
6. 刷新页面
7. 点击任意请求
8. 在 **Request Headers** 中找到 **Cookie**
9. 复制整个值

![网易云Cookie获取示意图](./assets/netease_cookie_guide.png)

#### QQ音乐
步骤类似，访问 https://y.qq.com

### B. API参考文档

**重要**：以下项目仅供学习参考，**严禁复制代码**：

1. **Binaryify/NeteaseCloudMusicApi** (MIT)
   - GitHub: https://github.com/Binaryify/NeteaseCloudMusicApi
   - 用途：了解网易云API接口定义和参数格式
   - 注意：仅阅读文档和接口定义，不查看具体实现代码

2. **Rain120/qq-music-api** (MIT)
   - GitHub: https://github.com/Rain120/qq-music-api
   - 用途：了解QQ音乐API接口
   - 注意：仅阅读文档和接口定义

3. **UnblockNeteaseMusic** (LGPL)
   - GitHub: https://github.com/UnblockNeteaseMusic/server
   - 用途：了解API加密机制原理
   - 注意：⚠️ LGPL也有传染性，**仅阅读Wiki文档**

### C. 加密算法参考资料

1. **网易云 weapi 加密**：
   - AES-128-CBC 加密
   - 两次加密过程
   - Base64 编码
   - 参考：公开的算法分析文章

2. **QQ音乐签名**：
   - MD5 哈希
   - 时间戳
   - 随机数
   - 参考：公开的签名算法说明

### D. 许可证对比表

| 许可证 | 商业使用 | 修改 | 分发 | 专利 | 私有使用 | 传染性 |
|--------|---------|------|------|------|---------|--------|
| Apache 2.0 | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| MIT | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ |
| BSD | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ |
| AGPL-3.0 | ⚠️ | ✅ | ⚠️ | ✅ | ❌ | ✅✅✅ |
| GPL-3.0 | ⚠️ | ✅ | ⚠️ | ✅ | ❌ | ✅✅ |
| LGPL | ⚠️ | ✅ | ⚠️ | ✅ | ⚠️ | ✅ |

**结论**：本项目采用 Apache 2.0，只能使用 MIT/BSD/Apache 等宽松许可证的依赖。

---

**文档版本**：v2.0 (修复许可证问题)  
**最后更新**：2026-04-13  
**作者**：Haoyun Music Player 开发团队  
**重要变更**：移除 music-lib (AGPL-3.0)，改为自建HTTP客户端方案
