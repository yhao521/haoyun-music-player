<script setup lang="ts">
import { ref, onMounted } from "vue";

// 返回主界面
const goBack = () => {
  window.location.hash = "#/main";
};

// 刷新设置（预留功能）
const refreshSettings = () => {
  console.log("刷新设置");
  // TODO: 实现刷新设置功能
};

onMounted(() => {
  console.log("[SettingsView] 设置页面已加载");
});
</script>

<template>
  <div class="settings-container">
    <!-- 顶部标题栏 -->
    <div class="header">
      <!-- <button class="back-btn" @click="goBack" title="返回">
        <span class="back-icon">←</span>
      </button> -->
      
      <div class="title-section">
        <h1 class="title">⚙️ 设置</h1>
      </div>
      
      <button class="refresh-btn" @click="refreshSettings" title="刷新">
        <span class="refresh-icon">🔄</span>
      </button>
    </div>

    <!-- 设置内容区域 -->
    <div class="settings-content">
      <div class="settings-section">
        <h2 class="section-title">通用设置</h2>
        
        <div class="setting-item">
          <label class="setting-label">
            <input type="checkbox" class="setting-checkbox" />
            <span>开机自动启动</span>
          </label>
          <p class="setting-description">应用程序将在系统启动时自动运行</p>
        </div>

        <div class="setting-item">
          <label class="setting-label">
            <input type="checkbox" class="setting-checkbox" checked />
            <span>保持系统唤醒</span>
          </label>
          <p class="setting-description">播放音乐时阻止系统进入睡眠模式</p>
        </div>

        <div class="setting-item">
          <label class="setting-label">主题模式</label>
          <select class="setting-select">
            <option value="auto">跟随系统</option>
            <option value="light">浅色模式</option>
            <option value="dark">深色模式</option>
          </select>
        </div>
      </div>

      <div class="settings-section">
        <h2 class="section-title">播放设置</h2>
        
        <div class="setting-item">
          <label class="setting-label">默认播放模式</label>
          <select class="setting-select">
            <option value="loop">列表循环</option>
            <option value="order">顺序播放</option>
            <option value="random">随机播放</option>
            <option value="single">单曲循环</option>
          </select>
        </div>

        <div class="setting-item">
          <label class="setting-label">
            <input type="checkbox" class="setting-checkbox" checked />
            <span>显示歌词</span>
          </label>
          <p class="setting-description">在主界面显示同步歌词</p>
        </div>

        <div class="setting-item">
          <label class="setting-label">音量</label>
          <input type="range" class="setting-slider" min="0" max="100" value="80" />
          <span class="slider-value">80%</span>
        </div>
      </div>

      <div class="settings-section">
        <h2 class="section-title">媒体键设置</h2>
        
        <div class="setting-item">
          <label class="setting-label">
            <input type="checkbox" class="setting-checkbox" checked />
            <span>启用媒体键控制</span>
          </label>
          <p class="setting-description">允许使用键盘媒体键控制播放</p>
        </div>
      </div>

      <div class="settings-section">
        <h2 class="section-title">关于</h2>
        
        <div class="about-info">
          <p class="app-name">Haoyun Music Player</p>
          <p class="app-version">版本 0.5.0</p>
          <p class="app-desc">基于 Wails v3 + Vue 3 构建的跨平台桌面音乐播放器</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.settings-container {
  width: 100%;
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
  color: #ffffff;
  overflow: hidden;
}

/* 顶部标题栏 */
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  background: rgba(0, 0, 0, 0.3);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  flex-shrink: 0;
}

.back-btn,
.refresh-btn {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: #ffffff;
  padding: 8px 12px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s ease;
  white-space: nowrap;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.back-btn:hover,
.refresh-btn:hover {
  background: rgba(255, 255, 255, 0.2);
  border-color: rgba(255, 255, 255, 0.3);
  transform: translateY(-1px);
}

.back-btn:active,
.refresh-btn:active {
  transform: translateY(0);
}

.back-icon,
.refresh-icon {
  font-size: 16px;
}

.title-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #ffffff;
}

.subtitle {
  margin: 0;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
}

/* 设置内容区域 */
.settings-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  min-height: 0; /* 关键：允许 Flex 子项滚动 */
}

.settings-section {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.section-title {
  margin: 0 0 12px 0;
  font-size: 16px;
  font-weight: 600;
  color: #4fc3f7;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  padding-bottom: 8px;
}

.setting-item {
  margin-bottom: 16px;
}

.setting-item:last-child {
  margin-bottom: 0;
}

.setting-label {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  color: #ffffff;
  cursor: pointer;
  margin-bottom: 4px;
}

.setting-description {
  margin: 4px 0 0 24px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.5);
  line-height: 1.4;
}

.setting-checkbox {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: #4fc3f7;
}

.setting-select {
  flex: 1;
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  color: #ffffff;
  font-size: 14px;
  cursor: pointer;
  outline: none;
  transition: all 0.2s ease;
}

.setting-select:hover {
  border-color: rgba(255, 255, 255, 0.3);
}

.setting-select:focus {
  border-color: #4fc3f7;
  box-shadow: 0 0 0 2px rgba(79, 195, 247, 0.2);
}

.setting-select option {
  background: #1a1a2e;
  color: #ffffff;
}

.setting-slider {
  flex: 1;
  height: 6px;
  -webkit-appearance: none;
  appearance: none;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
  outline: none;
  cursor: pointer;
}

.setting-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 16px;
  height: 16px;
  background: #4fc3f7;
  border-radius: 50%;
  cursor: pointer;
  transition: all 0.2s ease;
}

.setting-slider::-webkit-slider-thumb:hover {
  background: #29b6f6;
  transform: scale(1.1);
}

.slider-value {
  min-width: 40px;
  text-align: right;
  font-size: 14px;
  color: #4fc3f7;
  font-weight: 500;
}

.about-info {
  text-align: center;
  padding: 12px 0;
}

.app-name {
  margin: 0 0 8px 0;
  font-size: 18px;
  font-weight: 600;
  color: #4fc3f7;
}

.app-version {
  margin: 0 0 8px 0;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.7);
}

.app-desc {
  margin: 0;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.5);
  line-height: 1.5;
}

/* 滚动条样式 */
.settings-content::-webkit-scrollbar {
  width: 8px;
}

.settings-content::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 4px;
}

.settings-content::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 4px;
}

.settings-content::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}
</style>
