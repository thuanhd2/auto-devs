import { useState, useCallback, useEffect } from 'react'
import { Save, Eye, FileText, Download, Undo, Redo } from 'lucide-react'
import { toast } from 'sonner'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import { PlanPreview } from './plan-preview'

interface PlanEditorProps {
  initialValue: string
  onSave: (content: string) => Promise<void> | void
  onCancel: () => void
  isLoading?: boolean
  autoSave?: boolean
  autoSaveInterval?: number
}

interface HistoryState {
  content: string
  timestamp: number
  cursor?: number
}

export function PlanEditor({
  initialValue,
  onSave,
  onCancel,
  isLoading = false,
  autoSave = true,
  autoSaveInterval = 3000,
}: PlanEditorProps) {
  const [content, setContent] = useState(initialValue)
  const [activeTab, setActiveTab] = useState<'editor' | 'preview'>('editor')
  const [isDirty, setIsDirty] = useState(false)
  const [autoSaveEnabled, setAutoSaveEnabled] = useState(autoSave)
  const [lastAutoSave, setLastAutoSave] = useState<Date | null>(null)
  const [isSaving, setIsSaving] = useState(false)

  // History management for undo/redo
  const [history, setHistory] = useState<HistoryState[]>([
    { content: initialValue, timestamp: Date.now() },
  ])
  const [historyIndex, setHistoryIndex] = useState(0)

  // Update dirty state when content changes
  useEffect(() => {
    setIsDirty(content !== initialValue)
  }, [content, initialValue])

  // Auto-save functionality
  useEffect(() => {
    if (!autoSaveEnabled || !isDirty || isSaving) return

    const timeoutId = setTimeout(async () => {
      try {
        setIsSaving(true)
        await onSave(content)
        setLastAutoSave(new Date())
        setIsDirty(false)
        toast.success('Auto-saved', { duration: 1000 })
      } catch (error) {
        toast.error('Auto-save failed')
      } finally {
        setIsSaving(false)
      }
    }, autoSaveInterval)

    return () => clearTimeout(timeoutId)
  }, [content, isDirty, autoSaveEnabled, autoSaveInterval, onSave, isSaving])

  const handleContentChange = useCallback(
    (newContent: string) => {
      setContent(newContent)

      // Add to history if content is significantly different
      const lastHistoryItem = history[historyIndex]
      if (lastHistoryItem && newContent !== lastHistoryItem.content) {
        const newHistoryItem: HistoryState = {
          content: newContent,
          timestamp: Date.now(),
        }

        // Remove any history items after current index and add new item
        const newHistory = history.slice(0, historyIndex + 1)
        newHistory.push(newHistoryItem)

        // Keep history to reasonable size (max 50 items)
        if (newHistory.length > 50) {
          newHistory.shift()
        } else {
          setHistoryIndex(historyIndex + 1)
        }

        setHistory(newHistory)
      }
    },
    [history, historyIndex]
  )

  const handleUndo = useCallback(() => {
    if (historyIndex > 0) {
      const newIndex = historyIndex - 1
      setHistoryIndex(newIndex)
      setContent(history[newIndex].content)
    }
  }, [history, historyIndex])

  const handleRedo = useCallback(() => {
    if (historyIndex < history.length - 1) {
      const newIndex = historyIndex + 1
      setHistoryIndex(newIndex)
      setContent(history[newIndex].content)
    }
  }, [history, historyIndex])

  const handleManualSave = async () => {
    try {
      setIsSaving(true)
      await onSave(content)
      setIsDirty(false)
      toast.success('Plan saved successfully!')
    } catch (error) {
      toast.error('Failed to save plan')
    } finally {
      setIsSaving(false)
    }
  }

  const handleExportMarkdown = () => {
    const blob = new Blob([content], { type: 'text/markdown' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `plan-${new Date().toISOString().split('T')[0]}.md`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    toast.success('Plan exported as Markdown!')
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    // Ctrl+S or Cmd+S to save
    if ((e.ctrlKey || e.metaKey) && e.key === 's') {
      e.preventDefault()
      if (isDirty) {
        handleManualSave()
      }
    }
    // Ctrl+Z or Cmd+Z to undo
    if ((e.ctrlKey || e.metaKey) && e.key === 'z' && !e.shiftKey) {
      e.preventDefault()
      handleUndo()
    }
    // Ctrl+Y or Cmd+Shift+Z to redo
    if (
      ((e.ctrlKey || e.metaKey) && e.key === 'y') ||
      ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === 'z')
    ) {
      e.preventDefault()
      handleRedo()
    }
  }

  const canUndo = historyIndex > 0
  const canRedo = historyIndex < history.length - 1
  const showSaveButton = isDirty && !autoSaveEnabled

  return (
    <div className='flex h-full w-full flex-col' onKeyDown={handleKeyDown}>
      {/* Editor Header */}
      <CardHeader className='pb-3'>
        <div className='flex items-center justify-between'>
          <div className='flex items-center gap-3'>
            <CardTitle className='text-lg'>Edit Plan</CardTitle>
            {isDirty && (
              <Badge
                variant='outline'
                className='border-orange-200 text-orange-600'
              >
                Unsaved Changes
              </Badge>
            )}
            {isSaving && (
              <Badge
                variant='outline'
                className='border-blue-200 text-blue-600'
              >
                Saving...
              </Badge>
            )}
          </div>

          <div className='flex items-center gap-4'>
            {/* Auto-save toggle */}
            <div className='flex items-center space-x-2'>
              <Switch
                id='auto-save'
                checked={autoSaveEnabled}
                onCheckedChange={setAutoSaveEnabled}
                disabled={isLoading}
              />
              <Label htmlFor='auto-save' className='text-sm'>
                Auto-save
              </Label>
            </div>

            {/* History controls */}
            <div className='flex items-center gap-1'>
              <Button
                variant='ghost'
                size='sm'
                onClick={handleUndo}
                disabled={!canUndo || isLoading}
                title='Undo (Ctrl+Z)'
              >
                <Undo className='h-4 w-4' />
              </Button>
              <Button
                variant='ghost'
                size='sm'
                onClick={handleRedo}
                disabled={!canRedo || isLoading}
                title='Redo (Ctrl+Y)'
              >
                <Redo className='h-4 w-4' />
              </Button>
            </div>

            {/* Export */}
            <Button
              variant='ghost'
              size='sm'
              onClick={handleExportMarkdown}
              disabled={isLoading}
            >
              <Download className='mr-2 h-4 w-4' />
              Export
            </Button>
          </div>
        </div>

        {/* Status line */}
        <div className='flex items-center justify-between text-sm text-gray-500'>
          <div className='flex items-center gap-4'>
            <span>{content.length} characters</span>
            <span>{content.split('\n').length} lines</span>
            <span>{content.split(/\s+/).filter(Boolean).length} words</span>
          </div>
          {lastAutoSave && autoSaveEnabled && (
            <span>Auto-saved at {lastAutoSave.toLocaleTimeString()}</span>
          )}
        </div>
      </CardHeader>

      {/* Editor Tabs */}
      <CardContent className='flex-1 p-0'>
        <Tabs
          value={activeTab}
          onValueChange={(value: any) => setActiveTab(value)}
          className='flex h-full flex-col'
        >
          <TabsList className='mx-6 grid w-48 grid-cols-2'>
            <TabsTrigger value='editor' className='flex items-center gap-2'>
              <FileText className='h-4 w-4' />
              Editor
            </TabsTrigger>
            <TabsTrigger value='preview' className='flex items-center gap-2'>
              <Eye className='h-4 w-4' />
              Preview
            </TabsTrigger>
          </TabsList>

          <TabsContent value='editor' className='flex-1 px-6 pb-6'>
            <Textarea
              value={content}
              onChange={(e) => handleContentChange(e.target.value)}
              placeholder='Enter your implementation plan in Markdown format...'
              className='h-full min-h-[400px] w-full resize-none font-mono text-sm'
              disabled={isLoading}
            />
          </TabsContent>

          <TabsContent value='preview' className='flex-1 px-6 pb-6'>
            <div className='h-full overflow-auto rounded-md border'>
              <PlanPreview content={content} />
            </div>
          </TabsContent>
        </Tabs>
      </CardContent>

      {/* Editor Footer */}
      <CardFooter className='flex justify-between pt-4'>
        <div className='text-sm text-gray-500'>
          {autoSaveEnabled
            ? 'Changes are automatically saved'
            : 'Manual save mode - remember to save your changes'}
        </div>

        <div className='flex items-center gap-3'>
          <Button
            variant='outline'
            onClick={onCancel}
            disabled={isLoading || isSaving}
          >
            Cancel
          </Button>

          {showSaveButton && (
            <Button
              onClick={handleManualSave}
              disabled={isLoading || isSaving}
              className='flex items-center gap-2'
            >
              <Save className='h-4 w-4' />
              Save Changes
            </Button>
          )}
        </div>
      </CardFooter>
    </div>
  )
}
