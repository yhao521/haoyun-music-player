<template>
  <div class="donate-container">
    <div class="donate-content">
      <!-- 咖啡图标 -->
      <div class="coffee-icon">☕</div>

      <!-- 标题 -->
      <h1 class="donate-title">{{ t("donate.title") }}</h1>

      <!-- 副标题 -->
      <p class="donate-subtitle">{{ t("donate.subtitle") }}</p>

      <!-- 支付方式 -->
      <div class="payment-methods">
        <div class="payment-card alipay" @click="showAlipayQR">
          <div class="heart-icon">💙</div>
          <div class="payment-name">{{ t("donate.alipay") }}</div>
          <div class="payment-hint">{{ t("donate.scanToDonate") }}</div>
        </div>

        <div class="payment-card wechat" @click="showWechatQR">
          <div class="heart-icon">💚</div>
          <div class="payment-name">{{ t("donate.wechat") }}</div>
          <div class="payment-hint">{{ t("donate.scanToDonate") }}</div>
        </div>
      </div>

      <!-- 感谢语 -->
      <div class="thank-you-section">
        <h2 class="thank-you-title">🙏 {{ t("donate.thankYou") }}</h2>
        <p class="thank-you-desc">{{ t("donate.thankYouDesc") }}</p>
        <ul class="usage-list">
          <li>{{ t("donate.usage1") }}</li>
          <li>{{ t("donate.usage2") }}</li>
          <li>{{ t("donate.usage3") }}</li>
          <li>{{ t("donate.usage4") }}</li>
        </ul>
      </div>

      <!-- 页脚 -->
      <div class="footer">
        {{ t("donate.footer") }}
      </div>
    </div>

    <!-- 二维码弹窗 -->
    <div v-if="showQR" class="qr-modal" @click.self="closeQR">
      <div class="qr-content">
        <button class="close-btn" @click="closeQR">✕</button>
        <div class="qr-header">
          <h3>{{ qrTitle }}</h3>
        </div>
        <div class="qr-image-container">
          <img :src="qrImage" :alt="qrTitle" class="qr-image" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { t } from "../i18n";

const showQR = ref(false);
const qrTitle = ref("");
const qrImage = ref("");

// 显示支付宝二维码
const showAlipayQR = () => {
  qrTitle.value = t("donate.alipay");
  qrImage.value = "/alipay-qr.png";
  showQR.value = true;
};

// 显示微信二维码
const showWechatQR = () => {
  qrTitle.value = t("donate.wechat");
  qrImage.value = "/wechat-qr.png";
  showQR.value = true;
};

// 关闭二维码弹窗
const closeQR = () => {
  showQR.value = false;
};
</script>

<style scoped>
.donate-container {
  width: 100%;
  height: 100%;
  overflow-y: auto;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
  color: #e8e8e8;
  font-family:
    -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue",
    Arial, sans-serif;
}

.donate-content {
  max-width: 800px;
  margin: 0 auto;
  padding: 40px 20px;
  text-align: center;
}

.coffee-icon {
  font-size: 64px;
  margin-bottom: 20px;
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-10px);
  }
}

.donate-title {
  font-size: 32px;
  font-weight: 600;
  margin-bottom: 16px;
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.donate-subtitle {
  font-size: 16px;
  color: #a0a0a0;
  line-height: 1.6;
  margin-bottom: 40px;
  max-width: 600px;
  margin-left: auto;
  margin-right: auto;
}

.payment-methods {
  display: flex;
  gap: 30px;
  justify-content: center;
  margin-bottom: 50px;
  flex-wrap: wrap;
}

.payment-card {
  width: 280px;
  padding: 30px;
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.payment-card::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  opacity: 0.1;
  transition: opacity 0.3s ease;
}

.payment-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
}

.payment-card:hover::before {
  opacity: 0.2;
}

.payment-card.alipay {
  background: linear-gradient(135deg, #1677ff 0%, #4096ff 100%);
  box-shadow: 0 4px 15px rgba(22, 119, 255, 0.3);
}

.payment-card.alipay::before {
  background: #1677ff;
}

.payment-card.wechat {
  background: linear-gradient(135deg, #07c160 0%, #06ad56 100%);
  box-shadow: 0 4px 15px rgba(7, 193, 96, 0.3);
}

.payment-card.wechat::before {
  background: #07c160;
}

.heart-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.payment-name {
  font-size: 24px;
  font-weight: 600;
  color: #fff;
  margin-bottom: 8px;
}

.payment-hint {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.8);
}

.thank-you-section {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  padding: 30px;
  margin-bottom: 40px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.thank-you-title {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 16px;
  color: #ffd700;
}

.thank-you-desc {
  font-size: 15px;
  color: #a0a0a0;
  margin-bottom: 16px;
  line-height: 1.6;
}

.usage-list {
  list-style: none;
  padding: 0;
  margin: 0;
  text-align: left;
  display: inline-block;
}

.usage-list li {
  font-size: 14px;
  color: #c0c0c0;
  margin-bottom: 8px;
  padding-left: 20px;
  position: relative;
}

.usage-list li::before {
  content: "•";
  position: absolute;
  left: 0;
  color: #667eea;
  font-weight: bold;
}

.footer {
  font-size: 13px;
  color: #808080;
  padding-top: 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

/* 二维码弹窗 */
.qr-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(5px);
}

.qr-content {
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
  border-radius: 20px;
  padding: 40px;
  position: relative;
  max-width: 500px;
  width: 90%;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.close-btn {
  position: absolute;
  top: 15px;
  right: 15px;
  width: 32px;
  height: 32px;
  border: none;
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
  border-radius: 50%;
  cursor: pointer;
  font-size: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.2);
  transform: rotate(90deg);
}

.qr-header {
  margin-bottom: 30px;
}

.qr-header h3 {
  font-size: 24px;
  font-weight: 600;
  color: #fff;
  margin: 0;
}

.qr-image-container {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.qr-image {
  max-width: 100%;
  max-height: 400px;
  object-fit: contain;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .payment-methods {
    flex-direction: column;
    align-items: center;
  }

  .payment-card {
    width: 100%;
    max-width: 320px;
  }

  .donate-title {
    font-size: 24px;
  }

  .qr-content {
    padding: 30px 20px;
  }
}
</style>
