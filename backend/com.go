package backend

import (
	"bufio"
	"changeme/backend/pkg/file"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/options"
	runtime2 "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// App struct
type Com struct {
	ctx context.Context
	app *application.App
	// Base
}

// NewApp creates a new App application struct
// NewApp 创建一个新的 App 应用程序
func NewCom() *Com {
	return &Com{}
}
func (a *Com) SetApp(app *application.App) {
	a.app = app
}

func (a *Com) IsMacOS() bool {
	return runtime.GOOS == "darwin"
}

func (a *Com) SelectPathDownload() string {
	// if filetype == "" {
	// 	filetype = "*.txt;*.json"
	// }
	filePath := file.GetDownloadPath()
	path, err := a.app.Dialog.OpenFile().
		SetTitle("选择目录").
		SetDirectory(filePath).
		CanChooseDirectories(true).
		CanChooseFiles(false).
		PromptForSingleSelection()

	if err != nil || path == "" {
		return fmt.Sprintf("err %s!", err)
	}

	// selection, err := runtime2.OpenDirectoryDialog(a.ctx, runtime2.OpenDialogOptions{
	// 	Title:            "选择目录",
	// 	DefaultDirectory: filePath,
	// })
	// if err != nil {
	// 	return fmt.Sprintf("err %s!", err)
	// }
	return path
}

// ExtractAndDecodeBase64 extracts base64 content from data URL and decodes it to plain string
func (a *Com) ExtractAndDecodeBase64(dataURL string) (string, error) {
	// 正则表达式匹配 data:[mime-type];base64, 的格式
	re := regexp.MustCompile(`data:.+?;base64,`)
	base64String := re.ReplaceAllString(dataURL, "")

	// 解码base64字符串
	decodedBytes, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %v", err)
	}

	return string(decodedBytes), nil
}

// SaveFile 选择需要处理的文件
func (a *Com) SaveFile(data string, fileName string, filetype string) (resp string) {

	if filetype == "" {
		filetype = "*.txt;*.json;*.zip"
	}
	a.app.Logger.Info("SaveFile", "data", data)
	a.app.Logger.Info("SaveFile", "fileName", fileName)
	a.app.Logger.Info("SaveFile", "filetype", filetype)
	if (len(strings.Split(data, "base64,")) == 2) && strings.HasPrefix(data, "data:") {
		data, _ = a.ExtractAndDecodeBase64(data)
	}
	a.app.Logger.Info("SaveFile-ExtractAndDecodeBase64", "data", data)

	filePath := file.GetDownloadPath()
	path, err := a.app.Dialog.OpenFile().
		SetTitle("选择保存目录").
		AddFilter("Text Files", filetype).
		SetDirectory(filePath).
		CanChooseDirectories(true).
		CanChooseFiles(false).
		PromptForSingleSelection()

	if err != nil || path == "" {
		resp = fmt.Sprintf("err %s!", err)
		return
	}
	a.app.Logger.Info("SaveFile-path", "path", path)

	file, err := os.Create(path + "/" + fileName)
	if err != nil {
		resp = err.Error()
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, _ = writer.WriteString(data)

	writer.Flush()
	resp = path
	return
}

// var secondInstanceArgs []string
func (a *Com) OnSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	secondInstanceArgs := secondInstanceData.Args

	println("user opened second instance", strings.Join(secondInstanceData.Args, ","))
	println("user opened second from", secondInstanceData.WorkingDirectory)
	runtime2.WindowUnminimise(a.ctx)
	runtime2.Show(a.ctx)
	go runtime2.EventsEmit(a.ctx, "launchArgs", secondInstanceArgs)
}
