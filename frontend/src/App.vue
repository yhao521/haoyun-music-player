<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import AppMain from "./components/AppMain.vue";
import BrowseView from "./views/BrowseView.vue";

// 当前视图
const currentView = ref<string>("browse");

// 检查 URL 路径 - 支持 hash 模式和 path 模式
const checkRoute = () => {
  const hash = window.location.hash;
  const pathname = window.location.pathname;

  // 检查 hash 路由（#/browse）
  if (hash === "#/browse" || hash.startsWith("#/browse/")) {
    currentView.value = "browse";
    return;
  }

  // 检查 path 路由（/browse）
  if (pathname === "/browse" || pathname.startsWith("/browse/")) {
    currentView.value = "browse";
    return;
  }

  // 默认显示主界面
  currentView.value = "browse";
};

// 监听路由变化
const handleHashChange = () => {
  checkRoute();
};

// 生命周期
onMounted(() => {
  checkRoute();
  window.addEventListener("hashchange", handleHashChange);
});

onUnmounted(() => {
  window.removeEventListener("hashchange", handleHashChange);
});
</script>

<template>
  <component :is="currentView === 'browse' ? BrowseView : AppMain" />
</template>

<style scoped>
/* 全局容器样式 */
</style>
