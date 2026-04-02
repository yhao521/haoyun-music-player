package backend

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/go-audio/wav"
	"github.com/mewkiz/flac"
	"github.com/tosone/minimp3"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// MP3Streamer 基于 minimp3 的流式读取器，实现 io.Reader 接口供 oto 使用
type MP3Streamer struct {
	pcmData    []byte     // 完全解码后的 PCM 数据
	pos        int        // 当前读取位置 (字节)
	sampleRate int        // 采样率
	channels   int        // 声道数
	mu         sync.Mutex // 并发保护
	closed     bool       // 是否已关闭
}

// NewMP3Streamer 创建 MP3 流式读取器
func NewMP3Streamer(file *os.File) (*MP3Streamer, error) {
	// 读取整个文件到内存
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败：%w", err)
	}

	// 使用 minimp3 的 DecodeFull 一次性解码所有数据
	decoder, pcmData, err := minimp3.DecodeFull(data)
	if err != nil {
		return nil, fmt.Errorf("MP3 解码失败：%w", err)
	}

	streamer := &MP3Streamer{
		pcmData:    pcmData,
		pos:        0,
		sampleRate: decoder.SampleRate,
		channels:   decoder.Channels,
		closed:     false,
	}

	return streamer, nil
}

// Read 实现 io.Reader 接口，供 oto 使用
func (m *MP3Streamer) Read(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed || m.pos >= len(m.pcmData) {
		return 0, io.EOF
	}

	n = copy(p, m.pcmData[m.pos:])
	m.pos += n
	
	if m.pos >= len(m.pcmData) {
		return n, io.EOF
	}
	
	return n, nil
}

// Len 返回总时长 (秒)
func (m *MP3Streamer) Len() int {
	if m.sampleRate == 0 || m.channels == 0 {
		return 0
	}
	bytesPerSecond := m.sampleRate * m.channels * 2 // 16-bit = 2 bytes per channel
	if bytesPerSecond == 0 {
		return 0
	}
	return len(m.pcmData) / bytesPerSecond
}

// Position 返回当前位置 (秒)
func (m *MP3Streamer) Position() int {
	if m.sampleRate == 0 || m.channels == 0 {
		return 0
	}
	bytesPerSecond := m.sampleRate * m.channels * 2
	if bytesPerSecond == 0 {
		return 0
	}
	return m.pos / bytesPerSecond
}

// Seek 跳转到指定位置 (秒)
func (m *MP3Streamer) Seek(position int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if position < 0 {
		position = 0
	}
	
	bytesPerSecond := m.sampleRate * m.channels * 2
	m.pos = position * bytesPerSecond
	
	if m.pos > len(m.pcmData) {
		m.pos = len(m.pcmData)
	}
	
	return nil
}

// Close 关闭流
func (m *MP3Streamer) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.closed {
		m.pcmData = nil
		m.closed = true
	}
	return nil
}

// PcmStreamer 通用 PCM 流式读取器 (用于 WAV, FLAC 等)
type PcmStreamer struct {
	pcmData    []byte     // PCM 数据
	pos        int        // 当前位置
	sampleRate int        // 采样率
	channels   int        // 声道数
	mu         sync.Mutex // 并发保护
	closed     bool       // 是否已关闭
}

// Read 实现 io.Reader 接口
func (p *PcmStreamer) Read(data []byte) (n int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed || p.pos >= len(p.pcmData) {
		return 0, io.EOF
	}

	n = copy(data, p.pcmData[p.pos:])
	p.pos += n
	
	if p.pos >= len(p.pcmData) {
		return n, io.EOF
	}
	
	return n, nil
}

// Len 返回总时长 (秒)
func (p *PcmStreamer) Len() int {
	if p.sampleRate == 0 || p.channels == 0 {
		return 0
	}
	bytesPerSecond := p.sampleRate * p.channels * 2
	if bytesPerSecond == 0 {
		return 0
	}
	return len(p.pcmData) / bytesPerSecond
}

// Position 返回当前位置 (秒)
func (p *PcmStreamer) Position() int {
	if p.sampleRate == 0 || p.channels == 0 {
		return 0
	}
	bytesPerSecond := p.sampleRate * p.channels * 2
	if bytesPerSecond == 0 {
		return 0
	}
	return p.pos / bytesPerSecond
}

// Seek 跳转到指定位置 (秒)
func (p *PcmStreamer) Seek(position int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if position < 0 {
		position = 0
	}
	
	bytesPerSecond := p.sampleRate * p.channels * 2
	p.pos = position * bytesPerSecond
	
	if p.pos > len(p.pcmData) {
		p.pos = len(p.pcmData)
	}
	
	return nil
}

// Close 关闭流
func (p *PcmStreamer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if !p.closed {
		p.pcmData = nil
		p.closed = true
	}
	return nil
}

