<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import { Events } from "@wailsio/runtime";
// import {
//   AddToPlaylist,
//   // LoadFile,
//   Next,
//   // OpenFilePicker,
//   Play,
//   PlayIndex,
//   Previous,
//   // Seek,
//   SetVolume,
//   TogglePlayPause,
// } from "../bindings/github.com/yhao521/wailsMusicPlay/backend/musicservice";

// 播放状态
const isPlaying = ref(false);
const currentPosition = ref(0);
const duration = ref(0);
const volume = ref(0.7);
const currentTrack = ref("");
const playlist = ref<string[]>([]);

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
    // const result = await TogglePlayPause();
    // isPlaying.value = result;
  } catch (error) {
    console.error("Failed to toggle play/pause:", error);
  }
};

// 下一首
const next = async () => {
  try {
    // await Next();
  } catch (error) {
    console.error("Failed to play next:", error);
  }
};

// 上一首
const previous = async () => {
  try {
    // await Previous();
  } catch (error) {
    console.error("Failed to play previous:", error);
  }
};

// 调节音量
const setVolume = async (value: number) => {
  try {
    // await SetVolume(value);
  } catch (error) {
    console.error("Failed to set volume:", error);
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
    // await AddToPlaylist(path);
  } catch (error) {
    console.error("Failed to add to playlist:", error);
  }
};

// 播放指定歌曲
const playIndex = async (index: number) => {
  try {
    // await PlayIndex(index);
  } catch (error) {
    console.error("Failed to play index:", error);
  }
};

// 监听事件 - Wails v3 使用 Events.On
const listenToEvents = () => {
  // 监听播放状态变化
  Events.On("playbackStateChanged", (state: any) => {
    isPlaying.value = state === "playing";
  });

  // 监听播放进度
  Events.On("playbackProgress", (data: any) => {
    currentPosition.value = data.position;
    duration.value = data.duration;
  });

  // 监听播放列表更新
  Events.On("playlistUpdated", (tracks: any) => {
    playlist.value = tracks;
  });

  // 监听当前歌曲变化
  Events.On("currentTrackChanged", (track: any) => {
    currentTrack.value = track;
  });

  console.log("Music Player initialized");
};

// 清理事件监听
const cleanupEvents = () => {
  window.runtime.EventsOff("playbackStateChanged");
  window.runtime.EventsOff("playbackProgress");
  window.runtime.EventsOff("playlistUpdated");
  window.runtime.EventsOff("currentTrackChanged");
};

// 生命周期
onMounted(() => {
  listenToEvents();
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
        <h2 class="track-title">{{ currentTrack || "未播放音乐" }}</h2>
        <p class="track-artist">未知艺术家</p>
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

    <!-- 操作按钮 -->
    <div class="actions">
      <button class="action-btn" @click="openFile">📂 打开文件</button>
    </div>

    <!-- 播放列表 -->
    <div class="playlist-section" v-if="playlist.length > 0">
      <h3>播放列表 ({{ playlist.length }})</h3>
      <div class="playlist">
        <div
          v-for="(track, index) in playlist"
          :key="index"
          class="playlist-item"
          :class="{ active: currentTrack === track }"
          @click="playIndex(index)"
        >
          <span class="track-number">{{ index + 1 }}</span>
          <span class="track-name">{{ track.split("/").pop() }}</span>
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
  padding: 20px;
  background: linear-gradient(135deg, #1e3c72 0%, #2a5298 100%);
  color: white;
  font-family:
    -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen, Ubuntu,
    Cantarell, sans-serif;
}

.header {
  text-align: center;
  margin-bottom: 20px;
}

.header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}

.album-art {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 30px;
  padding: 20px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  backdrop-filter: blur(10px);
}

.album-cover {
  width: 80px;
  height: 80px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 40px;
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.2);
}

.track-info {
  flex: 1;
  overflow: hidden;
}

.track-title {
  margin: 0 0 8px 0;
  font-size: 16px;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.track-artist {
  margin: 0;
  font-size: 14px;
  opacity: 0.8;
}

.progress-section {
  margin-bottom: 20px;
}

.time-display {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  opacity: 0.8;
  margin-bottom: 8px;
}

.progress-bar {
  width: 100%;
  height: 6px;
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
  width: 16px;
  height: 16px;
  background: #fff;
  border-radius: 50%;
  cursor: pointer;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.3);
}

.progress-bar::-moz-range-thumb {
  width: 16px;
  height: 16px;
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
  gap: 20px;
  margin-bottom: 25px;
}

.control-btn {
  background: rgba(255, 255, 255, 0.15);
  border: none;
  color: white;
  width: 50px;
  height: 50px;
  border-radius: 50%;
  font-size: 24px;
  cursor: pointer;
  transition: all 0.3s ease;
  backdrop-filter: blur(10px);
}

.control-btn:hover {
  background: rgba(255, 255, 255, 0.25);
  transform: scale(1.1);
}

.control-btn.play-btn {
  width: 65px;
  height: 65px;
  font-size: 28px;
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
  gap: 10px;
  margin-bottom: 20px;
  padding: 0 10px;
}

.volume-icon {
  font-size: 20px;
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
  width: 14px;
  height: 14px;
  background: #fff;
  border-radius: 50%;
  cursor: pointer;
}

.volume-slider::-moz-range-thumb {
  width: 14px;
  height: 14px;
  background: #fff;
  border-radius: 50%;
  cursor: pointer;
  border: none;
}

.actions {
  display: flex;
  justify-content: center;
  margin-bottom: 20px;
}

.action-btn {
  background: rgba(255, 255, 255, 0.2);
  border: none;
  color: white;
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s ease;
  backdrop-filter: blur(10px);
}

.action-btn:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: translateY(-2px);
}

.playlist-section {
  flex: 1;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 15px;
  backdrop-filter: blur(10px);
}

.playlist-section h3 {
  margin: 0 0 15px 0;
  font-size: 16px;
  font-weight: 600;
}

.playlist {
  max-height: 200px;
  overflow-y: auto;
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
