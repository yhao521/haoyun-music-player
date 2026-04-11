<script setup lang="ts">
import { ref, onMounted, watch } from "vue";
import { t, setLocale, getLocale, type Locale } from "../i18n";
import { Events } from "@wailsio/runtime";
import { CompactLibraries } from "../../bindings/github.com/yhao521/haoyun-music-player/backend/musicservice";

// 当前语言
const currentLanguage = ref<Locale>(getLocale());

// 重启提示
const showRestartTip = ref(false);
const restartMessage = ref("");

// 压缩音乐库状态
const isCompacting = ref(false);
const compactResult = ref("");

// 设置项状态
const settings = ref({
  autoLaunch: false,
  keepAwake: true,
  theme: "auto",
  defaultPlayMode: "loop",
  showLyrics: true,
  defaultVolume: 80,
  enableMediaKeys: true,
});

// 返回主界面
const goBack = () => {
  window.location.hash = "#/main";
};

// 刷新设置（预留功能）
const refreshSettings = () => {
  console.log(t("common.refresh"));
  // TODO: 实现刷新设置功能
};

// 切换语言
const changeLanguage = (locale: Locale) => {
  // 更新前端语言
  setLocale(locale);
  currentLanguage.value = locale;

  // 通知后端切换语言并保存配置
  if (Events.Emit) {
    Events.Emit("changeLanguage", locale);
  }

  console.log(`✓ Language changed to: ${locale}`);

  // 显示重启提示
  showRestartTip.value = true;
  restartMessage.value =
    t("settings.languageChangedTip") ||
    "Language changed. Some interfaces require app restart to take full effect.";

  // 5秒后自动隐藏提示
  setTimeout(() => {
    showRestartTip.value = false;
  }, 5000);
};

// 关闭重启提示
const closeRestartTip = () => {
  showRestartTip.value = false;
};

// 重启应用
const restartApp = () => {
  if (Events.Emit) {
    Events.Emit("restartApp", {});
  }
};

// 压缩音乐库文件
const compactLibraries = async () => {
  if (isCompacting.value) return;

  isCompacting.value = true;
  compactResult.value = "";

  try {
    // 调用后端 API
    const response = await CompactLibraries();

    if (response && response !== undefined) {
      const count = response;
      compactResult.value = t("settings.compacted").replace(
        "{count}",
        count.toString(),
      );
      console.log(`✓ 压缩完成：${count} 个音乐库`);
    } else {
      compactResult.value = "压缩失败，请查看日志";
    }
  } catch (error) {
    console.error("压缩音乐库失败:", error);
    compactResult.value = "压缩失败：" + error;
  } finally {
    isCompacting.value = false;

    // 5秒后清除结果消息
    setTimeout(() => {
      compactResult.value = "";
    }, 5000);
  }
};

// 保存其他设置
const saveSetting = (key: string, value: any) => {
  if (Events.Emit) {
    Events.Emit("updateSetting", { [key]: value });
  }
  console.log(`✓ Setting saved: ${key} = ${value}`);
};

// 针对每个字段的独立 watcher 以单独保存
watch(
  () => settings.value.autoLaunch,
  (val) => saveSetting("autoLaunch", val),
);
watch(
  () => settings.value.keepAwake,
  (val) => saveSetting("keepAwake", val),
);
watch(
  () => settings.value.theme,
  (val) => saveSetting("theme", val),
);
watch(
  () => settings.value.defaultPlayMode,
  (val) => saveSetting("defaultPlayMode", val),
);
watch(
  () => settings.value.showLyrics,
  (val) => saveSetting("showLyrics", val),
);
watch(
  () => settings.value.defaultVolume,
  (val) => saveSetting("defaultVolume", val),
);
watch(
  () => settings.value.enableMediaKeys,
  (val) => saveSetting("enableMediaKeys", val),
);

