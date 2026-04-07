<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import { Events } from "@wailsio/runtime";
import { t } from "../i18n";
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
const playMode = ref("loop"); // order, loop, single, random - 默认为循环播放
const playlistCollapsed = ref(false); // 播放列表折叠状态
const isHoveringPlaylist = ref(false); // 鼠标是否悬停在播放列表区域
const playlistContainerRef = ref<HTMLElement | null>(null); // 播放列表容器引用

// 切换播放列表折叠状态
const togglePlaylist = () => {
  playlistCollapsed.value = !playlistCollapsed.value;
  console.log(
    "播放列表折叠状态:",
    playlistCollapsed.value ? "已折叠" : "已展开",
  );
};

// 定位到当前播放的歌曲
const scrollToCurrentTrack = () => {
  if (!currentTrack.value || !playlistContainerRef.value) {
    console.log("没有当前播放的歌曲或容器未就绪");
    return;
  }

  // 获取当前播放歌曲的索引
  const currentIndex = playlist.value.findIndex(
    (track) => track === currentTrack.value?.path,
  );

  if (currentIndex === -1) {
    console.log("当前歌曲不在播放列表中");
    return;
  }

  console.log(`定位到第 ${currentIndex + 1} 首歌曲`);

  // 使用 nextTick 确保 DOM 已更新
  setTimeout(() => {
    if (!playlistContainerRef.value) return;

    // 获取播放列表容器和当前歌曲项
    const container = playlistContainerRef.value;
    const items = container.querySelectorAll(".playlist-item");
    const currentItem = items[currentIndex];

    if (currentItem) {
      // 滚动到当前歌曲位置，居中显示
      currentItem.scrollIntoView({
        behavior: "smooth",
        block: "center",
      });

      // 添加高亮动画效果
      currentItem.classList.add("highlight-animation");
      setTimeout(() => {
        currentItem.classList.remove("highlight-animation");
      }, 1500);
    }
  }, 50);
};

// 播放模式图标映射
const playModeIcons = {
  order: "🔢", // 顺序播放
  loop: "🔁", // 循环播放
  single: "🔂", // 单曲循环
  random: "🔀", // 随机播放
};

// 播放模式中文名称
const playModeNames = {
  order: () => t("playMode.order"),
  loop: () => t("playMode.loop"),
  single: () => t("playMode.single"),
  random: () => t("playMode.random"),
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
  const modes = ["order", "loop", "single", "random"];
  const currentIndex = modes.indexOf(playMode.value);
  const nextIndex = (currentIndex + 1) % modes.length;
  const nextMode = modes[nextIndex];

  try {
    await SetPlayMode(nextMode);
    playMode.value = nextMode;
    console.log(
      `播放模式已切换为：${playModeNames[nextMode as keyof typeof playModeNames]}`,
    );
  } catch (error) {
    console.error("Failed to set play mode:", error);
  }
};

