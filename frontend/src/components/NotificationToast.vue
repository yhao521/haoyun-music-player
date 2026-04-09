<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import { Events } from "@wailsio/runtime";

interface Notification {
  id: number;
  title: string;
  message: string;
  type: "success" | "info" | "error";
}

const notifications = ref<Notification[]>([]);
let notificationId = 0;

// 显示通知
const showNotification = (title: string, message: string, type: "success" | "info" | "error" = "info") => {
  const id = ++notificationId;
  notifications.value.push({ id, title, message, type });
  
  // 3秒后自动移除
  setTimeout(() => {
    removeNotification(id);
  }, 3000);
};

// 移除通知
const removeNotification = (id: number) => {
  const index = notifications.value.findIndex(n => n.id === id);
  if (index > -1) {
    notifications.value.splice(index, 1);
  }
};

// 监听后端通知事件
let unsubscribe: (() => void) | null = null;

onMounted(() => {
  // 注册事件监听器
  Events.On("showNotification", (eventData: any) => {
    console.log("[Notification] 收到通知事件:", eventData);
    
    // Wails v3 事件数据结构: eventData.data 包含实际数据
    const data = eventData?.data || eventData;
    
    console.log("[Notification] 解析后的数据:", data);
    
    if (data && data.title && data.message) {
      showNotification(data.title, data.message, data.type || "info");
    } else {
      console.warn("[Notification] 无效的通知数据:", data);
    }
  });
});

onUnmounted(() => {
  // 清理事件监听器
  Events.Off("showNotification");
});
</script>

<template>
  <div class="notification-container">
    <TransitionGroup name="notification">
      <div
        v-for="notification in notifications"
        :key="notification.id"
        class="notification"
        :class="`notification-${notification.type}`"
      >
        <div class="notification-icon">
          <span v-if="notification.type === 'success'">✓</span>
          <span v-else-if="notification.type === 'error'">✕</span>
          <span v-else>ℹ</span>
        </div>
        <div class="notification-content">
          <div class="notification-title">{{ notification.title }}</div>
          <div class="notification-message">{{ notification.message }}</div>
        </div>
        <button class="notification-close" @click="removeNotification(notification.id)">×</button>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.notification-container {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-width: 400px;
}

.notification {
  display: flex;
  align-items: flex-start;
  padding: 12px 16px;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  background: white;
  min-width: 300px;
  animation: slideIn 0.3s ease-out;
}

.notification-success {
  border-left: 4px solid #52c41a;
  background: #f6ffed;
}

.notification-info {
  border-left: 4px solid #1890ff;
  background: #e6f7ff;
}

.notification-error {
  border-left: 4px solid #ff4d4f;
  background: #fff2f0;
}

.notification-icon {
  font-size: 20px;
  margin-right: 12px;
  flex-shrink: 0;
}

.notification-success .notification-icon {
  color: #52c41a;
}

.notification-info .notification-icon {
  color: #1890ff;
}

.notification-error .notification-icon {
  color: #ff4d4f;
}

.notification-content {
  flex: 1;
  min-width: 0;
}

.notification-title {
  font-weight: 600;
  font-size: 14px;
  margin-bottom: 4px;
  color: #262626;
}

.notification-message {
  font-size: 13px;
  color: #595959;
  line-height: 1.5;
  word-wrap: break-word;
}

.notification-close {
  background: none;
  border: none;
  font-size: 20px;
  color: #8c8c8c;
  cursor: pointer;
  padding: 0;
  margin-left: 8px;
  line-height: 1;
  transition: color 0.2s;
}

.notification-close:hover {
  color: #262626;
}

/* 动画效果 */
.notification-enter-active,
.notification-leave-active {
  transition: all 0.3s ease;
}

.notification-enter-from {
  opacity: 0;
  transform: translateX(100%);
}

.notification-leave-to {
  opacity: 0;
  transform: translateX(100%);
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateX(100%);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}
</style>
