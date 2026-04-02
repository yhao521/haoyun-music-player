# 音乐播放器 API 使用指南

## 一、前端调用示例 (TypeScript)

### 1.1 播放控制

```typescript
import { MusicService } from '../bindings/changeme'

// 播放音乐
await MusicService.Play()

// 暂停音乐
await MusicService.Pause()

// 停止播放
await MusicService.Stop()

// 切换播放/暂停
const isPlaying = await MusicService.TogglePlayPause()

// 播放下一首
await MusicService.Next()

// 播放上一首
await MusicService.Previous()

// 播放指定索引的歌曲
await MusicService.PlayIndex(5) // 播放第 6 首

// 设置音量 (0.0 - 1.0)
await MusicService.SetVolume(0.8)

// 获取音量
const volume = await MusicService.GetVolume()

// 设置播放模式
await MusicService.SetPlayMode('loop') // 'order', 'loop', 'random'

// 获取播放模式
const mode = await MusicService.GetPlayMode()

// 检查是否正在播放
const playing = await MusicService.IsPlaying()
```

### 1.2 播放列表管理

```typescript
// 添加到播放列表
await MusicService.AddToPlaylist('/path/to/song.mp3')

// 清空播放列表
await MusicService.ClearPlaylist()

// 获取播放列表
const playlist = await MusicService.GetPlaylist()
console.log(playlist) // ['/path/to/song1.mp3', '/path/to/song2.mp3', ...]
```

### 1.3 音乐库管理

```typescript
// 添加音乐库（打开目录选择对话框）
await MusicService.AddLibrary()

// 获取当前音乐库
const currentLib = await MusicService.GetCurrentLibrary()
console.log(currentLib)
// {
//   name: 'music',
//   path: '/Users/username/Music',
//   created_at: '2024-01-01T00:00:00Z',
//   updated_at: '2024-01-02T00:00:00Z',
//   tracks: [...]
// }

// 切换音乐库
await MusicService.SwitchLibrary('work')

// 刷新当前音乐库（重新扫描）
await MusicService.RefreshLibrary()

// 重命名音乐库
await MusicService.RenameLibrary('new-name')

// 获取所有音乐库名称
const libraries = await MusicService.GetLibraries()
console.log(libraries) // ['music', 'work', ...]

// 设置当前音乐库
await MusicService.SetCurrentLibrary('music')

// 获取当前音乐库的所有音轨路径
const tracks = await MusicService.GetCurrentLibraryTracks()
console.log(tracks) // ['/path/to/song1.mp3', '/path/to/song2.mp3', ...]

// 加载当前音乐库到播放列表并播放
await MusicService.LoadCurrentLibrary()
```

---

## 二、事件监听示例

### 2.1 监听播放状态变化

```typescript
import { EventsOn, EventsOff } from '@wailsio/runtime'

// 组件挂载时监听
EventsOn('playbackStateChanged', (state: string) => {
  console.log('播放状态变化:', state)
  
  switch (state) {
    case 'playing':
      // 更新播放按钮图标为暂停
      playButtonIcon = 'pause'
      break
    case 'paused':
      // 更新播放按钮图标为播放
      playButtonIcon = 'play'
      break
    case 'stopped':
      // 更新播放按钮图标为停止
      playButtonIcon = 'stop'
      break
  }
})

// 组件卸载时移除监听
EventsOff('playbackStateChanged')
```

### 2.2 监听当前歌曲变化

```typescript
EventsOn('currentTrackChanged', (filename: string) => {
  console.log('当前播放:', filename)
  currentTrackName.value = filename
  
  // 更新 UI 显示
  updateNowPlaying(filename)
})
```

### 2.3 监听播放列表更新

```typescript
EventsOn('playlistUpdated', (playlist: string[]) => {
  console.log('播放列表更新:', playlist)
  playlistItems.value = playlist
  
  // 刷新播放列表 UI
  refreshPlaylistUI()
})
```

### 2.4 监听音乐库更新

