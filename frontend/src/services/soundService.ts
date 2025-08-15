class SoundService {
  private audioCache: Map<string, HTMLAudioElement> = new Map()
  private isEnabled: boolean = true
  private volume: number = 0.5

  constructor() {
    this.loadSettings()
    this.preloadSounds()
  }

  private loadSettings() {
    try {
      const settings = localStorage.getItem('soundSettings')
      if (settings) {
        const parsed = JSON.parse(settings)
        this.isEnabled = parsed.enabled ?? true
        this.volume = parsed.volume ?? 0.5
      }
    } catch (error) {
      console.warn('Failed to load sound settings:', error)
    }
  }

  private saveSettings() {
    try {
      localStorage.setItem('soundSettings', JSON.stringify({
        enabled: this.isEnabled,
        volume: this.volume
      }))
    } catch (error) {
      console.warn('Failed to save sound settings:', error)
    }
  }

  private preloadSounds() {
    const sounds = [
      { key: 'plan-completed', path: '/sounds/done-plan.wav' },
      { key: 'code-completed', path: '/sounds/done-code.wav' }
    ]

    sounds.forEach(({ key, path }) => {
      try {
        const audio = new Audio(path)
        audio.preload = 'auto'
        audio.volume = this.volume
        this.audioCache.set(key, audio)
      } catch (error) {
        console.warn(`Failed to preload sound: ${path}`, error)
      }
    })
  }

  private async playSound(soundKey: string): Promise<void> {
    if (!this.isEnabled) {
      return
    }

    const audio = this.audioCache.get(soundKey)
    if (!audio) {
      console.warn(`Sound not found: ${soundKey}`)
      return
    }

    try {
      // Reset audio to beginning in case it was already played
      audio.currentTime = 0
      audio.volume = this.volume
      
      // Play the sound
      await audio.play()
    } catch (error) {
      // Handle autoplay restrictions and other errors gracefully
      if (error instanceof DOMException && error.name === 'NotAllowedError') {
        console.warn('Sound autoplay blocked by browser. User interaction required.')
      } else {
        console.warn(`Failed to play sound: ${soundKey}`, error)
      }
    }
  }

  public async playPlanCompletedSound(): Promise<void> {
    return this.playSound('plan-completed')
  }

  public async playCodeCompletedSound(): Promise<void> {
    return this.playSound('code-completed')
  }

  public setEnabled(enabled: boolean): void {
    this.isEnabled = enabled
    this.saveSettings()
  }

  public isEnabledValue(): boolean {
    return this.isEnabled
  }

  public setVolume(volume: number): void {
    this.volume = Math.max(0, Math.min(1, volume)) // Clamp between 0 and 1
    
    // Update all cached audio elements
    this.audioCache.forEach(audio => {
      audio.volume = this.volume
    })
    
    this.saveSettings()
  }

  public getVolume(): number {
    return this.volume
  }

  public async testPlanSound(): Promise<void> {
    return this.playPlanCompletedSound()
  }

  public async testCodeSound(): Promise<void> {
    return this.playCodeCompletedSound()
  }
}

// Create and export a singleton instance
export const soundService = new SoundService()

// Export the class for potential testing
export { SoundService }