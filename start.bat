@echo off
REM Haoyun Music Player - Windows 快速启动脚本

echo 🎵 Haoyun Music Player - 快速启动
echo ==================================

REM 检查 Go 是否安装
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo ❌ 错误：未找到 Go，请确保已安装 Go 1.25+
    pause
    exit /b 1
)

echo ✅ Go 已安装
go version

REM 检查 Node.js 是否安装
where node >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo ❌ 错误：未找到 Node.js，请确保已安装 Node.js 18+
    pause
    exit /b 1
)

echo ✅ Node.js 已安装
node --version

echo.
echo 📦 安装依赖...

REM 安装 Go 依赖
echo    → 安装 Go 依赖...
call go mod tidy

REM 安装前端依赖
echo    → 安装前端依赖...
cd frontend
call npm install
cd ..

echo.
echo ✅ 依赖安装完成!
echo.
echo 🚀 准备就绪!
echo.
echo    如果已安装 Wails v3，请运行:
echo    wails3 dev -config .\build\config.yml
echo.
echo    或者查看 QUICKSTART.md 了解更多选项
echo.
echo 🎵 Happy Coding!
pause