// AudioPlayer 音频播放器 (使用 oto + 各种解码器)
type AudioPlayer struct {
	mu           sync.Mutex
	isPlaying    bool
	paused       bool
	volume       float64
	otoCtx       *oto.Context       // oto 上下文 (全局唯一，只创建一次)
	player       *oto.Player        // oto 播放器
	streamer     AudioReader        // 音频流式读取器
	stopChan     chan struct{}      // 停止信号通道
	app          *application.App   // Wails 应用引用
	ctxInitialized bool             // Context 是否已初始化
}

// NewAudioPlayer 创建音频播放器实例
func NewAudioPlayer() *AudioPlayer {
	return &AudioPlayer{
		volume: 0.7,
		ctxInitialized: false,
	}
}

// SetApp 设置应用实例
func (ap *AudioPlayer) SetApp(app *application.App) {
	ap.app = app
}

// AudioReader 音频读取器接口
type AudioReader interface {
	io.Reader
	io.Closer
	Len() int
	Position() int
	Seek(position int) error
}

// loadAudioFile 加载音频文件
func (ap *AudioPlayer) loadAudioFile(path string) (AudioReader, int, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("打开文件失败：%w", err)
	}

	ext := filepath.Ext(path)

	switch ext {
	case ".mp3":
		// 使用 minimp3 解码 MP3 文件 (主要方案)
		streamer, err := NewMP3Streamer(file)
		if err != nil {
			file.Close()
			return nil, 0, 0, err
		}
		return streamer, streamer.sampleRate, streamer.channels, nil

	case ".wav":
		// 使用 go-audio/wav 解码 WAV 文件
		decoder := wav.NewDecoder(file)

		// 读取所有 PCM 数据
		pcmBuffer, err := decoder.FullPCMBuffer()
		if err != nil {
			file.Close()
			return nil, 0, 0, fmt.Errorf("WAV 解码失败：%w", err)
		}

		// 将 int 数组转换为 byte 数组 (16-bit PCM)
		pcmData := make([]byte, len(pcmBuffer.Data)*2)
		for i, sample := range pcmBuffer.Data {
			binary.LittleEndian.PutUint16(pcmData[i*2:i*2+2], uint16(sample))
		}

		// 创建流式读取器
		streamer := &PcmStreamer{
			pcmData:    pcmData,
			sampleRate: int(decoder.SampleRate),
			channels:   int(decoder.NumChans),
		}
		return streamer, streamer.sampleRate, streamer.channels, nil

	case ".flac":
		// 使用 mewkiz/flac 解码 FLAC 文件
		stream, err := flac.New(file)
		if err != nil {
			file.Close()
			return nil, 0, 0, fmt.Errorf("FLAC 解码失败：%w", err)
		}

		// 获取音频信息
		sampleRate := int(stream.Info.SampleRate)
		channels := int(stream.Info.NChannels)

		// 读取所有 PCM 数据 (使用 frame 包)
		var pcmData []byte
		for {
			frame, err := stream.ParseNext()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, 0, 0, fmt.Errorf("FLAC 解析失败：%w", err)
			}
			
			// 将帧数据转换为字节
			for _, subFrame := range frame.Subframes {
				for _, sample := range subFrame.Samples {
					buf := make([]byte, 2)
					binary.LittleEndian.PutUint16(buf, uint16(sample))
					pcmData = append(pcmData, buf...)
				}
			}
		}
		stream.Close()

		streamer := &PcmStreamer{
			pcmData:    pcmData,
			sampleRate: sampleRate,
			channels:   channels,
		}
		return streamer, streamer.sampleRate, streamer.channels, nil

	default:
		file.Close()
		return nil, 0, 0, fmt.Errorf("不支持的音频格式：%s", ext)
	}
}

// closeOto 关闭 oto 播放器 (保留 Context)
func (ap *AudioPlayer) closeOto() {
	if ap.player != nil {
		ap.player.Close()
		ap.player = nil
	}
	// 注意：oto v3 的 Context 只能创建一次，所以我们保持 ap.otoCtx 和 ap.ctxInitialized 不变
	// 这样后续调用 initOto 时会直接复用已创建的 Context
}

// initOto 初始化或复用 oto 音频上下文
func (ap *AudioPlayer) initOto(sampleRate, channelCount int) error {
	// 如果已经创建过 Context，直接复用（oto v3 的 Context 只能创建一次）
	if ap.ctxInitialized {
		return nil
	}

	// 使用固定的音频参数创建 Context（oto v3 限制：只能创建一次）
	// 大多数音乐都是 44100Hz 立体声，所以我们使用这个标准参数
	const (
		targetSampleRate = 44100
		targetChannels = 2
	)

	// 创建新的 oto 上下文
	ctx, readyChan, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   targetSampleRate,
		ChannelCount: targetChannels,
		Format:       oto.FormatSignedInt16LE,
		BufferSize:   time.Second / 10,
	})
	if err != nil {
		return fmt.Errorf("初始化 oto 上下文失败：%w", err)
	}

	// 等待初始化完成
	<-readyChan

	// 检查是否有错误
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("oto 上下文错误：%w", err)
	}

	ap.otoCtx = ctx
	ap.ctxInitialized = true
	return nil
}

