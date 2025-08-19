import { useState, useEffect } from 'react'
import { getAIs } from '@/types/task'
import { Bot, Play } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface ImplementationConfirmationDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  taskTitle: string
  onConfirm: (aiType: string) => void
}

export function ImplementationConfirmationDialog({
  open,
  onOpenChange,
  taskTitle,
  onConfirm,
}: ImplementationConfirmationDialogProps) {
  const [selectedAIType, setSelectedAIType] = useState<string>('')

  // Load AI type preference from localStorage
  useEffect(() => {
    const savedImplementingAI =
      localStorage.getItem('ai_preference_implementing') ||
      localStorage.getItem('ai_preference_planning') ||
      'claude-code'
    setSelectedAIType(savedImplementingAI)
  }, [])

  const handleConfirm = () => {
    if (selectedAIType) {
      // Save AI type preference to localStorage
      localStorage.setItem('ai_preference_implementing', selectedAIType)

      onConfirm(selectedAIType)
      onOpenChange(false)
    }
  }

  const handleCancel = () => {
    onOpenChange(false)
  }

  const ais = getAIs(false)

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-[425px]'>
        <DialogHeader>
          <DialogTitle className='flex items-center gap-2'>
            <Play className='h-5 w-5' />
            Approve Plan and Start Implementation
          </DialogTitle>
          <DialogDescription>
            Approve the plan and start implementing for task:{' '}
            <strong>{taskTitle}</strong>
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-4'>
          <div className='space-y-2'>
            <label className='flex items-center gap-2 text-sm font-medium'>
              <Bot className='h-4 w-4' />
              Select AI Assistant:
            </label>
            <Select value={selectedAIType} onValueChange={setSelectedAIType}>
              <SelectTrigger>
                <SelectValue placeholder='Select AI type' />
              </SelectTrigger>
              <SelectContent>
                {ais.map((ai) => (
                  <SelectItem key={ai.value} value={ai.value}>
                    <div className='flex items-center gap-2'>
                      <span>{ai.name}</span>
                      <span className='text-muted-foreground text-xs'>
                        ({ai.description})
                      </span>
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>

        <DialogFooter>
          <Button variant='outline' onClick={handleCancel}>
            Cancel
          </Button>
          <Button onClick={handleConfirm} disabled={!selectedAIType}>
            Approve and Start Implementation
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
