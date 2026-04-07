package backend

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

// ToolStatus 工具状态
type ToolStatus int

const (
	ToolNotInstalled ToolStatus = iota // 未安装
	ToolInstalling                      // 安装中
	ToolInstalled                       // 已安装
	ToolInstallFailed                   // 安装失败
)

// ToolInfo 工具信息
type ToolInfo struct {
	Name        string      // 工具名称
	Command     string      // 命令名
	Status      ToolStatus  // 状态
	Version     string      // 版本信息
	InstallHint string      // 安装提示
}

// DependencyManager 依赖管理器和安装器
type DependencyManager struct {
	mu       sync.RWMutex
	tools    map[string]*ToolInfo
	callback func(toolName string, status ToolStatus, message string) // 状态变化回调
}

// NewDependencyManager 创建依赖管理器
func NewDependencyManager() *DependencyManager {
	dm := &DependencyManager{
		tools: make(map[string]*ToolInfo),
	}
	
	// 初始化工具列表
	dm.initTools()
	
	return dm
}

// initTools 初始化需要检测的工具列表
func (dm *DependencyManager) initTools() {
	// FFmpeg - 音频解码
	dm.tools["ffmpeg"] = &ToolInfo{
		Name:    "FFmpeg",
		Command: "ffmpeg",
		Status:  ToolNotInstalled,
		InstallHint: dm.getFFmpegInstallHint(),
	}
	
	// ffprobe - 音频元数据提取（可选，随 FFmpeg 一起安装）
	dm.tools["ffprobe"] = &ToolInfo{
		Name:    "FFprobe",
		Command: "ffprobe",
		Status:  ToolNotInstalled,
		InstallHint: "随 FFmpeg 一起安装",
	}
}

// getFFmpegInstallHint 获取 FFmpeg 安装提示
func (dm *DependencyManager) getFFmpegInstallHint() string {
	switch runtime.GOOS {
	case "darwin":
		return "macOS: brew install ffmpeg"
	case "windows":
		return "Windows: choco install ffmpeg 或 scoop install ffmpeg"
	case "linux":
		return "Linux: sudo apt-get install ffmpeg 或 sudo dnf install ffmpeg"
	default:
		return "请从 https://ffmpeg.org/download.html 下载"
	}
}

// SetCallback 设置状态变化回调
func (dm *DependencyManager) SetCallback(callback func(toolName string, status ToolStatus, message string)) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.callback = callback
}

// emitCallback 安全地触发回调
func (dm *DependencyManager) emitCallback(toolName string, status ToolStatus, message string) {
	dm.mu.RLock()
	callback := dm.callback
	dm.mu.RUnlock()
	
	if callback != nil {
		callback(toolName, status, message)
	}
}

// CheckAllTools 检查所有工具的状态
func (dm *DependencyManager) CheckAllTools() map[string]*ToolInfo {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	for name, tool := range dm.tools {
		dm.checkToolUnsafe(name, tool)
	}
	
	return dm.tools
}

// checkToolUnsafe 检查单个工具状态（不加锁，内部使用）
func (dm *DependencyManager) checkToolUnsafe(name string, tool *ToolInfo) {
	// 查找可执行文件
	path, err := exec.LookPath(tool.Command)
	if err != nil {
		tool.Status = ToolNotInstalled
		tool.Version = ""
		log.Printf("⚠️  %s 未找到: %v", tool.Name, err)
		return
	}
	
	// 获取版本信息
	version := dm.getToolVersion(tool.Command)
	
	tool.Status = ToolInstalled
	tool.Version = version
	log.Printf("✅ %s 已安装: %s (版本: %s)", tool.Name, path, version)
}

// getToolVersion 获取工具版本信息
func (dm *DependencyManager) getToolVersion(command string) string {
	cmd := exec.Command(command, "-version")
	output, err := cmd.Output()
	if err != nil {
		return "未知"
	}
	
	// 提取第一行作为版本信息
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		// 只取第一行的前100个字符
		version := lines[0]
		if len(version) > 100 {
			version = version[:100] + "..."
		}
		return version
	}
	
	return "未知"
}