```typescript
EventsOn('libraryUpdated', (library: MusicLibrary) => {
  console.log('音乐库更新:', library)
  currentLibrary.value = library
  
  // 更新音乐库菜单
  updateLibraryMenu(library)
})
```

---

## 三、Vue3 组件示例

### 3.1 播放器控制组件

```vue
<template>
  <div class="player-controls">
    <button @click="previous">⏮</button>
    <button @click="togglePlay">{{ isPlaying ? '⏸' : '▶' }}</button>
    <button @click="stop">⏹</button>
    <button @click="next">⏭</button>
    
    <input 
      type="range" 
      min="0" 
      max="1" 
      step="0.01" 
      v-model="volume"
      @change="changeVolume"
    />
    
    <select v-model="playMode" @change="changePlayMode">
      <option value="order">顺序播放</option>
      <option value="loop">循环播放</option>
      <option value="random">随机播放</option>
    </select>
    
    <div class="current-track">
      正在播放：{{ currentTrack }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { MusicService } from '../bindings/changeme'
import { EventsOn, EventsOff } from '@wailsio/runtime'

const isPlaying = ref(false)
const volume = ref(0.7)
const playMode = ref('order')
const currentTrack = ref('')

// 播放控制
const togglePlay = async () => {
  try {
    const playing = await MusicService.TogglePlayPause()
    isPlaying.value = playing
  } catch (error) {
    console.error('切换播放失败:', error)
  }
}

const stop = async () => {
  try {
    await MusicService.Stop()
    isPlaying.value = false
  } catch (error) {
    console.error('停止播放失败:', error)
  }
}

const previous = async () => {
  try {
    await MusicService.Previous()
  } catch (error) {
    console.error('上一首失败:', error)
  }
}

const next = async () => {
  try {
    await MusicService.Next()
  } catch (error) {
    console.error('下一首失败:', error)
  }
}

const changeVolume = async () => {
  try {
    await MusicService.SetVolume(volume.value)
  } catch (error) {
    console.error('设置音量失败:', error)
  }
}

const changePlayMode = async () => {
  try {
    await MusicService.SetPlayMode(playMode.value)
  } catch (error) {
    console.error('设置播放模式失败:', error)
  }
}

// 事件监听
EventsOn('playbackStateChanged', (state: string) => {
  isPlaying.value = state === 'playing'
})

EventsOn('currentTrackChanged', (filename: string) => {
  currentTrack.value = filename
})

// 清理
onUnmounted(() => {
  EventsOff('playbackStateChanged')
  EventsOff('currentTrackChanged')
})
</script>

<style scoped>
.player-controls {
  display: flex;
  gap: 10px;
  align-items: center;
  padding: 20px;
}

button {
  padding: 10px 15px;
  font-size: 18px;
  cursor: pointer;
  border: none;
  border-radius: 5px;
  background: #007bff;
  color: white;
}

button:hover {
  background: #0056b3;
}

.current-track {
  margin-left: 20px;
  font-size: 14px;
  color: #666;
}
</style>
```

### 3.2 播放列表组件

```vue
<template>
  <div class="playlist">
    <h3>播放列表</h3>
    <ul>
      <li 
        v-for="(track, index) in playlist" 
        :key="track"
        :class="{ active: index === currentIndex }"
        @click="playIndex(index)"
      >
        {{ getFilename(track) }}
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { MusicService } from '../bindings/changeme'
import { EventsOn, EventsOff } from '@wailsio/runtime'

const playlist = ref<string[]>([])
const currentIndex = ref(-1)

const getFilename = (path: string) => {
  return path.split('/').pop() || path
}

const playIndex = async (index: number) => {
  try {
    await MusicService.PlayIndex(index)
    currentIndex.value = index
  } catch (error) {
    console.error('播放失败:', error)
  }
}

// 加载播放列表
const loadPlaylist = async () => {
  try {
    playlist.value = await MusicService.GetPlaylist()
  } catch (error) {
    console.error('获取播放列表失败:', error)
  }
}

// 事件监听
EventsOn('playlistUpdated', () => {
  loadPlaylist()
})

EventsOn('currentTrackChanged', async () => {
  const index = await MusicService.GetCurrentIndex()
  currentIndex.value = index
})

onMounted(() => {
  loadPlaylist()
})
</script>

<style scoped>
.playlist {
  padding: 20px;
}

ul {
  list-style: none;
  padding: 0;
}

li {
  padding: 10px;
  cursor: pointer;
  border-bottom: 1px solid #eee;
}

li:hover {
  background: #f5f5f5;
}

li.active {
  background: #007bff;
  color: white;
}
</style>
```

