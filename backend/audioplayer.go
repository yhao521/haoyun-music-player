package backend

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	mp3 "github.com/hajimehoshi/go-mp3"
	"github.com/ebitengine/oto/v3"
	"github.com/go-audio/wav"
	"github.com/mewkiz/flac"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// MP3Streamer 基于 go-mp3 的流式读取器，实现 io.Reader 接口供 oto 使用
type MP3Streamer struct {
	decoder    *mp3.Decoder // go-mp3 解码器
	sampleRate int          // 采样率
	channels   int          // 声道数
	bytesRead  int64        // 已读取的字节数（用于计算位置）
	mu         sync.Mutex   // 并发保护
	closed     bool         // 是否已关闭
}

// NewMP3Streamer 创建 MP3 流式读取器
func NewMP3Streamer(file *os.File) (*MP3Streamer, error) {
	// 使用 go-mp3 创建解码器（流式解码）
	decoder, err := mp3.NewDecoder(file)
	if err != nil {
		return nil, fmt.Errorf("MP3 解码器初始化失败：%w", err)
	}

	streamer := &MP3Streamer{
		decoder:    decoder,
		sampleRate: decoder.SampleRate(),
		channels:   2, // go-mp3 始终输出立体声
		bytesRead:  0,
		closed:     false,
	}

	return streamer, nil
}

// Read 实现 io.Reader 接口，供 oto 使用
func (m *MP3Streamer) Read(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return 0, io.EOF
	}

	// go-mp3 的 Decoder 直接实现 io.Reader
	n, err = m.decoder.Read(p)

	// 追踪已读取的字节数
	if n > 0 && err == nil {
		m.bytesRead += int64(n)
	}

	return n, err
}

// Len 返回总时长 (秒)
func (m *MP3Streamer) Len() int {
	if m.sampleRate == 0 || m.channels == 0 {
		return 0
	}

	// go-mp3 提供 Length() 方法返回采样数
	length := m.decoder.Length()
	bytesPerSecond := m.sampleRate * m.channels * 2 // 16-bit = 2 bytes per channel
	if bytesPerSecond == 0 {
		return 0
	}
	return int(length) / bytesPerSecond
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

	// 使用已读取的字节数计算当前位置
	return int(m.bytesRead) / bytesPerSecond
}

// Seek 跳转到指定位置 (秒)
func (m *MP3Streamer) Seek(position int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if position < 0 {
		position = 0
	}

	bytesPerSecond := m.sampleRate * m.channels * 2
	targetByte := int64(position * bytesPerSecond)

	// go-mp3 支持 Seek 方法
	actualPos, err := m.decoder.Seek(targetByte, io.SeekStart)
	if err != nil {
		return fmt.Errorf("MP3 Seek 失败：%w", err)
	}

	// 更新已读取字节数
	m.bytesRead = actualPos

	return nil
}

// Close 关闭流
func (m *MP3Streamer) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.closed {
		m.closed = true
		// go-mp3 的 Decoder 不需要显式关闭，底层文件由调用者管理
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
	mu             sync.Mutex
	isPlaying      bool
	paused         bool
	volume         float64
	currentPath    string           // 当前播放的文件路径
	pausePosition  int              // 暂停时的播放位置（秒）
	otoCtx         *oto.Context     // oto 上下文 (全局唯一，只创建一次)
	player         *oto.Player      // oto 播放器
	streamer       AudioReader      // 音频流式读取器
	stopChan       chan struct{}    // 停止信号通道
	app            *application.App // Wails 应用引用
	ctxInitialized bool             // Context 是否已初始化
}

