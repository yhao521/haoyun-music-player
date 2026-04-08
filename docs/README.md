# 📄 Docs 文件夹说明

此文件夹包含 Haoyun Music Player 的 **GitHub Pages 官网**文件。

## 📁 文件结构

```
docs/
├── index.html          # 官网主页（入口文件）
├── styles.css          # 网站样式表
├── script.js           # JavaScript 交互逻辑
├── DEPLOYMENT.md       # GitHub Pages 部署指南
└── *.md                # 自动复制的项目文档
```

## 🚀 快速开始

### 本地预览

您可以使用任何静态服务器在本地预览网站：

```bash
# 方法 1: 使用 Python
cd docs
python3 -m http.server 8080

# 方法 2: 使用 Node.js
npx serve docs

# 方法 3: 使用 PHP
php -S localhost:8080 -t docs
```

然后在浏览器中访问 `http://localhost:8080`

### 部署到 GitHub Pages

#### 自动部署（推荐）

项目已配置 GitHub Actions 自动部署：

1. 推送代码到 `main` 分支
2. GitHub Actions 会自动构建并部署
3. 访问 `https://<username>.github.io/haoyun-music-player/`

详细步骤请查看 [DEPLOYMENT.md](./DEPLOYMENT.md)

#### 手动部署

```bash
# 使用 gh-pages 工具
npm install -g gh-pages
gh-pages -d docs
```

## 🎨 自定义网站

### 修改内容

- **主页内容**：编辑 `index.html`
- **样式主题**：编辑 `styles.css` 中的 CSS 变量
- **交互功能**：编辑 `script.js`

### 添加截图

1. 将截图保存到 `docs/images/` 文件夹
2. 在 `index.html` 中替换占位符为真实图片

### 更换图标

- **Favicon**：替换 `favicon.png`（需创建此文件）
- **社交分享图**：创建 `preview.png`（1200x630 像素）

## 📊 网站特性

✅ **响应式设计** - 适配手机、平板、桌面  
✅ **现代化 UI** - 渐变、毛玻璃、动画效果  
✅ **SEO 优化** - Meta 标签、Open Graph、Twitter Card  
✅ **性能优化** - 轻量级、快速加载  
✅ **无障碍访问** - 语义化 HTML、键盘导航  
✅ **多语言支持** - 中英文档链接  

## 🔗 相关链接

- [官网首页](./index.html) - 产品介绍和下载链接
- [WEBSITE.md](../WEBSITE.md) - 完整产品页面
- [FEATURES.md](../FEATURES.md) - 功能特性详解
- [DEPLOYMENT.md](./DEPLOYMENT.md) - 部署指南

## 🛠️ 技术栈

- **HTML5** - 语义化结构
- **CSS3** - 现代样式和动画
- **Vanilla JavaScript** - 原生交互逻辑
- **Google Fonts** - Inter 字体

## 📝 更新日志

### v1.0.0 (2026-04-09)

- ✅ 创建完整的官网页面
- ✅ 实现响应式设计
- ✅ 添加 GitHub Actions 自动部署
- ✅ 集成所有项目文档
- ✅ SEO 优化和社交媒体支持

## 💡 提示

- 所有 `.md` 文档会在部署时自动从根目录复制到此文件夹
- 修改 `index.html` 后记得更新相关链接
- 建议使用 Chrome DevTools 测试不同设备的显示效果

## 🤝 贡献

欢迎改进官网设计！请：

1. Fork 本项目
2. 创建功能分支
3. 提交更改
4. 发起 Pull Request

---

**Made with ❤️ for Haoyun Music Player**
