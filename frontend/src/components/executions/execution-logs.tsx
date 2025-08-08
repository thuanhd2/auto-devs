import { useState, useEffect, useRef } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { cn } from '@/lib/utils'
import type { ExecutionLog, LogLevel, ExecutionLogFilters } from '@/types/execution'
import { LOG_LEVEL_COLORS } from '@/types/execution'
import { 
  Search, 
  RefreshCw, 
  Download,
  Terminal,
  AlertTriangle,
  Info,
  Bug,
  ExternalLink,
  ChevronDown,
  Pause,
  Play,
} from 'lucide-react'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'


interface ExecutionLogsProps {
  logs: ExecutionLog[]
  loading?: boolean
  error?: string
  onRefresh?: () => void
  onLoadMore?: () => void
  hasMore?: boolean
  filters?: ExecutionLogFilters
  onFiltersChange?: (filters: ExecutionLogFilters) => void
  autoScroll?: boolean
  showFilters?: boolean
  showDownload?: boolean
  className?: string
}

const logLevelIcons = {
  debug: Bug,
  info: Info,
  warn: AlertTriangle,
  error: ExternalLink,
} as const

export function ExecutionLogs({
  logs,
  loading = false,
  error,
  onRefresh,
  onLoadMore,
  hasMore = false,
  filters,
  onFiltersChange,
  autoScroll = true,
  showFilters = true,
  showDownload = true,
  className,
}: ExecutionLogsProps) {
  const [searchTerm, setSearchTerm] = useState(filters?.search || '')
  const [selectedLevels, setSelectedLevels] = useState<LogLevel[]>(
    filters?.levels || ['debug', 'info', 'warn', 'error']
  )
  const [isPaused, setIsPaused] = useState(false)
  const scrollAreaRef = useRef<HTMLDivElement>(null)
  const bottomRef = useRef<HTMLDivElement>(null)

  // Auto-scroll to bottom when new logs arrive
  useEffect(() => {
    if (autoScroll && !isPaused && bottomRef.current) {
      bottomRef.current.scrollIntoView({ behavior: 'smooth' })
    }
  }, [logs, autoScroll, isPaused])

  // Filter logs based on search term and selected levels
  const filteredLogs = logs.filter(log => {
    const matchesSearch = !searchTerm || 
      log.message.toLowerCase().includes(searchTerm.toLowerCase()) ||
      log.source.toLowerCase().includes(searchTerm.toLowerCase())
    
    const matchesLevel = selectedLevels.includes(log.level)
    
    return matchesSearch && matchesLevel
  })

  const handleSearch = (value: string) => {
    setSearchTerm(value)
    onFiltersChange?.({ ...filters, search: value || undefined })
  }

  const handleLevelToggle = (level: LogLevel) => {
    const newLevels = selectedLevels.includes(level)
      ? selectedLevels.filter(l => l !== level)
      : [...selectedLevels, level]
    
    setSelectedLevels(newLevels)
    onFiltersChange?.({ ...filters, levels: newLevels })
  }

  const handleDownload = () => {
    const logContent = filteredLogs
      .map(log => `[${log.timestamp}] ${log.level.toUpperCase()} ${log.source}: ${log.message}`)
      .join('\n')
    
    const blob = new Blob([logContent], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `execution-logs-${new Date().toISOString().slice(0, 10)}.txt`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  const logStats = {
    debug: logs.filter(l => l.level === 'debug').length,
    info: logs.filter(l => l.level === 'info').length,
    warn: logs.filter(l => l.level === 'warn').length,
    error: logs.filter(l => l.level === 'error').length,
  }

  return (
    <div className={cn('space-y-4', className)}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="space-y-1">
          <h3 className="text-lg font-semibold flex items-center gap-2">
            <Terminal className="h-5 w-5" />
            Execution Logs ({filteredLogs.length})
          </h3>
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <span>Showing {filteredLogs.length} of {logs.length} logs</span>
            {loading && (
              <Badge variant="secondary" className="gap-1">
                <RefreshCw className="h-3 w-3 animate-spin" />
                Loading...
              </Badge>
            )}
          </div>
        </div>
        
        <div className="flex items-center gap-2">
          <Button
            variant={isPaused ? "default" : "outline"}
            size="sm"
            onClick={() => setIsPaused(!isPaused)}
            className="gap-1"
          >
            {isPaused ? (
              <>
                <Play className="h-4 w-4" />
                Resume
              </>
            ) : (
              <>
                <Pause className="h-4 w-4" />
                Pause
              </>
            )}
          </Button>
          
          {onRefresh && (
            <Button
              variant="outline"
              size="sm"
              onClick={onRefresh}
              disabled={loading}
              className="gap-1"
            >
              <RefreshCw className={cn('h-4 w-4', loading && 'animate-spin')} />
              Refresh
            </Button>
          )}
          
          {showDownload && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleDownload}
              className="gap-1"
            >
              <Download className="h-4 w-4" />
              Download
            </Button>
          )}
        </div>
      </div>

      {/* Log Level Stats */}
      <div className="flex flex-wrap gap-2">
        {Object.entries(logStats).map(([level, count]) => {
          const Icon = logLevelIcons[level as LogLevel]
          const isSelected = selectedLevels.includes(level as LogLevel)
          
          return (
            <Button
              key={level}
              variant={isSelected ? "secondary" : "ghost"}
              size="sm"
              onClick={() => handleLevelToggle(level as LogLevel)}
              className={cn(
                'gap-1 h-auto py-1.5 px-2 text-xs font-medium',
                isSelected && 'ring-2 ring-primary'
              )}
            >
              <Icon className="h-3 w-3" />
              <span className="capitalize">{level}</span>
              <Badge variant="secondary" className="ml-1 h-4 text-xs px-1">
                {count}
              </Badge>
            </Button>
          )
        })}
      </div>

      {/* Filters */}
      {showFilters && (
        <div className="flex items-center gap-2">
          <div className="flex-1 max-w-xs">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search logs..."
                value={searchTerm}
                onChange={(e) => handleSearch(e.target.value)}
                className="pl-10"
              />
            </div>
          </div>
          
          <Select 
            defaultValue="timestamp-desc"
            onValueChange={(value) => {
              const [orderBy, orderDir] = value.split('-')
              onFiltersChange?.({ 
                ...filters, 
                order_by: orderBy as 'timestamp' | 'level' | 'source', 
                order_dir: orderDir as 'asc' | 'desc' 
              })
            }}
          >
            <SelectTrigger className="w-40">
              <SelectValue placeholder="Sort by" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="timestamp-desc">Latest first</SelectItem>
              <SelectItem value="timestamp-asc">Oldest first</SelectItem>
              <SelectItem value="level-asc">Level ↑</SelectItem>
              <SelectItem value="level-desc">Level ↓</SelectItem>
              <SelectItem value="source-asc">Source A-Z</SelectItem>
            </SelectContent>
          </Select>
        </div>
      )}

      {/* Error State */}
      {error && (
        <Card className="border-red-200 bg-red-50">
          <CardContent className="p-4">
            <div className="flex items-center gap-2 text-red-800">
              <AlertTriangle className="h-4 w-4" />
              <span className="font-medium">Error loading logs</span>
            </div>
            <p className="text-sm text-red-700 mt-1">{error}</p>
          </CardContent>
        </Card>
      )}

      {/* Log Display */}
      <Card className="h-96">
        <CardContent className="p-0">
          <ScrollArea 
            ref={scrollAreaRef}
            className="h-96 w-full"
          >
            <div className="p-4 space-y-1 font-mono text-sm">
              {filteredLogs.length === 0 ? (
                <div className="text-center text-muted-foreground py-8">
                  {loading ? 'Loading logs...' : 'No logs found'}
                </div>
              ) : (
                filteredLogs.map((log) => (
                  <LogEntry key={log.id} log={log} />
                ))
              )}
              
              {hasMore && onLoadMore && (
                <div className="text-center py-4">
                  <Button 
                    variant="ghost" 
                    onClick={onLoadMore}
                    disabled={loading}
                  >
                    {loading ? (
                      <>
                        <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                        Loading...
                      </>
                    ) : (
                      <>
                        <ChevronDown className="mr-2 h-4 w-4" />
                        Load More
                      </>
                    )}
                  </Button>
                </div>
              )}
              
              <div ref={bottomRef} />
            </div>
          </ScrollArea>
        </CardContent>
      </Card>
    </div>
  )
}

