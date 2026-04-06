<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import AppMain from "./components/AppMain.vue";
import BrowseView from "./views/BrowseView.vue";
import FavoritesView from "./views/FavoritesView.vue";
import SettingsView from "./views/SettingsView.vue";
import { Events } from "@wailsio/runtime";
// 当前视图
const currentView = ref<string>("main");

// Listen for events
const listenToBackendMessages = Events.On("windowUrl", (data: any) => {
  console.log("windowUrl:", data);
  console.log("📩 [测试消息] 收到后端测试消息:", data);
  console.log("   - 类型:", data.type);
  console.log("   - 消息:", data.message);
  console.log("   - 服务器时间:", data.serverTime);
  console.log("   - URL:", data.url);
  checkRoute();
});

// 检查 URL 路径 - 支持 hash 模式和 path 模式
const checkRoute = () => {
  const hash = window.location.hash;
  const pathname = window.location.pathname;
  const fullURL = window.location.href;

  console.log("[路由检查] fullURL:", fullURL);
  console.log("[路由检查] hash:", hash, "pathname:", pathname);

  // 检查 hash 路由（#/browse）
  if (hash === "#/browse" || hash.startsWith("#/browse/")) {
    console.log("[路由匹配] 匹配到 browse 视图");
    currentView.value = "browse";
    return;
  }

  // 检查 hash 路由（#/favorites）
  if (hash === "#/favorites" || hash.startsWith("#/favorites/")) {
    console.log("[路由匹配] 匹配到 favorites 视图");
    currentView.value = "favorites";
    return;
  }

  // 检查 hash 路由（#/settings）
  if (hash === "#/settings" || hash.startsWith("#/settings/")) {
    console.log("[路由匹配] 匹配到 settings 视图");
    currentView.value = "settings";
    return;
  }

  // 检查 path 路由（/browse）
  if (pathname === "/browse" || pathname.startsWith("/browse/")) {
    console.log("[路由匹配] 匹配到 browse 视图 (pathname)");
    currentView.value = "browse";
    return;
  }

  // 检查 path 路由（/favorites）
  if (pathname === "/favorites" || pathname.startsWith("/favorites/")) {
    console.log("[路由匹配] 匹配到 favorites 视图 (pathname)");
    currentView.value = "favorites";
    return;
  }

  // 检查 path 路由（/settings）
  if (pathname === "/settings" || pathname.startsWith("/settings/")) {
    console.log("[路由匹配] 匹配到 settings 视图 (pathname)");
    currentView.value = "settings";
    return;
  }

  // 默认显示主界面
  console.log("[路由匹配] 默认显示 main 视图");
  currentView.value = "main";
};

// 监听路由变化
const handleHashChange = () => {
  checkRoute();
};

// 生命周期
onMounted(() => {
  console.log(
    "[App.vue] 组件已挂载，初始 hash:",
    window.location.hash,
    "pathname:",
    window.location.pathname,
  );
  checkRoute();
  window.addEventListener("hashchange", handleHashChange);
  listenToBackendMessages();
});

onUnmounted(() => {
  window.removeEventListener("hashchange", handleHashChange);
  Events.Off("windowUrl");
});
</script>

<template>
  <component :is="currentView === 'browse' ? BrowseView : currentView === 'favorites' ? FavoritesView : currentView === 'settings' ? SettingsView : AppMain" />
</template>

<style scoped>
/* 全局容器样式 */
</style>
