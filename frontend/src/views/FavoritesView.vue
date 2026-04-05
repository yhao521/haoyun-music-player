<script setup lang="ts">
import { ref, onMounted } from "vue";
import type { HistoryRecord } from "../../bindings/github.com/yhao521/wailsMusicPlay/backend/models";
import { GetFavoriteTracks, AddToPlaylist, ClearPlaylist, PlayIndex } from "../../bindings/github.com/yhao521/wailsMusicPlay/backend/musicservice";

// 喜爱音乐列表
const favorites = ref<HistoryRecord[]>([]);
const isLoading = ref(false);
const error = ref<string>("");

// 格式化时间
const formatDuration = (seconds: number): string => {
  if (seconds === 0) return "--:--";
  const mins = Math.floor(seconds / 60);
  const secs = seconds % 60;
  return `${mins}:${secs.toString().padStart(2, "0")}`;
};

// 格式化文件大小
const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + " " + sizes[i];
};

// 加载喜爱音乐列表
const loadFavorites = async () => {
  isLoading.value = true;
  error.value = "";
  try {
    // 获取前 100 首喜爱音乐
    favorites.value = await GetFavoriteTracks(100);
    console.log("加载喜爱音乐成功:", favorites.value.length, "首");
  } catch (err) {
    console.error("加载喜爱音乐失败:", err);
    error.value = "加载失败，请稍后重试";
  } finally {
    isLoading.value = false;
  }
};

// 播放指定歌曲
const playTrack = async (track: HistoryRecord) => {
  try {
    // 清空当前播放列表
    await ClearPlaylist();
    
    // 添加该歌曲到播放列表
    await AddToPlaylist(track.path);
    
    // 播放第一首（索引 0）
    await PlayIndex(0);
    
    console.log("开始播放:", track.title);
  } catch (err) {
    console.error("播放失败:", err);
    error.value = "播放失败，请重试";
  }
};

// 返回主界面
const goBack = () => {
  window.location.hash = "#/";
};

// 刷新列表
const refreshList = async () => {
  await loadFavorites();
};

// 组件挂载时自动加载数据
onMounted(() => {
  loadFavorites();
});
</script>

<template>
  <div class="favorites-container">
    <!-- 头部导航 -->
    <div class="header">
      <h1>❤️ 喜爱音乐</h1>
      <button class="refresh-btn" @click="refreshList" :disabled="isLoading" title="刷新列表">
        🔄
      </button>
    </div>

    <!-- 错误提示 -->
    <div v-if="error" class="error-message">
      {{ error }}
    </div>

    <!-- 加载状态 -->
    <div v-if="isLoading" class="loading-state">
      <div class="spinner"></div>
      <p>加载中...</p>
    </div>

    <!-- 空状态 -->
    <div v-else-if="favorites.length === 0" class="empty-state">
      <div class="empty-icon">🎵</div>
      <h2>暂无喜爱音乐</h2>
      <p>多听几首歌，它们就会出现在这里哦~</p>
    </div>

    <!-- 歌曲列表 -->
    <div v-else class="tracks-table-container">
      <table class="tracks-table">
        <thead>
          <tr>
            <th class="col-rank">#</th>
            <th class="col-title">歌曲</th>
            <th class="col-artist">艺术家</th>
            <th class="col-album">专辑</th>
            <th class="col-count">次数</th>
            <th class="col-duration">时长</th>
            <th class="col-size">大小</th>
          </tr>
        </thead>
        <tbody>
          <tr 
            v-for="(track, index) in favorites" 
            :key="track.path"
            class="track-row"
            @click="playTrack(track)"
          >
            <td class="col-rank">{{ index + 1 }}</td>
            <td class="col-title">
              <div class="track-name">{{ track.title || '未知歌曲' }}</div>
            </td>
            <td class="col-artist">{{ track.artist || '未知艺术家' }}</td>
            <td class="col-album">{{ track.album || '未知专辑' }}</td>
            <td class="col-count">
              <span class="play-count-badge">{{ track.play_count }}</span>
            </td>
            <td class="col-duration">{{ formatDuration(track.duration) }}</td>
            <td class="col-size">{{ formatFileSize(track.file_size) }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 统计信息 -->
    <div v-if="!isLoading && favorites.length > 0" class="stats-info">
      共 {{ favorites.length }} 首
    </div>
  </div>
</template>

<style scoped>
.favorites-container {
  width: 100%;
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  overflow: hidden;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background: rgba(0, 0, 0, 0.2);
  backdrop-filter: blur(10px);
}

.header h1 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.refresh-btn {
  padding: 6px 10px;
  border: none;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.2);
  color: white;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s ease;
}

.refresh-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

.refresh-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.error-message {
  margin: 10px 16px;
  padding: 8px 12px;
  background: rgba(255, 100, 100, 0.3);
  border-radius: 4px;
  text-align: center;
  font-size: 13px;
}

.loading-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.spinner {
  width: 36px;
  height: 36px;
  border: 3px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.empty-icon {
  font-size: 60px;
  opacity: 0.5;
}

.empty-state h2 {
  margin: 0;
  font-size: 18px;
  opacity: 0.8;
}

.empty-state p {
  margin: 0;
  opacity: 0.6;
  font-size: 13px;
}

.tracks-table-container {
  flex: 1;
  overflow-y: auto;
  padding: 8px 12px;
}

.tracks-table {
  width: 100%;
  border-collapse: collapse;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  overflow: hidden;
  backdrop-filter: blur(10px);
}

.tracks-table thead {
  background: rgba(0, 0, 0, 0.3);
  position: sticky;
  top: 0;
  z-index: 10;
}

.tracks-table th {
  padding: 8px 10px;
  text-align: left;
  font-weight: 600;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: rgba(255, 255, 255, 0.9);
}

.track-row {
  cursor: pointer;
  transition: background 0.15s ease;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.track-row:hover {
  background: rgba(255, 255, 255, 0.12);
}

.track-row:last-child {
  border-bottom: none;
}

.tracks-table td {
  padding: 8px 10px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.95);
}

.col-rank {
  width: 40px;
  text-align: center;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.7);
  font-size: 11px;
}

.col-title {
  min-width: 180px;
  max-width: 250px;
}

.track-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 500;
}

.col-artist {
  min-width: 120px;
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.col-album {
  min-width: 120px;
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.col-count {
  width: 60px;
  text-align: center;
}

.play-count-badge {
  display: inline-block;
  padding: 2px 6px;
  background: rgba(255, 215, 0, 0.25);
  border-radius: 8px;
  font-weight: 600;
  font-size: 11px;
  color: #ffd700;
}

.col-duration {
  width: 65px;
  text-align: center;
  font-family: monospace;
  font-size: 11px;
}

.col-size {
  width: 70px;
  text-align: right;
  font-family: monospace;
  font-size: 11px;
  opacity: 0.8;
}

.stats-info {
  padding: 8px 16px;
  text-align: center;
  background: rgba(0, 0, 0, 0.2);
  font-size: 12px;
  opacity: 0.8;
}

/* 滚动条样式 */
.tracks-table-container::-webkit-scrollbar {
  width: 6px;
}

.tracks-table-container::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.1);
}

.tracks-table-container::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.3);
  border-radius: 3px;
}

.tracks-table-container::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.5);
}
</style>