function LogEntry({ log }: { log: ExecutionLog }) {
  const Icon = logLevelIcons[log.level]
  const colorClasses = LOG_LEVEL_COLORS[log.level]
  
  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleTimeString()
  }

  return (
    <div className="flex items-start gap-2 py-1 px-2 hover:bg-muted/50 rounded text-xs group">
      {/* Timestamp */}
      <span className="text-muted-foreground whitespace-nowrap">
        {formatTimestamp(log.timestamp)}
      </span>
      
      {/* Level Badge */}
      <Badge 
        variant="secondary" 
        className={cn('gap-1 h-5 text-xs px-1.5 whitespace-nowrap', colorClasses)}
      >
        <Icon className="h-2.5 w-2.5" />
        {log.level.toUpperCase()}
      </Badge>
      
      {/* Source */}
      <span className="text-muted-foreground whitespace-nowrap min-w-0 max-w-20 truncate">
        {log.source}
      </span>
      
      {/* Message */}
      <div className="flex-1 min-w-0">
        <span className={cn(
          'break-words',
          log.level === 'error' && 'text-red-700',
          log.level === 'warn' && 'text-yellow-700'
        )}>
          {log.message}
        </span>
        
        {/* Metadata */}
        {log.metadata && (
          <div className="text-muted-foreground text-xs mt-1 opacity-0 group-hover:opacity-100 transition-opacity">
            {JSON.stringify(log.metadata)}
          </div>
        )}
      </div>
    </div>
  )
}