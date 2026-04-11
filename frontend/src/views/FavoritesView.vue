<script setup lang="ts">
import { ref, onMounted } from "vue";
import { t } from "../i18n";
import type { HistoryRecord } from "../../bindings/github.com/yhao521/haoyun-music-player/backend/models";
import {
  GetFavoriteTracks,
  AddToPlaylist,
  ClearPlaylist,
  PlayIndex,
} from "../../bindings/github.com/yhao521/haoyun-music-player/backend/musicservice";

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
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + " " + sizes[i];
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
    error.value = t("favorites.loadFailed");
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
    error.value = t("favorites.playFailed");
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
      <div class="header-content">
        <h1>{{ t("favorites.title") }}</h1>
        <span class="track-count"
          >{{ favorites.length }} {{ t("favorites.songs") }}</span
        >
      </div>
      <button
        class="refresh-btn"
        @click="refreshList"
        :disabled="isLoading"
        :title="t('favorites.refreshList')"
      >
        <span class="refresh-icon" :class="{ rotating: isLoading }">🔄</span>
        <span class="refresh-text">{{
          isLoading ? t("favorites.refreshing") : t("common.refresh")
        }}</span>
      </button>
    </div>

    <!-- 错误提示 -->
    <div v-if="error" class="error-message">
      {{ error }}
    </div>

    <!-- 加载状态 -->
    <div v-if="isLoading" class="loading-state">
      <div class="spinner"></div>
      <p>{{ t("common.loading") }}</p>
    </div>

    <!-- 空状态 -->
    <div v-else-if="favorites.length === 0" class="empty-state">
      <div class="empty-icon">🎵</div>
      <h2>{{ t("favorites.noFavorites") }}</h2>
      <p>{{ t("favorites.listenMoreHint") }}</p>
    </div>

    <!-- 歌曲列表 -->
    <div v-else class="tracks-table-container">
      <table class="tracks-table">
        <thead>
          <tr>
            <th class="col-rank">{{ t("favorites.rank") }}</th>
            <th class="col-title">{{ t("favorites.song") }}</th>
            <th class="col-artist">{{ t("favorites.artist") }}</th>
            <th class="col-album">{{ t("favorites.album") }}</th>
            <th class="col-count">{{ t("favorites.count") }}</th>
            <th class="col-duration">{{ t("favorites.duration") }}</th>
            <th class="col-size">{{ t("favorites.size") }}</th>
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
              <div class="track-name">
                {{ track.title || t("favorites.unknownSong") }}
              </div>
            </td>
            <td class="col-artist">
              {{ track.artist || t("favorites.unknownArtist") }}
            </td>
            <td class="col-album">
              {{ track.album || t("favorites.unknownAlbum") }}
            </td>
            <td class="col-count">
              <span class="play-count-badge">{{ track.play_count }}</span>
            </td>
            <td class="col-duration">{{ formatDuration(track.duration) }}</td>
            <td class="col-size">{{ formatFileSize(track.file_size) }}</td>
          </tr>
        </tbody>
      </table>
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
  padding: 12px 20px;
  background: rgba(0, 0, 0, 0.25);
  backdrop-filter: blur(15px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.header-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.header h1 {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  letter-spacing: 0.5px;
  background: linear-gradient(135deg, #fff 0%, #f0f0f0 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.track-count {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.65);
  font-weight: 500;
}

.refresh-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px 20px;
  min-width: 90px;
  height: 36px;
  white-space: nowrap;
  border: 1px solid rgba(255, 255, 255, 0.25);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.12);
  color: white;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  line-height: 1;
  transition: all 0.25s ease;
  backdrop-filter: blur(10px);
  flex-shrink: 0;
}

.refresh-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
  line-height: 1;
  flex-shrink: 0;
  transition: transform 0.3s ease;
}

.refresh-icon.rotating {
  animation: spin 1s linear infinite;
}

.refresh-text {
  display: inline-block;
  font-weight: 500;
  line-height: 1;
  flex-shrink: 0;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.error-message {
  margin: 12px 20px;
  padding: 10px 14px;
  background: rgba(255, 100, 100, 0.2);
  border: 1px solid rgba(255, 100, 100, 0.3);
  border-radius: 8px;
  text-align: center;
  font-size: 13px;
  backdrop-filter: blur(10px);
}

.loading-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid rgba(255, 255, 255, 0.2);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
}

.empty-icon {
  font-size: 64px;
  opacity: 0.4;
}

.empty-state h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  opacity: 0.8;
}

.empty-state p {
  margin: 0;
  opacity: 0.6;
  font-size: 14px;
}

.tracks-table-container {
  flex: 1;
  overflow-y: auto;
  padding: 12px 16px;
}

.tracks-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  background: rgba(255, 255, 255, 0.08);
  border-radius: 12px;
  overflow: hidden;
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.tracks-table thead {
  background: rgba(0, 0, 0, 0.35);
  position: sticky;
  top: 0;
  z-index: 10;
}

.tracks-table th {
  padding: 10px 12px;
  text-align: left;
  font-weight: 600;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.8px;
  color: rgba(255, 255, 255, 0.85);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.track-row {
  cursor: pointer;
  transition: all 0.2s ease;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.track-row:hover {
  background: rgba(255, 255, 255, 0.12);
  transform: scale(1.005);
}

.track-row:last-child {
  border-bottom: none;
}

.tracks-table td {
  padding: 10px 12px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.95);
}

.col-rank {
  width: 45px;
  text-align: center;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.6);
  font-size: 12px;
}

.col-title {
  min-width: 180px;
  max-width: 280px;
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
  width: 65px;
  text-align: center;
}

.play-count-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  height: 24px;
  padding: 2px 8px;
  background: linear-gradient(
    135deg,
    rgba(255, 215, 0, 0.3) 0%,
    rgba(255, 193, 7, 0.2) 100%
  );
  border: 1px solid rgba(255, 215, 0, 0.4);
  border-radius: 12px;
  font-weight: 700;
  font-size: 12px;
  color: #ffd700;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
}

.col-duration {
  width: 70px;
  text-align: center;
  font-family: "SF Mono", "Monaco", "Consolas", monospace;
  font-size: 12px;
  opacity: 0.85;
}

.col-size {
  width: 75px;
  text-align: right;
  font-family: "SF Mono", "Monaco", "Consolas", monospace;
  font-size: 11px;
  opacity: 0.75;
}

/* 滚动条样式 */
.tracks-table-container::-webkit-scrollbar {
  width: 8px;
}

.tracks-table-container::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.1);
  border-radius: 4px;
}

.tracks-table-container::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.25);
  border-radius: 4px;
  transition: background 0.2s ease;
}

.tracks-table-container::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.4);
}
</style>