// GetToolStatus 获取指定工具的状态
func (dm *DependencyManager) GetToolStatus(toolName string) (*ToolInfo, bool) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	tool, exists := dm.tools[toolName]
	return tool, exists
}

// GetAllTools 获取所有工具状态
func (dm *DependencyManager) GetAllTools() map[string]*ToolInfo {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	result := make(map[string]*ToolInfo)
	for k, v := range dm.tools {
		result[k] = v
	}
	
	return result
}

// InstallTool 安装指定工具（后台异步）
func (dm *DependencyManager) InstallTool(toolName string) error {
	dm.mu.Lock()
	tool, exists := dm.tools[toolName]
	if !exists {
		dm.mu.Unlock()
		return fmt.Errorf("未知工具: %s", toolName)
	}
	
	// 检查是否已在安装
	if tool.Status == ToolInstalling {
		dm.mu.Unlock()
		return fmt.Errorf("%s 正在安装中", tool.Name)
	}
	
	// 标记为安装中
	tool.Status = ToolInstalling
	dm.mu.Unlock()
	
	dm.emitCallback(toolName, ToolInstalling, fmt.Sprintf("正在安装 %s...", tool.Name))
	log.Printf("🔧 开始安装 %s...", tool.Name)
	
	// 异步安装
	go func() {
		err := dm.performInstall(toolName, tool)
		
		dm.mu.Lock()
		if err != nil {
			tool.Status = ToolInstallFailed
			log.Printf("❌ %s 安装失败: %v", tool.Name, err)
			dm.emitCallback(toolName, ToolInstallFailed, fmt.Sprintf("安装失败: %v", err))
		} else {
			tool.Status = ToolInstalled
			tool.Version = dm.getToolVersion(tool.Command)
			log.Printf("✅ %s 安装成功", tool.Name)
			dm.emitCallback(toolName, ToolInstalled, "安装成功")
		}
		dm.mu.Unlock()
	}()
	
	return nil
}

