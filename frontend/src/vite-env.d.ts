/// <reference types="vite/client" />

// Wails 运行时API 类型定义
declare interface Window {
  go: {
    main: {
      MusicService: {
        // 播放控制
        TogglePlayPause(): Promise<boolean>
        Play(): Promise<void>
        Pause(): Promise<void>
        Stop(): Promise<void>
        
        // 导航控制
        Next(): Promise<void>
        Previous(): Promise<void>
        PlayIndex(index: number): Promise<void>
        
        // 音量和进度
        SetVolume(volume: number): Promise<void>
        GetVolume(): Promise<number>
        Seek(position: number): Promise<void>
        GetDuration(): Promise<number>
        GetPosition(): Promise<number>
        
        // 状态查询
        IsPlaying(): Promise<boolean>
        
        // 文件管理
        LoadFile(path: string): Promise<void>
        AddToPlaylist(path: string): Promise<void>
        GetPlaylist(): Promise<string[]>
        OpenFilePicker(): Promise<string[]>
        
        // 元数据
        GetSongMetadata(path: string): Promise<{
          title: string
          artist: string
          album: string
          path: string
        }>
      }
    }
  }
  
  // Wails 运行时函数
  runtime: {
    EventsOn(event: string, callback: (...args: any[]) => void): void
    EventsOff(event: string): void
    EventsEmit(event: string, ...data: any[]): void
  }
}
