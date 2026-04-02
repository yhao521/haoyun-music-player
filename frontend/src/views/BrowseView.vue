<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import type { MusicLibrary, TrackInfo } from "../../bindings/github.com/yhao521/wailsMusicPlay/backend/models";
import {
  GetLibraries,
  GetCurrentLibrary,
  SwitchLibrary,
  PlayIndex,
  AddToPlaylist,
  ClearPlaylist,
  LoadCurrentLibrary,
} from "../../bindings/github.com/yhao521/wailsMusicPlay/backend/musicservice";

// 音乐库列表
const libraries = ref<string[]>([]);
const currentLibrary = ref<MusicLibrary | null>(null);
const selectedLibrary = ref<string>("");
const tracks = ref<TrackInfo[]>([]);
const isLoading = ref(false);
const searchQuery = ref("");

// 分页相关
const currentPage = ref(1);
const pageSize = ref(50); // 每页显示 50 首歌曲

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

// 加载所有音乐库列表
const loadLibraries = async () => {
  try {
    libraries.value = await GetLibraries();
    if (libraries.value.length > 0) {
      // 默认选择第一个音乐库
      selectedLibrary.value = libraries.value[0];
      await loadLibraryTracks(selectedLibrary.value);
    }
  } catch (error) {
    console.error("加载音乐库列表失败:", error);
  }
};

// 加载指定音乐库的音轨
const loadLibraryTracks = async (libName: string) => {
  isLoading.value = true;
  try {
    // 切换音乐库
    await SwitchLibrary(libName);
    // 获取当前音乐库详情
    currentLibrary.value = await GetCurrentLibrary();
    // 获取音轨列表
    if (currentLibrary.value && currentLibrary.value.tracks) {
      tracks.value = currentLibrary.value.tracks;
    } else {
      tracks.value = [];
    }
    // 重置页码到第一页
    currentPage.value = 1;
  } catch (error) {
    console.error("加载音乐库音轨失败:", error);
    tracks.value = [];
  } finally {
    isLoading.value = false;
  }
};

// 切换音乐库
const handleLibraryChange = async (event: Event) => {
  const target = event.target as HTMLSelectElement;
  selectedLibrary.value = target.value;
  await loadLibraryTracks(selectedLibrary.value);
};

// 播放指定索引的歌曲
const playTrack = async (index: number) => {
  try {
    // 清空当前播放列表
    await ClearPlaylist();
    
    // 将所有音轨添加到播放列表
    for (const track of tracks.value) {
      await AddToPlaylist(track.path);
    }
    
    // 播放指定索引的歌曲
    await PlayIndex(index);
    
    console.log(`开始播放第 ${index + 1} 首歌曲`);
  } catch (error) {
    console.error("播放歌曲失败:", error);
  }
};

// 双击播放
const handleDoubleClick = async (index: number) => {
  await playTrack(index);
};

// 搜索过滤
const filteredTracks = computed(() => {
  if (!searchQuery.value) return tracks.value;
  const query = searchQuery.value.toLowerCase();
  return tracks.value.filter((track) =>
    track.title.toLowerCase().includes(query) ||
    track.artist.toLowerCase().includes(query) ||
    track.album.toLowerCase().includes(query) ||
    track.filename.toLowerCase().includes(query)
  );
});

// 分页计算
const totalPages = computed(() => {
  return Math.ceil(filteredTracks.value.length / pageSize.value);
});

const paginatedTracks = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  const end = start + pageSize.value;
  return filteredTracks.value.slice(start, end);
});

// 页码显示范围
const pageRange = computed(() => {
  const range = 5; // 显示当前页前后 5 个页码
  const start = Math.max(1, currentPage.value - range);
  const end = Math.min(totalPages.value, currentPage.value + range);
  const pages: number[] = [];
  
  for (let i = start; i <= end; i++) {
    pages.push(i);
  }
  
  return pages;
});

// 切换页码
const changePage = (page: number) => {
  if (page < 1 || page > totalPages.value) return;
  currentPage.value = page;
  // 滚动到列表顶部
  setTimeout(() => {
    const container = document.querySelector('.tracks-container');
    if (container) {
      container.scrollTop = 0;
    }
  }, 100);
};

// 上一页/下一页
const prevPage = () => changePage(currentPage.value - 1);
const nextPage = () => changePage(currentPage.value + 1);

// 跳转到第一页/最后一页
const firstPage = () => changePage(1);
const lastPage = () => changePage(totalPages.value);

// 统计信息
const totalDuration = computed(() => {
  return tracks.value.reduce((sum, track) => sum + (track.duration || 0), 0);
});

const totalSize = computed(() => {
  return tracks.value.reduce((sum, track) => sum + (track.size || 0), 0);
});

// 生命周期
onMounted(() => {
  loadLibraries();
});