// 设置指定播放模式
const setPlayMode = async (mode: string) => {
  try {
    await SetPlayMode(mode);
    playMode.value = mode;
    console.log(
      `播放模式已设置为：${playModeNames[mode as keyof typeof playModeNames]}`,
    );
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

  return t("main.notPlaying");
};

// 获取艺术家信息
const displayArtist = (track: TrackInfo | string | null): string => {
  if (track && typeof track === "object") {
    if (
      track.artist &&
      track.artist.trim() !== "" &&
      track.artist !== t("main.unknownArtist")
    ) {
      return track.artist;
    }
  }
  return t("main.unknownArtist");
};

// 获取专辑信息
const displayAlbum = (track: TrackInfo | string | null): string => {
  if (track && typeof track === "object") {
    if (
      track.album &&
      track.album.trim() !== "" &&
      track.album !== t("main.unknownAlbum")
    ) {
      return track.album;
    }
  }
  return t("main.unknownAlbum");
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

  // 获取播放列表
  GetPlaylist()
    .then((tracks) => {
      console.log("初始化播放列表:", tracks, "长度:", tracks.length);
      if (tracks && Array.isArray(tracks)) {
        playlist.value = tracks;
        console.log(`✓ 播放列表已加载，共 ${tracks.length} 首歌曲`);
      } else {
        console.warn("⚠️ 播放列表数据格式异常:", tracks);
      }
    })
    .catch((error) => {
      console.error("❌ 获取播放列表失败:", error);
    });

  // 监听播放列表更新
  Events.On("playlistUpdated", (tracks: any) => {
    console.log("playlistUpdated", tracks);
    if (tracks && Array.isArray(tracks.data)) {
      playlist.value = tracks.data;
    } else if (Array.isArray(tracks)) {
      playlist.value = tracks;
    }
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
    console.log(
      `当前播放模式：${playModeNames[mode as keyof typeof playModeNames]}`,
    );
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
      <h1>{{ t("main.title") }}</h1>
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
      <button class="control-btn" @click="previous" :title="t('main.previousTrack')">⏮</button>
      <button
        class="control-btn play-btn"
        @click="togglePlayPause"
        :class="{ playing: isPlaying }"
      >
        {{ isPlaying ? "⏸" : "▶️" }}
      </button>
      <button class="control-btn" @click="next" :title="t('main.nextTrack')">⏭</button>
      
      <!-- 播放模式按钮 -->
      <button
        class="control-btn mode-btn"
        @click="togglePlayMode"
        :title="`${t('main.currentMode')}${playModeNames[playMode as keyof typeof playModeNames]()}${t('main.clickToSwitch')}`"
      >
        {{ playModeIcons[playMode as keyof typeof playModeIcons] }}
      </button>
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

    <!-- 操作按钮 -->
    <!--    <div class="actions">
      <button class="action-btn" @click="openFile">📂 打开文件</button>
    </div>
    -->

    <!-- 播放列表 -->
    <div
      class="playlist-section"
      v-if="playlist.length > 0"
      :class="{ collapsed: playlistCollapsed }"
      @mouseenter="isHoveringPlaylist = true"
      @mouseleave="isHoveringPlaylist = false"
    >
      <div class="playlist-header" @click="togglePlaylist">
        <h3>{{ t("main.playlist") }} ({{ playlist.length }})</h3>
        <div class="header-actions">
          <!-- 定位到当前歌曲按钮 -->
          <button
            v-if="isHoveringPlaylist && currentTrack"
            class="locate-btn"
            @click.stop="scrollToCurrentTrack"
            :title="t('main.locateCurrent')"
          >
            📍
          </button>
          <!-- 折叠/展开按钮 -->
          <button
            class="collapse-btn"
            :title="playlistCollapsed ? t('main.expandPlaylist') : t('main.collapsePlaylist')"
          >
            {{ playlistCollapsed ? "▼" : "▲" }}
          </button>
        </div>
      </div>
      <div
        class="playlist"
        v-show="!playlistCollapsed"
        ref="playlistContainerRef"
      >
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
  padding: 8px;
  background: linear-gradient(135deg, #1e3c72 0%, #2a5298 100%);
  color: white;
  font-family:
    -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen, Ubuntu,
    Cantarell, sans-serif;
  gap: 6px;
}

.header {
  text-align: center;
  margin-bottom: 2px;
}

.header h1 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.album-art {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
  padding: 6px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  backdrop-filter: blur(10px);
}

.album-cover {
  width: 42px;
  height: 42px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
  flex-shrink: 0;
}

.track-info {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 1px;
  min-width: 0;
}

.track-title {
  margin: 0;
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.2;
}

.track-artist {
  margin: 0;
  font-size: 10px;
  opacity: 0.7;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.2;
}

.track-album {
  margin: 0;
  font-size: 9px;
  opacity: 0.5;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.2;
}

.progress-section {
  margin-bottom: 4px;
}

.time-display {
  display: flex;
  justify-content: space-between;
  font-size: 9px;
  opacity: 0.7;
  margin-bottom: 2px;
}

.progress-bar {
  width: 100%;
  height: 4px;
  -webkit-appearance: none;
  appearance: none;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 2px;
  outline: none;
  cursor: pointer;
}

.progress-bar::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 12px;
  height: 12px;
  background: #fff;
  border-radius: 50%;
  cursor: pointer;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
}

.progress-bar::-moz-range-thumb {
  width: 12px;
  height: 12px;
  background: #fff;
  border-radius: 50%;
  cursor: pointer;
  border: none;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
}

.controls {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.control-btn {
  background: rgba(255, 255, 255, 0.15);
  border: none;
  color: white;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  font-size: 16px;
  cursor: pointer;
  transition: all 0.2s ease;
  backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  justify-content: center;
}

.control-btn:hover {
  background: rgba(255, 255, 255, 0.25);
  transform: scale(1.05);
}

.control-btn.play-btn {
  width: 40px;
  height: 40px;
  font-size: 18px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.4);
}

.control-btn.play-btn:hover {
  background: linear-gradient(135deg, #7c8eee 0%, #8659ac 100%);
  transform: scale(1.1);
}

.control-btn.mode-btn {
  width: 36px;
  height: 36px;
  font-size: 16px;
  background: rgba(102, 126, 234, 0.3);
  border: 1px solid rgba(102, 126, 234, 0.5);
}

.control-btn.mode-btn:hover {
  background: rgba(102, 126, 234, 0.5);
  border-color: rgba(102, 126, 234, 0.7);
  transform: scale(1.1);
}

.volume-section {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-bottom: 4px;
  padding: 0 4px;
}

.volume-icon {
  font-size: 14px;
}

.playlist-section {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  padding: 6px;
  backdrop-filter: blur(10px);
  display: flex;
  flex-direction: column;
  transition: all 0.3s ease;
  position: relative;
}

.playlist-section:hover {
  background: rgba(255, 255, 255, 0.12);
}

.playlist-section.collapsed {
  flex: 0 0 auto;
}

.playlist-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  user-select: none;
  margin-bottom: 4px;
  padding: 2px 4px;
  position: relative;
}

.playlist-header:hover {
  opacity: 0.8;
}

.playlist-header h3 {
  margin: 0;
  font-size: 12px;
  font-weight: 600;
  flex: 1;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.locate-btn {
  background: none;
  border: none;
  color: white;
  font-size: 14px;
  cursor: pointer;
  padding: 1px 4px;
  opacity: 0.7;
  transition: all 0.2s ease;
  animation: pulse 2s infinite;
}

.locate-btn:hover {
  opacity: 1;
  transform: scale(1.3);
}

.collapse-btn {
  background: none;
  border: none;
  color: white;
  font-size: 10px;
  cursor: pointer;
  padding: 1px 4px;
  opacity: 0.6;
  transition: all 0.2s ease;
}

.collapse-btn:hover {
  opacity: 1;
}

@keyframes pulse {
  0%, 100% {
    opacity: 0.7;
    transform: scale(1);
  }
  50% {
    opacity: 1;
    transform: scale(1.1);
  }
}

.playlist {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
  max-height: 320px;
}

.playlist-item {
  display: flex;
  align-items: center;
  padding: 5px 6px;
  margin-bottom: 2px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.playlist-item:hover {
  background: rgba(255, 255, 255, 0.15);
}

.playlist-item.active {
  background: rgba(102, 126, 234, 0.3);
  border-left: 2px solid #667eea;
}

.playlist-item.highlight-animation {
  animation: highlightPulse 1.5s ease-in-out;
}

@keyframes highlightPulse {
  0% {
    background: rgba(102, 126, 234, 0.3);
    transform: scale(1);
  }
  25% {
    background: rgba(102, 126, 234, 0.6);
    transform: scale(1.02);
  }
  50% {
    background: rgba(102, 126, 234, 0.3);
    transform: scale(1);
  }
  75% {
    background: rgba(102, 126, 234, 0.6);
    transform: scale(1.02);
  }
  100% {
    background: rgba(102, 126, 234, 0.3);
    transform: scale(1);
  }
}

.track-number {
  font-size: 9px;
  opacity: 0.5;
  margin-right: 6px;
  min-width: 14px;
}

.track-name {
  flex: 1;
  font-size: 10px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.2;
}

/* 滚动条样式 */
.playlist::-webkit-scrollbar {
  width: 4px;
}

.playlist::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 2px;
}

.playlist::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 2px;
}

.playlist::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}
</style>
