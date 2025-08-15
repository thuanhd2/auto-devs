import ComingSoon from '@/components/coming-soon'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { soundService } from '@/services/soundService'
import { toast } from 'sonner'

export function NotificationsForm() {
  const [soundEnabled, setSoundEnabled] = useState(soundService.isEnabledValue())
  const [volume, setVolume] = useState(soundService.getVolume())

  useEffect(() => {
    setSoundEnabled(soundService.isEnabledValue())
    setVolume(soundService.getVolume())
  }, [])

  const handleSoundToggle = (enabled: boolean) => {
    setSoundEnabled(enabled)
    soundService.setEnabled(enabled)
    toast.success(enabled ? 'Sound notifications enabled' : 'Sound notifications disabled')
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
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Sound Notifications</CardTitle>
          <CardDescription>
            Configure sound notifications for task status changes
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center justify-between">
            <div className="space-y-1">
              <Label htmlFor="sound-enabled">Enable sound notifications</Label>
              <p className="text-sm text-muted-foreground">
                Play sounds when tasks move to Plan Review or Code Review status
              </p>
            </div>
            <Switch
              id="sound-enabled"
              checked={soundEnabled}
              onCheckedChange={handleSoundToggle}
            />
          </div>

          {soundEnabled && (
            <>
              <div className="space-y-3">
                <Label htmlFor="volume">Volume</Label>
                <div className="flex items-center space-x-3">
                  <input
                    id="volume"
                    type="range"
                    min="0"
                    max="1"
                    step="0.1"
                    value={volume}
                    onChange={(e) => handleVolumeChange(parseFloat(e.target.value))}
                    className="flex-1 h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer"
                  />
                  <span className="text-sm text-muted-foreground w-10">
                    {Math.round(volume * 100)}%
                  </span>
                </div>
              </div>

              <div className="space-y-3">
                <Label>Test Sounds</Label>
                <div className="flex gap-3">
                  <Button variant="outline" onClick={testPlanSound}>
                    Test Plan Complete Sound
                  </Button>
                  <Button variant="outline" onClick={testCodeSound}>
                    Test Code Complete Sound
                  </Button>
                </div>
                <p className="text-sm text-muted-foreground">
                  • Plan Complete: Plays when task moves to "Plan Review" status
                  <br />
                  • Code Complete: Plays when task moves to "Code Review" status
                </p>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