### 3.3 音乐库菜单组件

```vue
<template>
  <div class="library-menu">
    <h3>音乐库</h3>
    <button @click="addLibrary">添加音乐库</button>
    <button @click="refreshLibrary">刷新当前音乐库</button>
    
    <ul>
      <li 
        v-for="lib in libraries" 
        :key="lib.name"
        :class="{ active: lib.name === currentLibName }"
        @click="switchLibrary(lib.name)"
      >
        {{ lib.name === currentLibName ? '✓ ' : '' }}{{ lib.name }}
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { MusicService, MusicLibrary } from '../bindings/changeme'
import { EventsOn, EventsOff } from '@wailsio/runtime'

const libraries = ref<MusicLibrary[]>([])
const currentLibName = ref('')

const loadLibraries = async () => {
  try {
    const libs = await MusicService.GetAllLibraries()
    libraries.value = libs
    
    const current = await MusicService.GetCurrentLibrary()
    if (current) {
      currentLibName.value = current.name
    }
  } catch (error) {
    console.error('加载音乐库失败:', error)
  }
}

const addLibrary = async () => {
  try {
    await MusicService.AddLibrary()
    await loadLibraries()
  } catch (error) {
    console.error('添加音乐库失败:', error)
  }
}

const refreshLibrary = async () => {
  try {
    await MusicService.RefreshLibrary()
    await loadLibraries()
  } catch (error) {
    console.error('刷新音乐库失败:', error)
  }
}

const switchLibrary = async (name: string) => {
  try {
    await MusicService.SwitchLibrary(name)
    currentLibName.value = name
    
    // 自动加载音乐库到播放列表
    await MusicService.LoadCurrentLibrary()
  } catch (error) {
    console.error('切换音乐库失败:', error)
  }
}

// 事件监听
EventsOn('libraryUpdated', () => {
  loadLibraries()
})

onMounted(() => {
  loadLibraries()
})
</script>

<style scoped>
.library-menu {
  padding: 20px;
}

button {
  margin: 5px;
  padding: 8px 12px;
  cursor: pointer;
}

ul {
  list-style: none;
  padding: 0;
}

li {
  padding: 8px;
  cursor: pointer;
  margin: 5px 0;
  border-radius: 4px;
}

li:hover {
  background: #f0f0f0;
}

li.active {
  background: #4CAF50;
  color: white;
}
</style>
```

---

## 四、完整应用示例 (App.vue)

```vue
<template>
  <div id="app">
    <header>
      <h1>🎵 Haoyun Music Player</h1>
    </header>
    
    <main>
      <Player />
      <Playlist />
      <LibraryMenu />
    </main>
  </div>
</template>

<script setup lang="ts">
import Player from './components/Player.vue'
import Playlist from './components/Playlist.vue'
import LibraryMenu from './components/LibraryMenu.vue'
</script>

<style>
#app {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
}

header {
  background: #42b983;
  color: white;
  padding: 20px;
}

main {
  padding: 20px;
}
</style>
```

---

## 五、Go 后端测试示例

### 5.1 单元测试