onMounted(() => {
  console.log("[SettingsView] 设置页面已加载");

  // 从后端加载配置
  if (Events.Emit && Events.On) {
    // 先设置监听器
    Events.On("getSettings:response", (response: any) => {
      console.log("从后端加载配置:", response);
      if (response) {
        // 应用加载的配置
        settings.value = {
          autoLaunch: response.autoLaunch ?? settings.value.autoLaunch,
          keepAwake: response.keepAwake ?? settings.value.keepAwake,
          theme: response.theme ?? settings.value.theme,
          defaultPlayMode:
            response.defaultPlayMode ?? settings.value.defaultPlayMode,
          showLyrics: response.showLyrics ?? settings.value.showLyrics,
          defaultVolume: response.defaultVolume ?? settings.value.defaultVolume,
          enableMediaKeys:
            response.enableMediaKeys ?? settings.value.enableMediaKeys,
        };

        // 应用语言设置
        if (response.language) {
          const locale = response.language as Locale;
          setLocale(locale);
          currentLanguage.value = locale;
          console.log(`✓ 已应用语言设置: ${locale}`);
        }
      }
    });

    // 请求配置
    Events.Emit("getSettings", {});
  }

  // 监听语言变化事件（后端发送的重启提示）
  if (Events.On) {
    Events.On("languageChanged", (data: any) => {
      console.log("收到语言切换事件:", data);
      if (data.needRestart) {
        showRestartTip.value = true;
        restartMessage.value = data.message || t("settings.languageChangedTip");

        // 5秒后自动隐藏
        setTimeout(() => {
          showRestartTip.value = false;
        }, 5000);
      }
    });
  }
});
</script>

<template>
  <div class="settings-container">
    <!-- 重启提示 -->
    <div v-if="showRestartTip" class="restart-tip">
      <div class="tip-content">
        <span class="tip-icon">ℹ️</span>
        <span class="tip-text">{{ restartMessage }}</span>
      </div>
      <div class="tip-actions">
        <button class="restart-btn" @click="restartApp">
          {{ t("settings.restartNow") || "立即重启" }}
        </button>
        <button class="close-tip-btn" @click="closeRestartTip">✕</button>
      </div>
    </div>

    <!-- 顶部标题栏 -->
    <div class="header">
      <!-- <button class="back-btn" @click="goBack" :title="t('common.back')">
        <span class="back-icon">←</span>
      </button> -->

      <div class="title-section">
        <h1 class="title">{{ t("settings.title") }}</h1>
      </div>

      <button
        class="refresh-btn"
        @click="refreshSettings"
        :title="t('common.refresh')"
      >
        <span class="refresh-icon">🔄</span>
      </button>
    </div>

    <!-- 设置内容区域 -->
    <div class="settings-content">
      <div class="settings-section">
        <h2 class="section-title">{{ t("settings.general") }}</h2>

        <div class="setting-item">
          <label class="setting-label">
            <input
              type="checkbox"
              class="setting-checkbox"
              v-model="settings.autoLaunch"
            />
            <span>{{ t("settings.autoLaunch") }}</span>
          </label>
          <p class="setting-description">{{ t("settings.autoLaunchDesc") }}</p>
        </div>

        <div class="setting-item">
          <label class="setting-label">
            <input
              type="checkbox"
              class="setting-checkbox"
              v-model="settings.keepAwake"
            />
            <span>{{ t("settings.keepAwake") }}</span>
          </label>
          <p class="setting-description">{{ t("settings.keepAwakeDesc") }}</p>
        </div>

        <div class="setting-item">
          <label class="setting-label">{{ t("settings.language") }}</label>
          <select
            class="setting-select"
            :value="currentLanguage"
            @change="
              changeLanguage(
                ($event.target as HTMLSelectElement).value as Locale,
              )
            "
          >
            <option value="zh-CN">{{ t("settings.chinese") }}</option>
            <option value="en-US">{{ t("settings.english") }}</option>
          </select>
        </div>

        <div class="setting-item">
          <label class="setting-label">{{ t("settings.theme") }}</label>
          <select class="setting-select" v-model="settings.theme">
            <option value="auto">{{ t("settings.followSystem") }}</option>
            <option value="light">{{ t("settings.lightMode") }}</option>
            <option value="dark">{{ t("settings.darkMode") }}</option>
          </select>
        </div>
      </div>

      <div class="settings-section">
        <h2 class="section-title">{{ t("settings.playback") }}</h2>

        <div class="setting-item">
          <label class="setting-label">{{
            t("settings.defaultPlayMode")
          }}</label>
          <select class="setting-select" v-model="settings.defaultPlayMode">
            <option value="loop">{{ t("playMode.loop", "循环播放") }}</option>
            <option value="order">{{ t("playMode.order", "顺序播放") }}</option>
            <option value="random">
              {{ t("playMode.random", "随机播放") }}
            </option>
            <option value="single">
              {{ t("playMode.single", "单曲循环") }}
            </option>
          </select>
        </div>

        <div class="setting-item">
          <label class="setting-label">
            <input
              type="checkbox"
              class="setting-checkbox"
              v-model="settings.showLyrics"
            />
            <span>{{ t("settings.showLyrics") }}</span>
          </label>
          <p class="setting-description">{{ t("settings.showLyricsDesc") }}</p>
        </div>

        <div class="setting-item">
          <label class="setting-label">{{ t("settings.volume") }}</label>
          <input
            type="range"
            class="setting-slider"
            min="0"
            max="100"
            v-model.number="settings.defaultVolume"
          />
          <span class="slider-value">{{ settings.defaultVolume }}%</span>
        </div>
      </div>

      <div class="settings-section">
        <h2 class="section-title">{{ t("settings.mediaKeys") }}</h2>

        <div class="setting-item">
          <label class="setting-label">
            <input
              type="checkbox"
              class="setting-checkbox"
              v-model="settings.enableMediaKeys"
            />
            <span>{{ t("settings.enableMediaKeys") }}</span>
          </label>
          <p class="setting-description">
            {{ t("settings.enableMediaKeysDesc") }}
          </p>
        </div>
      </div>

      <div class="settings-section">
        <h2 class="section-title">{{ t("settings.storage") }}</h2>

        <div class="setting-item">
          <label class="setting-label">{{
            t("settings.compactLibraries")
          }}</label>
          <p class="setting-description">
            {{ t("settings.compactLibrariesDesc") }}
          </p>
          <button
            class="action-btn"
            @click="compactLibraries"
            :disabled="isCompacting"
          >
            {{
              isCompacting
                ? t("settings.compacting")
                : t("settings.compactLibraries")
            }}
          </button>
          <p v-if="compactResult" class="result-message">{{ compactResult }}</p>
        </div>
      </div>

      <div class="settings-section">
        <h2 class="section-title">{{ t("settings.about") }}</h2>

        <div class="about-info">
          <p class="app-name">{{ t("settings.appName") }}</p>
          <p class="app-version">{{ t("settings.appVersion") }}</p>
          <p class="app-desc">{{ t("settings.appDesc") }}</p>
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

