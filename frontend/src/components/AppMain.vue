<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import { Events } from "@wailsio/runtime";
import {
  AddToPlaylist,
  // LoadFile,
  Next,
  // OpenFilePicker,
  Play,
  PlayIndex,
  Previous,
  // Seek,
  SetVolume,
  TogglePlayPause,
  GetPlaylist,
  SetPlayMode,
  GetPlayMode,
} from "../../bindings/github.com/yhao521/wailsMusicPlay/backend/musicservice";

// TrackInfo 音乐文件信息
interface TrackInfo {
  path: string;
  filename: string;
  title: string;
  artist: string;
  album: string;
  duration: number;
  size: number;
}

// 播放状态
const isPlaying = ref(false);
const currentPosition = ref(0);
const duration = ref(0);
const volume = ref(0.7);
const currentTrack = ref<TrackInfo | null>(null);
const playlist = ref<string[]>([]);
const playMode = ref("order"); // order, loop, random

// 播放模式图标映射
const playModeIcons = {
  order: "🔢",   // 顺序播放
  loop: "🔁",    // 循环播放
  random: "🔀",  // 随机播放
};

// 播放模式中文名称
const playModeNames = {
  order: "顺序播放",
  loop: "循环播放",
  random: "随机播放",
};

// 格式化时间
const formatTime = (seconds: number): string => {
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins}:${secs.toString().padStart(2, "0")}`;
};

// 计算进度百分比
const progressPercent = computed(() => {
  if (duration.value === 0) return 0;
  return (currentPosition.value / duration.value) * 100;
});

// 播放/暂停
const togglePlayPause = async () => {
  try {
    const result = await TogglePlayPause();
    console.log("togglePlayPause.result:", result);
    isPlaying.value = result;
  } catch (error) {
    console.error("Failed to toggle play/pause:", error);
  }
};

// 下一首
const next = async () => {
  isPlaying.value = false;
  try {
    await Next();
  } catch (error) {
    console.error("Failed to play next:", error);
  }
};

// 上一首
const previous = async () => {
  isPlaying.value = false;
  try {
    await Previous();
  } catch (error) {
    console.error("Failed to play previous:", error);
  }
};

// 调节音量
const setVolume = async (value: number) => {
  try {
    await SetVolume(value);
  } catch (error) {
    console.error("Failed to set volume:", error);
  }
};

// 切换播放模式
const togglePlayMode = async () => {
  const modes = ["order", "loop", "random"];
  const currentIndex = modes.indexOf(playMode.value);
  const nextIndex = (currentIndex + 1) % modes.length;
  const nextMode = modes[nextIndex];
  
  try {
    await SetPlayMode(nextMode);
    playMode.value = nextMode;
    console.log(`播放模式已切换为：${playModeNames[nextMode as keyof typeof playModeNames]}`);
  } catch (error) {
    console.error("Failed to set play mode:", error);
  }
};

// 设置指定播放模式
const setPlayMode = async (mode: string) => {
  try {
    await SetPlayMode(mode);
    playMode.value = mode;
    console.log(`播放模式已设置为：${playModeNames[mode as keyof typeof playModeNames]}`);
  } catch (error) {
    console.error("Failed to set play mode:", error);
  }
};

// 跳转进度
const seek = async (value: number) => {
  try {
    // await Seek(value);
  } catch (error) {
    console.error("Failed to seek:", error);
  }
};

// 打开文件
const openFile = async () => {
  try {
    // 注意：当前版本需要通过系统托盘或菜单栏打开文件
    // const path = await OpenFilePicker();
    // if (path && path.length > 0) {
    //   await LoadFile(path[0]);
    //   await Play();
    // }
  } catch (error) {
    console.error("Failed to open file:", error);
  }
};

// 添加到播放列表
const addToPlaylist = async (path: string) => {
  try {
    await AddToPlaylist(path);
  } catch (error) {
    console.error("Failed to add to playlist:", error);
  }
};

// 播放指定歌曲
const playIndex = async (index: number) => {
  try {
    await PlayIndex(index);
  } catch (error) {
    console.error("Failed to play index:", error);
  }
};

// 从完整路径提取文件名
const getFileName = (path: string): string => {
  // 严格检查：确保 path 是有效的字符串
  if (path == null || typeof path !== "string") {
    return "";
  }
  // 如果路径为空字符串，直接返回
  if (path.trim() === "") {
    return "";
  }
  const parts = path.split("/");
  return parts[parts.length - 1] || path;
};

// 安全显示歌曲标题
const displayTrackTitle = (track: TrackInfo | string | null): string => {
  // 如果是字符串，按旧逻辑处理
  if (typeof track === "string") {
    return getFileName(track);
  }

  // 如果是 TrackInfo 对象
  if (track && typeof track === "object") {
    // 优先使用元数据中的 title，如果没有则使用 filename
    if (track.title && track.title.trim() !== "") {
      return track.title;
    }
    if (track.filename && track.filename.trim() !== "") {
      return getFileName(track.filename);
    }
    if (track.path && track.path.trim() !== "") {
      return getFileName(track.path);
    }
  }

  return "未播放音乐";
};

// 获取艺术家信息
const displayArtist = (track: TrackInfo | string | null): string => {
  if (track && typeof track === "object") {
    if (
      track.artist &&
      track.artist.trim() !== "" &&
      track.artist !== "未知艺术家"
    ) {
      return track.artist;
    }
  }
  return "未知艺术家";
};

// 获取专辑信息
const displayAlbum = (track: TrackInfo | string | null): string => {
  if (track && typeof track === "object") {
    if (
      track.album &&
      track.album.trim() !== "" &&
      track.album !== "未知专辑"
    ) {
      return track.album;
    }
  }
  return "未知专辑";
};

// 判断是否为当前播放的歌曲
const isCurrentTrack = (track: string): boolean => {
  // 严格检查：确保 track 是有效的字符串
  if (track == null || typeof track !== "string") {
    return false;
  }
  if (!currentTrack.value) {
    return false;
  }
  // 比较路径
  return currentTrack.value.path === track;
};

// 监听事件 - Wails v3 使用 Events.On
const listenToEvents = () => {
  // 监听播放状态变化
  Events.On("playbackStateChanged", (state: any) => {
    console.debug("playbackStateChanged", state, state.data === "playing");
    isPlaying.value = state.data === "playing";
    console.debug("playbackStateChanged", isPlaying.value);
  });

  // 监听播放进度
  Events.On("playbackProgress", (data: any) => {
    console.debug("playbackProgress", data);
    currentPosition.value = data.position;
    duration.value = data.duration;
  });
  GetPlaylist()
    .then((tracks) => {
      console.log("GetPlaylist", tracks);
      playlist.value = tracks;
    })
    .catch(() => {});
  // 监听播放列表更新
  Events.On("playlistUpdated", (tracks: any) => {
    console.log("playlistUpdated", tracks);
    playlist.value = tracks;
  });

  // 监听当前歌曲变化
  Events.On("currentTrackChanged", (track: any) => {
    console.log("currentTrackChanged", track);

    // 兼容旧版本：如果 track 是字符串，转换为 TrackInfo 对象
    if (typeof track.data === "string") {
      currentTrack.value = {
        path: track,
        filename: getFileName(track),
        title: "",
        artist: "",
        album: "",
        duration: 0,
        size: 0,
      };
    } else if (track && typeof track === "object") {
      // 如果是 TrackInfo 对象，直接使用
      currentTrack.value = track.data as TrackInfo;
    } else {
      currentTrack.value = null;
    }
  });

  console.log("Music Player initialized");
};

// 初始化播放模式
const initPlayMode = async () => {
  try {
    const mode = await GetPlayMode();
    playMode.value = mode;
    console.log(`当前播放模式：${playModeNames[mode as keyof typeof playModeNames]}`);
  } catch (error) {
    console.error("Failed to get play mode:", error);
    playMode.value = "order"; // 默认顺序播放
  }
};

// 清理事件监听
const cleanupEvents = () => {
  Events.Off("playbackStateChanged");
  Events.Off("playbackProgress");
  Events.Off("playlistUpdated");
  Events.Off("currentTrackChanged");
};

// 生命周期
onMounted(() => {
  listenToEvents();
  initPlayMode(); // 初始化播放模式
});

onUnmounted(() => {
  cleanupEvents();
});
</script>

<template>
  <div class="player-container">
    <!-- 头部 -->
    <div class="header">
      <h1>🎵 Haoyun Music</h1>
    </div>

    <!-- 专辑封面区域 -->
    <div class="album-art">
      <div class="album-cover">
        <div class="music-icon">🎵</div>
      </div>
      <div class="track-info">
        <h2 class="track-title">{{ displayTrackTitle(currentTrack) }}</h2>
        <p class="track-artist">{{ displayArtist(currentTrack) }}</p>
        <p class="track-album">{{ displayAlbum(currentTrack) }}</p>
      </div>
    </div>

    <!-- 进度条 -->
    <div class="progress-section">
      <div class="time-display">
        <span>{{ formatTime(currentPosition) }}</span>
        <span>{{ formatTime(duration) }}</span>
      </div>
      <input
        type="range"
        class="progress-bar"
        :value="currentPosition"
        :max="duration || 100"
        @input="seek(Number(($event.target as HTMLInputElement).value))"
      />
    </div>

    <!-- 播放控制 -->
    <div class="controls">
      <button class="control-btn" @click="previous" title="上一首">⏮</button>
      <button
        class="control-btn play-btn"
        @click="togglePlayPause"
        :class="{ playing: isPlaying }"
      >
        {{ isPlaying ? "⏸" : "▶️" }}
      </button>
      <button class="control-btn" @click="next" title="下一首">⏭</button>
    </div>

    <!-- 音量控制 -->
    <div class="volume-section">
      <span class="volume-icon">🔊</span>
      <input
        type="range"
        class="volume-slider"
        :value="volume"
        min="0"
        max="1"
        step="0.01"
        @input="setVolume(Number(($event.target as HTMLInputElement).value))"
      />
    </div>

    <!-- 播放模式选择 -->
    <div class="play-mode-section">
      <button
        class="play-mode-btn"
        @click="togglePlayMode"
        :title="`当前：${playModeNames[playMode as keyof typeof playModeNames]}，点击切换`"
      >
        <span class="mode-icon">{{ playModeIcons[playMode as keyof typeof playModeIcons] }}</span>
        <span class="mode-text">{{ playModeNames[playMode as keyof typeof playModeNames] }}</span>
      </button>
      
      <!-- 快速切换按钮组 -->
      <div class="play-mode-options">
        <button
          v-for="(name, mode) in playModeNames"
          :key="mode"
          class="mode-option-btn"
          :class="{ active: playMode === mode }"
          @click="setPlayMode(mode)"
          :title="name"
        >
          {{ playModeIcons[mode as keyof typeof playModeIcons] }}
        </button>
      </div>
    </div>

    <!-- 操作按钮 -->
    <!--    <div class="actions">
      <button class="action-btn" @click="openFile">📂 打开文件</button>
    </div>
    -->

    <!-- 播放列表 -->
    <div class="playlist-section" v-if="playlist.length > 0">
      <h3>播放列表 ({{ playlist.length }})</h3>
      <div class="playlist">
        <div
          v-for="(track, index) in playlist"
          :key="index"
          class="playlist-item"
          :class="{ active: isCurrentTrack(track) }"
          @click="playIndex(index)"
        >
          <span class="track-number">{{ index + 1 }}</span>
          <span class="track-name">{{ getFileName(track) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.player-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  padding: 15px;
  background: linear-gradient(135deg, #1e3c72 0%, #2a5298 100%);
  color: white;
  font-family:
    -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen, Ubuntu,
    Cantarell, sans-serif;
}

.header {
  text-align: center;
  margin-bottom: 15px;
}

.header h1 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.album-art {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  padding: 15px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 10px;
  backdrop-filter: blur(10px);
}

.album-cover {
  width: 60px;
  height: 60px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.2);
}

.track-info {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 4px;
}

.track-title {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.track-artist {
  margin: 0;
  font-size: 12px;
  opacity: 0.8;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.track-album {
  margin: 0;
  font-size: 11px;
  opacity: 0.6;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.progress-section {
  margin-bottom: 15px;
}

.time-display {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  opacity: 0.8;
  margin-bottom: 6px;
}

.progress-bar {
  width: 100%;
  height: 5px;
  -webkit-appearance: none;
  appearance: none;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
  outline: none;
  cursor: pointer;
}

.progress-bar::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 14px;
  height: 14px;
  background: #fff;
  border-radius: 50%;
  cursor: pointer;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.3);
}

.progress-bar::-moz-range-thumb {
  width: 14px;
  height: 14px;
  background: #fff;
  border-radius: 50%;
  cursor: pointer;
  border: none;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.3);
}

.controls {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 15px;
  margin-bottom: 15px;
}

.control-btn {
  background: rgba(255, 255, 255, 0.15);
  border: none;
  color: white;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  font-size: 20px;
  cursor: pointer;
  transition: all 0.3s ease;
  backdrop-filter: blur(10px);
}

.control-btn:hover {
  background: rgba(255, 255, 255, 0.25);
  transform: scale(1.1);
}

.control-btn.play-btn {
  width: 50px;
  height: 50px;
  font-size: 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
}

.control-btn.play-btn:hover {
  background: linear-gradient(135deg, #7c8eee 0%, #8659ac 100%);
  transform: scale(1.15);
}

.volume-section {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 15px;
  padding: 0 10px;
}

.volume-icon {
  font-size: 18px;
}

.volume-slider {
  flex: 1;
  height: 4px;
  -webkit-appearance: none;
  appearance: none;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 2px;
  outline: none;
  cursor: pointer;
}

.volume-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 12px;
  height: 12px;
  background: #fff;
  border-radius: 50%;
  cursor: pointer;
}

.volume-slider::-moz-range-thumb {
  width: 12px;
  height: 12px;
  background: #fff;
  border-radius: 50%;
  cursor: pointer;
  border: none;
}

/* 播放模式选择器样式 */
.play-mode-section {
  margin-bottom: 15px;
  padding: 0 10px;
}

.play-mode-btn {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px 15px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 8px;
  color: white;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s ease;
  backdrop-filter: blur(10px);
  margin-bottom: 10px;
}

.play-mode-btn:hover {
  background: rgba(255, 255, 255, 0.2);
  border-color: rgba(255, 255, 255, 0.3);
  transform: translateY(-2px);
}

.mode-icon {
  font-size: 20px;
}

.mode-text {
  font-weight: 500;
}

.play-mode-options {
  display: flex;
  gap: 8px;
  justify-content: center;
}

.mode-option-btn {
  flex: 1;
  padding: 8px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  color: white;
  font-size: 18px;
  cursor: pointer;
  transition: all 0.2s ease;
  opacity: 0.6;
}

.mode-option-btn:hover {
  background: rgba(255, 255, 255, 0.15);
  opacity: 0.8;
  transform: scale(1.05);
}

.mode-option-btn.active {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-color: rgba(102, 126, 234, 0.5);
  opacity: 1;
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.3);
}

.playlist-section {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 10px;
  padding: 12px;
  backdrop-filter: blur(10px);
  display: flex;
  flex-direction: column;
}

.playlist-section h3 {
  margin: 0 0 10px 0;
  font-size: 14px;
  font-weight: 600;
  text-align: center;
}

.playlist {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
}

.playlist-item {
  display: flex;
  align-items: center;
  padding: 10px;
  margin-bottom: 5px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.playlist-item:hover {
  background: rgba(255, 255, 255, 0.15);
}

.playlist-item.active {
  background: rgba(102, 126, 234, 0.3);
  border-left: 3px solid #667eea;
}

.track-number {
  font-size: 12px;
  opacity: 0.6;
  margin-right: 10px;
  min-width: 25px;
}

.track-name {
  flex: 1;
  font-size: 13px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 滚动条样式 */
.playlist::-webkit-scrollbar {
  width: 6px;
}

.playlist::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 3px;
}

.playlist::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
}

.playlist::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}
</style>