```go
package backend

import (
    "testing"
)

func TestPlaylistManager_AddToPlaylist(t *testing.T) {
    pm := NewPlaylistManager()
    
    err := pm.AddToPlaylist("/path/to/song.mp3")
    if err != nil {
        t.Errorf("添加失败：%v", err)
    }
    
    playlist, _ := pm.GetPlaylist()
    if len(playlist) != 1 {
        t.Errorf("期望播放列表长度为 1, 实际为%d", len(playlist))
    }
}

func TestPlaylistManager_Next(t *testing.T) {
    pm := NewPlaylistManager()
    pm.AddToPlaylist("/path/to/song1.mp3")
    pm.AddToPlaylist("/path/to/song2.mp3")
    pm.AddToPlaylist("/path/to/song3.mp3")
    
    pm.PlayIndex(0)
    
    pm.Next()
    index, _ := pm.GetCurrentIndex()
    if index != 1 {
        t.Errorf("期望索引为 1, 实际为%d", index)
    }
}

func TestLibraryManager_ScanDirectory(t *testing.T) {
    lm := NewLibraryManager()
    
    tracks, err := lm.scanDirectory("/path/to/music")
    if err != nil {
        t.Errorf("扫描失败：%v", err)
    }
    
    if len(tracks) == 0 {
        t.Error("期望扫描到音乐文件")
    }
}
```

### 5.2 集成测试

```go
func TestMusicService_FullPlayback(t *testing.T) {
    ms := NewMusicService()
    
    // 初始化
    ms.Init()
    
    // 添加音乐库
    ms.AddToLibrary("/test/music")
    
    // 加载播放列表
    ms.LoadCurrentLibrary()
    
    // 播放
    ms.Play()
    
    // 验证播放状态
    playing, _ := ms.IsPlaying()
    if !playing {
        t.Error("期望正在播放")
    }
    
    // 下一首
    ms.Next()
    
    // 暂停
    ms.Pause()
    
    playing, _ = ms.IsPlaying()
    if playing {
        t.Error("期望已暂停")
    }
}
```

---

## 六、常见问题解答

### Q1: 如何获取当前播放的歌曲信息？

```typescript
const playlist = await MusicService.GetPlaylist()
const index = await MusicService.GetCurrentIndex()
const currentSong = playlist[index]
```

### Q2: 如何实现播放进度条？

后端需要实现进度事件（待扩展）：
```go
// 在 AudioPlayer 中添加
go func() {
    for ap.isPlaying {
        position := ap.GetPosition()
        duration := ap.GetDuration()
        ap.app.Event.Emit("playbackProgress", map[string]float64{
            "position": position,
            "duration": duration,
        })
        time.Sleep(time.Second)
    }
}()
```

前端监听：
```typescript
EventsOn('playbackProgress', ({ position, duration }) => {
  progress.value = position
  totalDuration.value = duration
})
```

### Q3: 如何处理播放列表为空的情况？

```typescript
try {
  await MusicService.Play()
} catch (error) {
  if (error.includes('播放列表为空')) {
    // 提示用户添加音乐
    alert('播放列表为空，请先添加音乐库')
  }
}
```

### Q4: 如何在后台播放？

Wails 应用默认支持后台播放。当窗口关闭时，Go 后端进程仍在运行，音乐继续播放。

---

## 七、总结

本 API 指南提供了完整的前后端调用示例，包括：

✅ **播放控制 API** - Play, Pause, Stop, Next, Previous 等  
✅ **播放列表 API** - AddToPlaylist, ClearPlaylist, GetPlaylist 等  
✅ **音乐库 API** - AddLibrary, SwitchLibrary, RefreshLibrary 等  
✅ **事件监听** - 播放状态、歌曲变化、播放列表更新等  
✅ **Vue3 组件示例** - Player, Playlist, LibraryMenu  
✅ **测试示例** - 单元测试和集成测试  

使用这些示例，您可以快速构建一个功能完整的音乐播放器前端界面！

---

**文档版本**: v1.0  
**更新日期**: 2026-04-02  
**技术栈**: Wails3 + Vue3 + TypeScript + beep v2