onUnmounted(() => {
  // 清理逻辑
});
</script>

<template>
  <div class="browse-container">
    <!-- 头部 -->
    <div class="header">
      <h1>🎵 浏览音乐库</h1>
    </div>

    <!-- 控制栏 -->
    <div class="controls-bar">
      <!-- 音乐库选择器 -->
      <div class="library-selector">
        <label>音乐库:</label>
        <select :value="selectedLibrary" @change="handleLibraryChange">
          <option
            v-for="lib in libraries"
            :key="lib"
            :value="lib"
          >
            {{ lib }}
          </option>
          <option v-if="libraries.length === 0" disabled>
            暂无音乐库
          </option>
        </select>
      </div>

      <!-- 搜索框 -->
      <div class="search-box">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="搜索歌曲、艺术家、专辑..."
          class="search-input"
        />
        <span class="search-icon">🔍</span>
      </div>
    </div>

    <!-- 统计信息 -->
    <div class="stats-bar" v-if="currentLibrary">
      <span class="stat-item">
        📁 {{ currentLibrary.name }}
      </span>
      <span class="stat-item">
        🎵 {{ tracks.length }} 首歌曲
      </span>
      <span class="stat-item">
        ⏱️ 总时长：{{ formatDuration(totalDuration) }}
      </span>
      <span class="stat-item">
        💾 总大小：{{ formatFileSize(totalSize) }}
      </span>
      <span class="stat-item">
        📂 路径：{{ currentLibrary.path }}
      </span>
    </div>

    <!-- 加载提示 -->
    <div class="loading" v-if="isLoading">
      <div class="spinner"></div>
      <p>正在加载音乐库...</p>
    </div>

    <!-- 歌曲列表 -->
    <div class="tracks-container" v-else>
      <div class="tracks-header">
        <div class="track-number">#</div>
        <div class="track-title">标题</div>
        <div class="track-artist">艺术家</div>
        <div class="track-album">专辑</div>
        <div class="track-duration">时长</div>
        <div class="track-size">大小</div>
      </div>

      <div
        v-for="(track, index) in paginatedTracks"
        :key="track.path"
        class="track-item"
        :class="{ 'even': index % 2 === 1 }"
        @dblclick="() => handleDoubleClick((currentPage - 1) * pageSize + index)"
        title="双击播放"
      >
        <div class="track-number">{{ (currentPage - 1) * pageSize + index + 1 }}</div>
        <div class="track-title" :title="track.title">
          {{ track.title || track.filename }}
        </div>
        <div class="track-artist">{{ track.artist || "未知" }}</div>
        <div class="track-album">{{ track.album || "未知" }}</div>
        <div class="track-duration">{{ formatDuration(track.duration) }}</div>
        <div class="track-size">{{ formatFileSize(track.size) }}</div>
      </div>

      <!-- 空状态提示 -->
      <div class="empty-state" v-if="filteredTracks.length === 0">
        <div class="empty-icon">🎵</div>
        <p v-if="tracks.length === 0">
          该音乐库中没有歌曲<br/>
          <small>请通过系统托盘菜单添加音乐库</small>
        </p>
        <p v-else>
          没有找到匹配的歌曲<br/>
          <small>尝试其他搜索关键词</small>
        </p>
      </div>
    </div>

    <!-- 分页控件 -->
    <div class="pagination-container" v-if="totalPages > 1">
      <div class="pagination-info">
        显示第 {{ (currentPage - 1) * pageSize + 1 }} - {{ Math.min(currentPage * pageSize, filteredTracks.length) }} 首，共 {{ filteredTracks.length }} 首
      </div>
      
      <div class="pagination-controls">
        <button 
          class="page-btn" 
          @click="firstPage" 
          :disabled="currentPage === 1"
          title="首页"
        >
          ⏮
        </button>
        <button 
          class="page-btn" 
          @click="prevPage" 
          :disabled="currentPage === 1"
          title="上一页"
        >
          ◀
        </button>
        
        <div class="page-numbers">
          <span 
            v-for="page in pageRange" 
            :key="page"
            class="page-number"
            :class="{ active: currentPage === page }"
            @click="changePage(page)"
          >
            {{ page }}
          </span>
        </div>
        
        <button 
          class="page-btn" 
          @click="nextPage" 
          :disabled="currentPage === totalPages"
          title="下一页"
        >
          ▶
        </button>
        <button 
          class="page-btn" 
          @click="lastPage" 
          :disabled="currentPage === totalPages"
          title="末页"
        >
          ⏭
        </button>
      </div>
      
      <div class="page-size-selector">
        <label>每页显示：</label>
        <select v-model="pageSize" @change="currentPage = 1">
          <option :value="20">20 首</option>
          <option :value="50">50 首</option>
          <option :value="100">100 首</option>
          <option :value="200">200 首</option>
        </select>
      </div>
    </div>

    <!-- 底部操作提示 -->
    <div class="footer-hint">
      <p>💡 双击歌曲即可播放 | 使用搜索框快速查找歌曲</p>
    </div>
  </div>