// Play 播放音频文件
func (ap *AudioPlayer) Play(path string) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	// 停止当前播放并清理资源
	ap.stopPlayback()

	// 加载音频文件
	reader, sampleRate, channels, err := ap.loadAudioFile(path)
	if err != nil {
		return err
	}

	// 初始化 oto Context（只在首次调用时创建）
	if err := ap.initOto(sampleRate, channels); err != nil {
		reader.Close()
		return err
	}

	// 创建 oto player
	ap.player = ap.otoCtx.NewPlayer(reader)
	
	// 保存流引用
	ap.streamer = reader

	// 启动播放
	ap.player.Play()

	// 启动监控协程
	ap.stopChan = make(chan struct{}, 1)
	go ap.monitorPlayback()

	ap.isPlaying = true
	ap.paused = false

	if ap.app != nil {
		ap.app.Event.Emit("playbackStateChanged", "playing")
	}

	return nil
}

// monitorPlayback 监控播放状态
func (ap *AudioPlayer) monitorPlayback() {
	for {
		select {
		case <-ap.stopChan:
			return
		default:
		}

		// 检查是否暂停
		ap.mu.Lock()
		if ap.paused {
			ap.mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			continue
		}
		ap.mu.Unlock()

		// 检查播放是否完成 (oto v3 使用 IsPlaying())
		if ap.player != nil && !ap.player.IsPlaying() {
			ap.mu.Lock()
			ap.isPlaying = false
			ap.mu.Unlock()

			if ap.app != nil {
				ap.app.Event.Emit("playbackStateChanged", "stopped")
			}
			return
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// stopPlayback 停止播放并清理资源
func (ap *AudioPlayer) stopPlayback() {
	if ap.stopChan != nil {
		select {
		case ap.stopChan <- struct{}{}:
		default:
		}
	}

	ap.closeOto()

	if ap.streamer != nil {
		ap.streamer.Close()
		ap.streamer = nil
	}

	ap.isPlaying = false
	ap.paused = false
}

// Pause 暂停播放
func (ap *AudioPlayer) Pause() error {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	if ap.streamer == nil {
		return fmt.Errorf("当前没有播放的音乐")
	}

	ap.paused = true
	ap.isPlaying = false

	if ap.app != nil {
		ap.app.Event.Emit("playbackStateChanged", "paused")
	}

	return nil
}

// Stop 停止播放
func (ap *AudioPlayer) Stop() error {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	ap.stopPlayback()

	if ap.app != nil {
		ap.app.Event.Emit("playbackStateChanged", "stopped")
	}

	return nil
}

// SetVolume 设置音量
func (ap *AudioPlayer) SetVolume(volume float64) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	if volume < 0 || volume > 1 {
		return fmt.Errorf("音量必须在 0 到 1 之间")
	}

	ap.volume = volume

	if ap.player != nil {
		ap.player.SetVolume(volume)
	}

	return nil
}

// GetVolume 获取音量
func (ap *AudioPlayer) GetVolume() (float64, error) {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	return ap.volume, nil
}

// Seek 跳转到指定位置 (秒)
func (ap *AudioPlayer) Seek(position float64) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	if ap.streamer == nil {
		return fmt.Errorf("当前没有播放的音乐")
	}

	// 调用 MP3Streamer 的 Seek 方法
	err := ap.streamer.Seek(int(position))
	if err != nil {
		return err
	}

	// 重置 oto player 以从新位置开始播放
	if ap.player != nil {
		ap.player.Reset()
		go func() {
			ap.player.Play()
		}()
	}

	return nil
}

// GetDuration 获取歌曲总时长 (秒)
func (ap *AudioPlayer) GetDuration() (float64, error) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	if ap.streamer == nil {
		return 0, fmt.Errorf("当前没有加载音乐")
	}

	// 调用 MP3Streamer 的 Len 方法
	return float64(ap.streamer.Len()), nil
}

// GetPosition 获取当前播放位置 (秒)
func (ap *AudioPlayer) GetPosition() (float64, error) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	if ap.streamer == nil {
		return 0, fmt.Errorf("当前没有播放的音乐")
	}

	// 调用 MP3Streamer 的 Position 方法
	return float64(ap.streamer.Position()), nil
}

// IsPlaying 检查是否正在播放
func (ap *AudioPlayer) IsPlaying() (bool, error) {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	return ap.isPlaying && !ap.paused, nil
}

// TogglePlayPause 切换播放/暂停
func (ap *AudioPlayer) TogglePlayPause() (bool, error) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	if ap.streamer == nil {
		return false, fmt.Errorf("当前没有播放的音乐")
	}

	ap.paused = !ap.paused
	ap.isPlaying = !ap.paused

	state := "playing"
	if ap.paused {
		state = "paused"
	}

	if ap.app != nil {
		ap.app.Event.Emit("playbackStateChanged", state)
	}

	return ap.isPlaying, nil
}