// NewAudioPlayer 创建音频播放器实例
func NewAudioPlayer() *AudioPlayer {
	return &AudioPlayer{
		volume:         0.7,
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
		// 使用 go-mp3 解码 MP3 文件 (流式解码)
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
		targetChannels   = 2
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

	// 保存流引用和文件路径
	ap.streamer = reader
	ap.currentPath = path

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
				// 发出播放结束事件，由上层（MusicService）根据播放模式决定是否自动播放下一首
				ap.app.Event.Emit("playbackEnded", nil)
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

	log.Printf("[Pause] 开始暂停 - isPlaying: %v, paused: %v, player: %v, streamer: %v", 
		ap.isPlaying, ap.paused, ap.player != nil, ap.streamer != nil)

	if ap.streamer == nil {
		return fmt.Errorf("当前没有播放的音乐")
	}

	// Oto v3.4+ 正确暂停方式：
	// 1. 调用 player.Pause() 停止声音输出
	if ap.player != nil {
		ap.player.Pause()
		log.Println("[Pause] 已调用 player.Pause()")
	}

	// 2. 保存当前播放位置（用于断点续播）
	ap.pausePosition = ap.streamer.Position()
	log.Printf("[Pause] 已保存播放位置：%d 秒", ap.pausePosition)

	// 3. 关闭打开的 file 句柄（streamer）
	if ap.streamer != nil {
		ap.streamer.Close()
		ap.streamer = nil
		log.Println("[Pause] 已关闭 streamer")
	}

	// 4. 将 player 设为 nil，等待 GC 回收
	ap.player = nil

	// 5. 发送停止信号给监控协程
	if ap.stopChan != nil {
		select {
		case ap.stopChan <- struct{}{}:
			log.Println("[Pause] 已发送停止信号")
		default:
			log.Println("[Pause] 停止信号通道已满")
		}
	}

	// 6. 更新状态
	ap.paused = true
	ap.isPlaying = false

	log.Printf("[Pause] 暂停完成 - isPlaying: %v, paused: %v, position: %d", ap.isPlaying, ap.paused, ap.pausePosition)

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
	
	log.Printf("[TogglePlayPause] 开始 - isPlaying: %v, paused: %v, player: %v, streamer: %v, position: %d", 
		ap.isPlaying, ap.paused, ap.player != nil, ap.streamer != nil, ap.pausePosition)

	if ap.streamer == nil && ap.currentPath == "" {
		ap.mu.Unlock()
		return false, fmt.Errorf("当前没有播放的音乐")
	}

	// 如果当前是暂停状态，需要恢复播放
	if ap.paused {
		log.Println("[TogglePlayPause] 从暂停状态恢复")
		
		// 由于 streamer 已关闭，需要重新加载文件并播放
		if ap.currentPath != "" {
			// 保存要跳转的位置
			resumePosition := ap.pausePosition
			log.Printf("[TogglePlayPause] 准备从 %d 秒位置恢复播放", resumePosition)
			
			// ⭐ 关键：先释放锁，避免在 Play 中死锁
			ap.mu.Unlock()
			
			// 重新播放当前文件（会从头开始）
			log.Println("[TogglePlayPause] 调用 Play 方法...")
			err := ap.Play(ap.currentPath)
			
			if err != nil {
				log.Printf("[TogglePlayPause] 重新播放失败：%v", err)
				return false, err
			}
			
			log.Println("[TogglePlayPause] Play 方法调用完成")
			
			// 等待一小段时间确保播放已经开始
			time.Sleep(50 * time.Millisecond)
			
			// 重新加锁
			ap.mu.Lock()
			
			// 跳转到之前保存的播放位置（断点续播）
			if resumePosition > 0 {
				log.Printf("[TogglePlayPause] 跳转到播放位置：%d 秒", resumePosition)
				
				// ⭐ Seek 也需要释放锁以避免死锁
				ap.mu.Unlock()
				err = ap.Seek(float64(resumePosition))
				ap.mu.Lock()
				
				if err != nil {
					log.Printf("[TogglePlayPause] Seek 失败：%v", err)
					// Seek 失败不影响播放，继续返回成功
				} else {
					log.Printf("[TogglePlayPause] Seek 成功，当前位置：%d 秒", int(float64(resumePosition)))
				}
			}
			
			// 更新状态
			ap.paused = false
			ap.isPlaying = true
			
			log.Println("[TogglePlayPause] 重新播放完成")
			ap.mu.Unlock()
			
			if ap.app != nil {
				ap.app.Event.Emit("playbackStateChanged", "playing")
			}
			
			return true, nil
		} else {
			ap.mu.Unlock()
			log.Println("[TogglePlayPause] 没有保存的文件路径")
			return false, fmt.Errorf("无法恢复播放：文件路径丢失")
		}
	} else {
		log.Println("[TogglePlayPause] 执行暂停")
		
		// Oto v3.4+ 正确暂停方式：
		// 1. 调用 player.Pause() 停止声音输出
		if ap.player != nil {
			ap.player.Pause()
			log.Println("[TogglePlayPause] 已调用 player.Pause()")
		}

		// 2. 保存当前播放位置（用于断点续播）
		ap.pausePosition = ap.streamer.Position()
		log.Printf("[TogglePlayPause] 已保存播放位置：%d 秒", ap.pausePosition)

		// 3. 关闭打开的 file 句柄（streamer）
		if ap.streamer != nil {
			ap.streamer.Close()
			ap.streamer = nil
			log.Println("[TogglePlayPause] 已关闭 streamer")
		}

		// 4. 将 player 设为 nil，等待 GC 回收
		ap.player = nil

		// 5. 发送停止信号给监控协程
		if ap.stopChan != nil {
			select {
			case ap.stopChan <- struct{}{}:
				log.Println("[TogglePlayPause] 已发送停止信号")
			default:
				log.Println("[TogglePlayPause] 停止信号通道已满")
			}
		}

		// 6. 更新状态
		ap.paused = true
		ap.isPlaying = false

		log.Printf("[TogglePlayPause] 暂停完成 - isPlaying: %v, paused: %v, position: %d", ap.isPlaying, ap.paused, ap.pausePosition)

		ap.mu.Unlock()
		
		if ap.app != nil {
			ap.app.Event.Emit("playbackStateChanged", "paused")
		}

		return false, nil
	}
}