</template>

<style scoped>
.browse-container {
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
  font-size: 28px;
  font-weight: 600;
}

.controls-bar {
  display: flex;
  gap: 20px;
  margin-bottom: 15px;
  align-items: center;
}

.library-selector {
  display: flex;
  align-items: center;
  gap: 10px;
}

.library-selector label {
  font-size: 14px;
  font-weight: 500;
}

.library-selector select {
  padding: 8px 12px;
  border-radius: 6px;
  border: none;
  background: rgba(255, 255, 255, 0.2);
  color: white;
  font-size: 14px;
  cursor: pointer;
  min-width: 150px;
}

.library-selector select option {
  background: #1e3c72;
  color: white;
}

.search-box {
  flex: 1;
  position: relative;
}

.search-input {
  width: 100%;
  padding: 8px 40px 8px 12px;
  border-radius: 6px;
  border: none;
  background: rgba(255, 255, 255, 0.2);
  color: white;
  font-size: 14px;
}

.search-input::placeholder {
  color: rgba(255, 255, 255, 0.6);
}

.search-icon {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 16px;
}

.stats-bar {
  display: flex;
  gap: 20px;
  padding: 12px 15px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  margin-bottom: 15px;
  font-size: 13px;
  backdrop-filter: blur(10px);
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 5px;
  white-space: nowrap;
}

.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
}

.spinner {
  width: 50px;
  height: 50px;
  border: 4px solid rgba(255, 255, 255, 0.2);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 20px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.loading p {
  margin: 0;
  font-size: 14px;
  opacity: 0.8;
}

.tracks-container {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  display: flex;
  flex-direction: column;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  backdrop-filter: blur(10px);
  min-height: 0; /* 关键：允许 flex 子项缩小 */
}

.tracks-header {
  display: flex;
  padding: 12px 15px;
  background: rgba(255, 255, 255, 0.15);
  font-weight: 600;
  font-size: 13px;
  border-bottom: 2px solid rgba(255, 255, 255, 0.2);
}

.track-item {
  display: flex;
  padding: 10px 15px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 13px;
}

.track-item:hover {
  background: rgba(255, 255, 255, 0.15);
}

.track-item.even {
  background: rgba(255, 255, 255, 0.03);
}

.track-number {
  width: 40px;
  text-align: center;
  opacity: 0.6;
  font-size: 12px;
}

.track-title {
  flex: 1.5;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.track-artist {
  flex: 1;
  opacity: 0.8;
}

.track-album {
  flex: 1;
  opacity: 0.7;
}

.track-duration {
  width: 80px;
  text-align: right;
  font-family: "Courier New", monospace;
}

.track-size {
  width: 90px;
  text-align: right;
  opacity: 0.7;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
  flex: 1;
}

.empty-icon {
  font-size: 64px;
  margin-bottom: 20px;
  opacity: 0.5;
}

.empty-state p {
  margin: 0;
  font-size: 16px;
  opacity: 0.8;
  line-height: 1.6;
}

.empty-state small {
  opacity: 0.6;
  font-size: 12px;
}

.footer-hint {
  margin-top: 15px;
  text-align: center;
  font-size: 12px;
  opacity: 0.7;
}

.footer-hint p {
  margin: 0;
}

/* 分页控件样式 */
.pagination-container {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 20px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  margin-top: 15px;
  backdrop-filter: blur(10px);
}

.pagination-info {
  font-size: 13px;
  opacity: 0.9;
  white-space: nowrap;
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-btn {
  background: rgba(255, 255, 255, 0.15);
  border: none;
  color: white;
  width: 36px;
  height: 36px;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.page-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.25);
  transform: scale(1.05);
}

.page-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.page-numbers {
  display: flex;
  gap: 5px;
  align-items: center;
}

.page-number {
  background: rgba(255, 255, 255, 0.1);
  color: white;
  min-width: 36px;
  height: 36px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s ease;
  user-select: none;
}

.page-number:hover {
  background: rgba(255, 255, 255, 0.25);
}

.page-number.active {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.4);
}

.page-size-selector {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.page-size-selector label {
  opacity: 0.9;
}

.page-size-selector select {
  padding: 6px 10px;
  border-radius: 6px;
  border: none;
  background: rgba(255, 255, 255, 0.2);
  color: white;
  font-size: 13px;
  cursor: pointer;
}

.page-size-selector select option {
  background: #1e3c72;
  color: white;
}

/* 滚动条样式 */
.tracks-container::-webkit-scrollbar {
  width: 8px;
}

.tracks-container::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
}

.tracks-container::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 4px;
}

.tracks-container::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}
</style>