# GitHub Pages 部署指南

本文档介绍如何将 Haoyun Music Player 官网部署到 GitHub Pages。

## 📋 前置要求

- GitHub 账号
- Git 已安装并配置
- 项目已推送到 GitHub 仓库

## 🚀 自动部署（推荐）

### 方法一：使用 GitHub Actions

项目已配置自动化部署工作流，只需以下步骤：

1. **启用 GitHub Pages**
   - 进入仓库的 **Settings** → **Pages**
   - 在 "Source" 下选择 **GitHub Actions**
   - 保存设置

2. **推送代码**
   ```bash
   git add .
   git commit -m "docs: 添加官网页面"
   git push origin main
   ```

3. **查看部署状态**
   - 进入 **Actions** 标签页
   - 查看 "Deploy to GitHub Pages" 工作流运行状态
   - 成功后会显示部署 URL

4. **访问网站**
   - 默认地址：`https://<username>.github.io/haoyun-music-player/`
   - 可在 **Settings** → **Pages** 中查看具体 URL

### 自定义域名（可选）

如果想使用自定义域名：

1. 在 `docs/` 文件夹中创建 `CNAME` 文件：
   ```
   music.yourdomain.com
   ```

2. 在域名 DNS 设置中添加 CNAME 记录指向 `<username>.github.io`

3. 在 **Settings** → **Pages** → **Custom domain** 中输入您的域名

## 🔧 手动部署

如果不想使用 GitHub Actions，可以手动部署：

### 方法二：使用 gh-pages 分支

```bash
# 1. 安装 gh-pages 工具
npm install -g gh-pages

# 2. 构建文档（将 markdown 转换为 HTML）
# 可以使用 marked、markdown-it 等工具

# 3. 部署到 gh-pages 分支
gh-pages -d docs
```

### 方法三：直接推送到 gh-pages 分支

```bash
# 1. 创建 gh-pages 分支
git checkout --orphan gh-pages

# 2. 清空当前内容
git rm -rf .

# 3. 复制 docs 文件夹内容到根目录
cp -r docs/* .

# 4. 提交并推送
git add .
git commit -m "Deploy to GitHub Pages"
git push origin gh-pages

# 5. 切换回主分支
git checkout main
```

然后在 **Settings** → **Pages** 中选择 `gh-pages` 分支作为源。

## 📁 文件结构

```
docs/
├── index.html          # 主页 HTML
├── styles.css          # 样式文件
├── script.js           # JavaScript 交互
├── favicon.png         # 网站图标（需添加）
├── preview.png         # 预览图片（需添加，用于社交媒体分享）
├── README.md           # 复制自根目录
├── WEBSITE.md          # 复制自根目录
├── FEATURES.md         # 复制自根目录
└── ...                 # 其他文档文件
```

## 🎨 自定义网站

### 修改主题颜色

编辑 `docs/styles.css` 中的 CSS 变量：

```css
:root {
    --primary-color: #4A90E2;      /* 主色调 */
    --secondary-color: #6C5CE7;    /* 次要色调 */
    --accent-color: #00D9FF;       /* 强调色 */
    --dark-bg: #1B2636;            /* 深色背景 */
}
```

### 添加真实截图

1. 截取应用界面图片
2. 保存到 `docs/images/` 文件夹
3. 在 `index.html` 中替换占位符：

```html
<div class="screenshot-placeholder">
    <img src="./images/main-player.png" alt="主播放器界面">
</div>
```

### 修改网站信息

编辑 `docs/index.html` 的 `<head>` 部分：

```html
<title>您的网站标题</title>
<meta name="description" content="网站描述">
<meta property="og:title" content="社交媒体标题">
```

## 🔍 SEO 优化

网站已包含以下 SEO 优化：

- ✅ Meta 描述标签
- ✅ Open Graph 协议（Facebook、LinkedIn）
- ✅ Twitter Card
- ✅ 语义化 HTML 结构
- ✅ 响应式设计
- ✅ 快速加载速度

进一步提升：

1. 添加 `sitemap.xml`
2. 创建 `robots.txt`
3. 提交到搜索引擎
4. 添加结构化数据（Schema.org）

## 📊 添加分析工具

### Google Analytics

在 `docs/index.html` 的 `</head>` 前添加：

```html
<!-- Google Analytics -->
<script async src="https://www.googletagmanager.com/gtag/js?id=GA_MEASUREMENT_ID"></script>
<script>
  window.dataLayer = window.dataLayer || [];
  function gtag(){dataLayer.push(arguments);}
  gtag('js', new Date());
  gtag('config', 'GA_MEASUREMENT_ID');
</script>
```

### Umami（开源替代）

```html
<script defer src="https://umami.example.com/script.js" data-website-id="your-website-id"></script>
```

## 🐛 故障排除

### 问题：页面显示 404

**解决方案**：
1. 确认 `docs/index.html` 文件存在
2. 检查 GitHub Pages 设置是否正确
3. 等待几分钟让 DNS 传播

### 问题：样式未加载

**解决方案**：
1. 检查浏览器控制台是否有 CORS 错误
2. 确认 CSS 文件路径正确
3. 清除浏览器缓存

### 问题：GitHub Actions 失败

**解决方案**：
1. 查看 Actions 日志了解具体错误
2. 确认文件权限正确
3. 检查工作流配置文件语法

### 问题：自定义域名不生效

**解决方案**：
1. 检查 DNS 记录是否正确
2. 确认 `CNAME` 文件格式正确
3. 等待 DNS 传播（最多 48 小时）

## 📝 更新网站

每次更新文档后：

```bash
# 1. 修改文档文件
# 2. 提交更改
git add .
git commit -m "docs: 更新网站内容"
git push origin main

# 3. GitHub Actions 会自动重新部署
```

## 🔗 相关链接

- [GitHub Pages 官方文档](https://docs.github.com/en/pages)
- [GitHub Actions 文档](https://docs.github.com/en/actions)
- [自定义域名指南](https://docs.github.com/en/pages/configuring-a-custom-domain-for-your-github-pages-site)

## 💡 最佳实践

1. **定期备份**：虽然 GitHub 有版本控制，但重要更改建议额外备份
2. **测试链接**：部署前检查所有内部链接是否正确
3. **性能优化**：压缩图片、最小化 CSS/JS
4. **移动端优先**：确保在移动设备上显示良好
5. **无障碍访问**：遵循 WCAG 标准，提高可访问性

---

**祝您部署成功！** 🎉

如有问题，请查看 [GitHub Pages 官方文档](https://docs.github.com/en/pages) 或提交 Issue。
