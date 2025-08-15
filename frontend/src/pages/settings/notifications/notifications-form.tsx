import { useState, useEffect } from 'react'
import { soundService } from '@/services/soundService'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'

export function NotificationsForm() {
  const [soundEnabled, setSoundEnabled] = useState(
    soundService.isEnabledValue()
  )
  const [volume, setVolume] = useState(soundService.getVolume())
  const [soundStatus, setSoundStatus] = useState(soundService.getStatus())

  useEffect(() => {
    setSoundEnabled(soundService.isEnabledValue())
    setVolume(soundService.getVolume())
    setSoundStatus(soundService.getStatus())

    // Update status periodically to show initialization progress
    const interval = setInterval(() => {
      setSoundStatus(soundService.getStatus())
    }, 1000)

    return () => clearInterval(interval)
  }, [])

  const handleSoundToggle = (enabled: boolean) => {
    setSoundEnabled(enabled)
    soundService.setEnabled(enabled)
    toast.success(
      enabled ? 'Sound notifications enabled' : 'Sound notifications disabled'
    )
  }

  const handleVolumeChange = (newVolume: number) => {
    setVolume(newVolume)
    soundService.setVolume(newVolume)
  }

  const testPlanSound = async () => {
    try {
      await soundService.testPlanSound()
      toast.success('Plan completion sound played!')
    } catch (error) {
      toast.error('Failed to play sound. Check browser permissions.')
    }
  }

  const testCodeSound = async () => {
    try {
      await soundService.testCodeSound()
      toast.success('Code completion sound played!')
    } catch (error) {
      toast.error('Failed to play sound. Check browser permissions.')
    }
  }

  return (
    <div className='space-y-6'>
      <Card>
        <CardHeader>
          <CardTitle>Sound Notifications</CardTitle>
          <CardDescription>
            Configure sound notifications for task status changes
          </CardDescription>
        </CardHeader>
        <CardContent className='space-y-6'>
          <div className='flex items-center justify-between'>
            <div className='space-y-1'>
              <Label htmlFor='sound-enabled'>Enable sound notifications</Label>
              <p className='text-muted-foreground text-sm'>
                Play sounds when tasks move to Plan Review or Code Review status
              </p>
            </div>
            <Switch
              id='sound-enabled'
              checked={soundEnabled}
              onCheckedChange={handleSoundToggle}
            />
          </div>

          {soundEnabled && (
            <>
              <div className='space-y-3'>
                <Label htmlFor='volume'>Volume</Label>
                <div className='flex items-center space-x-3'>
                  <input
                    id='volume'
                    type='range'
                    min='0'
                    max='1'
                    step='0.1'
                    value={volume}
                    onChange={(e) =>
                      handleVolumeChange(parseFloat(e.target.value))
                    }
                    className='h-2 flex-1 cursor-pointer appearance-none rounded-lg bg-gray-200'
                  />
                  <span className='text-muted-foreground w-10 text-sm'>
                    {Math.round(volume * 100)}%
                  </span>
                </div>
              </div>

              <div className='space-y-3'>
                <Label>Test Sounds</Label>
                <div className='flex gap-3'>
                  <Button variant='outline' onClick={testPlanSound}>
                    Test Plan Complete Sound
                  </Button>
                  <Button variant='outline' onClick={testCodeSound}>
                    Test Code Complete Sound
                  </Button>
                  <Button
                    variant='outline'
                    onClick={() => soundService.debugAudioElements()}
                    className='text-xs'
                  >
                    Debug Console
                  </Button>
                </div>
                <p className='text-muted-foreground text-sm'>
                  • Plan Complete: Plays when task moves to "Plan Review" status
                  <br />• Code Complete: Plays when task moves to "Code Review"
                  status
                </p>
              </div>

              {/* Sound Service Status */}
              <div className='space-y-3'>
                <Label>Sound Service Status</Label>
                <div className='bg-muted space-y-2 rounded-lg p-3 text-sm'>
                  <div className='flex items-center gap-2'>
                    <span className='font-medium'>Status:</span>
                    <span
                      className={`rounded px-2 py-1 text-xs ${
                        soundStatus.isInitialized
                          ? 'bg-green-100 text-green-800'
                          : 'bg-yellow-100 text-yellow-800'
                      }`}
                    >
                      {soundStatus.isInitialized ? 'Ready' : 'Initializing...'}
                    </span>
                  </div>
                  <div className='flex items-center gap-2'>
                    <span className='font-medium'>Loaded Sounds:</span>
                    <span className='text-muted-foreground'>
                      {soundStatus.loadedSounds.length > 0
                        ? soundStatus.loadedSounds.join(', ')
                        : 'None'}
                    </span>
                  </div>
                  {!soundStatus.isInitialized && (
                    <p className='text-muted-foreground text-xs'>
                      Sound service is initializing. Please wait a moment before
                      testing sounds.
                    </p>
                  )}
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