// performInstall 执行实际安装
func (dm *DependencyManager) performInstall(toolName string, tool *ToolInfo) error {
	switch runtime.GOOS {
	case "darwin":
		return dm.installOnMacOS(toolName)
	case "windows":
		return dm.installOnWindows(toolName)
	case "linux":
		return dm.installOnLinux(toolName)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// installOnMacOS 在 macOS 上安装
func (dm *DependencyManager) installOnMacOS(toolName string) error {
	switch toolName {
	case "ffmpeg", "ffprobe":
		// 检查 Homebrew 是否安装
		if _, err := exec.LookPath("brew"); err != nil {
			return fmt.Errorf("Homebrew 未安装，请先安装 Homebrew:\n/bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\"")
		}
		
		// 使用 Homebrew 安装 FFmpeg
		cmd := exec.Command("brew", "install", "ffmpeg")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		log.Println("📦 执行: brew install ffmpeg")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Homebrew 安装失败: %w", err)
		}
		
		return nil
	default:
		return fmt.Errorf("macOS 上不支持自动安装此工具: %s", toolName)
	}
}

// installOnWindows 在 Windows 上安装
func (dm *DependencyManager) installOnWindows(toolName string) error {
	switch toolName {
	case "ffmpeg", "ffprobe":
		// 尝试使用 Chocolatey
		if _, err := exec.LookPath("choco"); err == nil {
			cmd := exec.Command("choco", "install", "ffmpeg", "-y")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			
			log.Println("📦 执行: choco install ffmpeg -y")
			if err := cmd.Run(); err == nil {
				return nil
			}
			log.Printf("Chocolatey 安装失败，尝试 Scoop...")
		}
		
		// 尝试使用 Scoop
		if _, err := exec.LookPath("scoop"); err == nil {
			cmd := exec.Command("scoop", "install", "ffmpeg")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			
			log.Println("📦 执行: scoop install ffmpeg")
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
		
		return fmt.Errorf("Windows 自动安装失败，请手动安装:\n1. 访问 https://ffmpeg.org/download.html\n2. 下载 Windows 版本\n3. 解压并添加到 PATH")
	default:
		return fmt.Errorf("Windows 上不支持自动安装此工具: %s", toolName)
	}
}

// installOnLinux 在 Linux 上安装
func (dm *DependencyManager) installOnLinux(toolName string) error {
	switch toolName {
	case "ffmpeg", "ffprobe":
		// 尝试 apt-get (Debian/Ubuntu)
		if _, err := exec.LookPath("apt-get"); err == nil {
			cmd := exec.Command("sudo", "apt-get", "update")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Printf("apt-get update 失败: %v", err)
			}
			
			cmd = exec.Command("sudo", "apt-get", "install", "-y", "ffmpeg")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			
			log.Println("📦 执行: sudo apt-get install -y ffmpeg")
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
		
		// 尝试 dnf (Fedora/RHEL)
		if _, err := exec.LookPath("dnf"); err == nil {
			cmd := exec.Command("sudo", "dnf", "install", "-y", "ffmpeg")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			
			log.Println("📦 执行: sudo dnf install -y ffmpeg")
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
		
		// 尝试 pacman (Arch Linux)
		if _, err := exec.LookPath("pacman"); err == nil {
			cmd := exec.Command("sudo", "pacman", "-S", "--noconfirm", "ffmpeg")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			
			log.Println("📦 执行: sudo pacman -S --noconfirm ffmpeg")
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
		
		return fmt.Errorf("Linux 自动安装失败，请手动安装:\nsudo apt-get install ffmpeg (Debian/Ubuntu)\nsudo dnf install ffmpeg (Fedora)\nsudo pacman -S ffmpeg (Arch)")
	default:
		return fmt.Errorf("Linux 上不支持自动安装此工具: %s", toolName)
	}
}

// GetInstallSummary 获取安装状态摘要
func (dm *DependencyManager) GetInstallSummary() string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	var summary strings.Builder
	summary.WriteString("=== 依赖工具状态 ===\n\n")
	
	allInstalled := true
	for _, tool := range dm.tools {
		statusIcon := "❌"
		switch tool.Status {
		case ToolInstalled:
			statusIcon = "✅"
		case ToolInstalling:
			statusIcon = "🔧"
			allInstalled = false
		case ToolInstallFailed:
			statusIcon = "⚠️"
			allInstalled = false
		case ToolNotInstalled:
			allInstalled = false
		}
		
		summary.WriteString(fmt.Sprintf("%s %s: %s\n", statusIcon, tool.Name, dm.getStatusText(tool.Status)))
		if tool.Version != "" && tool.Status == ToolInstalled {
			summary.WriteString(fmt.Sprintf("   版本: %s\n", tool.Version))
		}
		if tool.Status == ToolNotInstalled || tool.Status == ToolInstallFailed {
			summary.WriteString(fmt.Sprintf("   提示: %s\n", tool.InstallHint))
		}
		summary.WriteString("\n")
	}
	
	if allInstalled {
		summary.WriteString("✅ 所有依赖工具已就绪\n")
	} else {
		summary.WriteString("⚠️ 部分依赖工具缺失，建议安装以获得完整功能\n")
	}
	
	return summary.String()
}

// getStatusText 获取状态文本
func (dm *DependencyManager) getStatusText(status ToolStatus) string {
	switch status {
	case ToolNotInstalled:
		return "未安装"
	case ToolInstalling:
		return "安装中"
	case ToolInstalled:
		return "已安装"
	case ToolInstallFailed:
		return "安装失败"
	default:
		return "未知"
	}
}

// NeedInstall 检查是否需要安装任何工具
func (dm *DependencyManager) NeedInstall() bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	for _, tool := range dm.tools {
		if tool.Status == ToolNotInstalled || tool.Status == ToolInstallFailed {
			return true
		}
	}
	
	return false
}

// GetMissingTools 获取缺失的工具列表
func (dm *DependencyManager) GetMissingTools() []*ToolInfo {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	var missing []*ToolInfo
	for _, tool := range dm.tools {
		if tool.Status == ToolNotInstalled || tool.Status == ToolInstallFailed {
			missing = append(missing, tool)
		}
	}
	
	return missing
}