/* 重启提示 */
.restart-tip {
  position: fixed;
  top: 20px;
  right: 20px;
  background: rgba(79, 195, 247, 0.15);
  border: 1px solid rgba(79, 195, 247, 0.4);
  border-radius: 8px;
  padding: 12px 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(10px);
  z-index: 1000;
  animation: slideIn 0.3s ease-out;
  max-width: 400px;
}

@keyframes slideIn {
  from {
    transform: translateX(100%);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

.tip-content {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.tip-icon {
  font-size: 18px;
}

.tip-text {
  font-size: 13px;
  color: #ffffff;
  line-height: 1.4;
}

.tip-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.restart-btn {
  background: rgba(79, 195, 247, 0.3);
  border: 1px solid rgba(79, 195, 247, 0.5);
  color: #ffffff;
  padding: 6px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.restart-btn:hover {
  background: rgba(79, 195, 247, 0.5);
  border-color: rgba(79, 195, 247, 0.7);
  transform: translateY(-1px);
}

.restart-btn:active {
  transform: translateY(0);
}

.close-tip-btn {
  background: transparent;
  border: none;
  color: rgba(255, 255, 255, 0.6);
  font-size: 16px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: all 0.2s ease;
  line-height: 1;
}

.close-tip-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #ffffff;
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

/* 操作按钮样式 */
.action-btn {
  padding: 10px 20px;
  background: linear-gradient(135deg, #4fc3f7 0%, #29b6f6 100%);
  border: none;
  border-radius: 8px;
  color: white;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  margin-top: 12px;
}

.action-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(79, 195, 247, 0.4);
}

.action-btn:active:not(:disabled) {
  transform: translateY(0);
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.result-message {
  margin-top: 12px;
  padding: 10px 14px;
  background: rgba(76, 175, 80, 0.15);
  border: 1px solid rgba(76, 175, 80, 0.4);
  border-radius: 6px;
  color: #81c784;
  font-size: 13px;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
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
