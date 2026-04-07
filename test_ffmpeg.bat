@echo off
chcp 65001 >nul
echo ==========================================
echo   Haoyun Music Player - FFmpeg 测试工具
echo ==========================================
echo.

REM 检查 Go 是否安装
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ 错误: 未找到 Go，请先安装 Go 1.25+
    pause
    exit /b 1
)

echo ✅ Go 已安装
for /f "tokens=*" %%i in ('go version') do set GO_VERSION=%%i
echo    %GO_VERSION%
echo.

REM 检查 FFmpeg 是否安装
where ffmpeg >nul 2>&1
if %errorlevel% neq 0 (
    echo ⚠️  警告: 未找到 FFmpeg
    echo.
    echo 请安装 FFmpeg:
    echo   Chocolatey: choco install ffmpeg
    echo   Scoop: scoop install ffmpeg
    echo   手动下载: https://ffmpeg.org/download.html
    echo.
    set /p CONTINUE="是否继续运行测试？(y/n) "
    if /i not "%CONTINUE%"=="y" exit /b 1
) else (
    echo ✅ FFmpeg 已安装
    for /f "tokens=*" %%i in ('ffmpeg -version ^| findstr /C:"ffmpeg version"') do echo    %%i
    echo.
)

REM 进入 tests 目录
cd /d "%~dp0tests"

REM 编译测试程序
echo 🔨 编译测试程序...
go build -o test_ffmpeg_bin.exe test_ffmpeg.go

if %errorlevel% neq 0 (
    echo ❌ 编译失败
    pause
    exit /b 1
)

echo ✅ 编译成功
echo.

REM 运行测试
echo 🚀 运行测试...
echo.
test_ffmpeg_bin.exe

REM 清理
del test_ffmpeg_bin.exe

echo.
echo ==========================================
echo   测试完成
echo ==========================================
pause