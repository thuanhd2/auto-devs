class SoundService {
  private audioCache: Map<string, HTMLAudioElement> = new Map()
  private isEnabled: boolean = true
  private volume: number = 0.5
  private isInitialized: boolean = false

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
      localStorage.setItem(
        'soundSettings',
        JSON.stringify({
          enabled: this.isEnabled,
          volume: this.volume,
        })
      )
    } catch (error) {
      console.warn('Failed to save sound settings:', error)
    }
  }

  private async preloadSounds() {
    const sounds = [
      { key: 'plan-completed', path: '/sounds/done-plan.wav' },
      { key: 'code-completed', path: '/sounds/done-code.wav' },
    ]

    const loadPromises = sounds.map(async ({ key, path }) => {
      try {
        const audio = new Audio()

        // Set up event listeners for better error handling
        audio.addEventListener('error', (e) => {
          console.warn(`Failed to load sound: ${path}`, e)
          // Remove from cache if loading fails
          this.audioCache.delete(key)
        })

        audio.addEventListener('canplaythrough', () => {
          audio.volume = this.volume
          this.audioCache.set(key, audio)
          console.log(`Sound loaded successfully: ${key}`)
        })

        // Set source and start loading
        audio.src = path
        audio.preload = 'auto'

        // Try to load the audio with timeout
        try {
          const loadPromise = audio.load()
          const timeoutPromise = new Promise((_, reject) =>
            setTimeout(() => reject(new Error('Load timeout')), 10000)
          )

          await Promise.race([loadPromise, timeoutPromise])
        } catch (loadError) {
          console.warn(`Failed to load sound: ${path}`, loadError)
          // Remove from cache if loading fails
          this.audioCache.delete(key)
        }
      } catch (error) {
        console.warn(`Failed to create audio element for: ${path}`, error)
      }
    })

    try {
      await Promise.allSettled(loadPromises)
      this.isInitialized = true
      console.log(
        'Sound preloading completed. Loaded sounds:',
        Array.from(this.audioCache.keys())
      )
    } catch (error) {
      console.warn('Some sounds failed to preload:', error)
    }
  }

  private async playSound(soundKey: string): Promise<void> {
    if (!this.isEnabled) {
      return
    }

    if (!this.isInitialized) {
      console.warn('Sound service not yet initialized, retrying in 100ms...')
      await new Promise((resolve) => setTimeout(resolve, 100))
      if (!this.isInitialized) {
        console.warn(
          'Sound service still not initialized, skipping sound playback'
        )
        return
      }
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
        console.warn(
          'Sound autoplay blocked by browser. User interaction required.'
        )
      } else if (
        error instanceof DOMException &&
        error.name === 'NotSupportedError'
      ) {
        console.warn(
          `Sound format not supported or file not found: ${soundKey}`
        )
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
    this.audioCache.forEach((audio) => {
      audio.volume = this.volume
    })

    this.saveSettings()
  }

  public getVolume(): number {
    return this.volume
  }

  public getStatus(): {
    isEnabled: boolean
    volume: number
    isInitialized: boolean
    loadedSounds: string[]
  } {
    return {
      isEnabled: this.isEnabled,
      volume: this.volume,
      isInitialized: this.isInitialized,
      loadedSounds: Array.from(this.audioCache.keys()),
    }
  }

  public debugAudioElements(): void {
    console.log('=== Sound Service Debug ===')
    console.log('Enabled:', this.isEnabled)
    console.log('Volume:', this.volume)
    console.log('Initialized:', this.isInitialized)
    console.log('Audio Cache Size:', this.audioCache.size)

    this.audioCache.forEach((audio, key) => {
      console.log(`Sound: ${key}`)
      console.log('  - Ready State:', audio.readyState)
      console.log('  - Network State:', audio.networkState)
      console.log('  - Error:', audio.error)
      console.log('  - Src:', audio.src)
      console.log('  - Volume:', audio.volume)
    })
    console.log('========================')
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
