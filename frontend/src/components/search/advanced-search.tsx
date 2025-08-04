import * as React from 'react'
import { useState, useCallback, useRef, useEffect } from 'react'
import { Search, Clock, Bookmark, X, Filter, ChevronDown, ChevronUp } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Separator } from '@/components/ui/separator'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command'
import { cn } from '@/lib/utils'

export interface SearchFilter {
  id: string
  label: string
  value: string
  type: 'text' | 'select' | 'date' | 'status' | 'user'
  options?: { label: string; value: string }[]
}

export interface SavedSearch {
  id: string
  name: string
  query: string
  filters: SearchFilter[]
  timestamp: Date
  count?: number
}

export interface SearchSuggestion {
  id: string
  text: string
  type: 'query' | 'filter' | 'recent'
  category?: string
  metadata?: any
}

interface AdvancedSearchProps {
  placeholder?: string
  onSearch: (query: string, filters: SearchFilter[]) => void
  onSaveSearch?: (search: Omit<SavedSearch, 'id' | 'timestamp'>) => void
  availableFilters?: Omit<SearchFilter, 'value'>[]
  savedSearches?: SavedSearch[]
  recentSearches?: string[]
  suggestions?: SearchSuggestion[]
  className?: string
  showFilters?: boolean
  showHistory?: boolean
  showSavedSearches?: boolean
  debounceMs?: number
}

export function AdvancedSearch({
  placeholder = 'Search...',
  onSearch,
  onSaveSearch,
  availableFilters = [],
  savedSearches = [],
  recentSearches = [],
  suggestions = [],
  className,
  showFilters = true,
  showHistory = true,
  showSavedSearches = true,
  debounceMs = 300,
}: AdvancedSearchProps) {
  const [query, setQuery] = useState('')
  const [filters, setFilters] = useState<SearchFilter[]>([])
  const [isExpanded, setIsExpanded] = useState(false)
  const [showSuggestions, setShowSuggestions] = useState(false)
  const [filteredSuggestions, setFilteredSuggestions] = useState<SearchSuggestion[]>([])
  const [highlightedIndex, setHighlightedIndex] = useState(-1)
  
  const searchRef = useRef<HTMLInputElement>(null)
  const suggestionsRef = useRef<HTMLDivElement>(null)

  // Debounce search
  const debouncedSearch = useCallback(
    debounce((searchQuery: string, searchFilters: SearchFilter[]) => {
      onSearch(searchQuery, searchFilters)
    }, debounceMs),
    [onSearch, debounceMs]
  )

  // Filter suggestions based on query
  useEffect(() => {
    if (!query.trim()) {
      setFilteredSuggestions([])
      return
    }

    const queryLower = query.toLowerCase()
    const filtered = suggestions.filter(suggestion =>
      suggestion.text.toLowerCase().includes(queryLower)
    ).slice(0, 10)

    setFilteredSuggestions(filtered)
  }, [query, suggestions])

  // Handle search input change
  const handleQueryChange = (value: string) => {
    setQuery(value)
    setShowSuggestions(true)
    setHighlightedIndex(-1)
    debouncedSearch(value, filters)
  }

  // Handle filter changes
  const handleFilterChange = (filterId: string, value: string) => {
    const updatedFilters = filters.map(filter =>
      filter.id === filterId ? { ...filter, value } : filter
    )
    setFilters(updatedFilters)
    debouncedSearch(query, updatedFilters)
  }

  // Add filter
  const addFilter = (filterTemplate: Omit<SearchFilter, 'value'>) => {
    const newFilter: SearchFilter = {
      ...filterTemplate,
      value: '',
    }
    setFilters(prev => [...prev, newFilter])
  }

  // Remove filter
  const removeFilter = (filterId: string) => {
    const updatedFilters = filters.filter(filter => filter.id !== filterId)
    setFilters(updatedFilters)
    debouncedSearch(query, updatedFilters)
  }

  // Handle keyboard navigation
  const handleKeyDown = (event: React.KeyboardEvent) => {
    if (!showSuggestions || filteredSuggestions.length === 0) return

    switch (event.key) {
      case 'ArrowDown':
        event.preventDefault()
        setHighlightedIndex(prev => 
          prev < filteredSuggestions.length - 1 ? prev + 1 : 0
        )
        break
      case 'ArrowUp':
        event.preventDefault()
        setHighlightedIndex(prev => 
          prev > 0 ? prev - 1 : filteredSuggestions.length - 1
        )
        break
      case 'Enter':
        event.preventDefault()
        if (highlightedIndex >= 0) {
          selectSuggestion(filteredSuggestions[highlightedIndex])
        } else {
          setShowSuggestions(false)
        }
        break
      case 'Escape':
        setShowSuggestions(false)
        setHighlightedIndex(-1)
        break
    }
  }

  // Select suggestion
  const selectSuggestion = (suggestion: SearchSuggestion) => {
    setQuery(suggestion.text)
    setShowSuggestions(false)
    setHighlightedIndex(-1)
    debouncedSearch(suggestion.text, filters)
  }

  // Load saved search
  const loadSavedSearch = (savedSearch: SavedSearch) => {
    setQuery(savedSearch.query)
    setFilters(savedSearch.filters)
    debouncedSearch(savedSearch.query, savedSearch.filters)
  }

  // Clear all
  const clearAll = () => {
    setQuery('')
    setFilters([])
    setShowSuggestions(false)
    onSearch('', [])
  }

  // Save current search
  const saveCurrentSearch = () => {
    if (!onSaveSearch || (!query && filters.length === 0)) return
    
    const name = prompt('Enter name for saved search:')
    if (name) {
      onSaveSearch({
        name,
        query,
        filters: [...filters],
      })
    }
  }

  return (
    <div className={cn('relative w-full', className)}>
      {/* Main Search Input */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          ref={searchRef}
          type="text"
          placeholder={placeholder}
          value={query}
          onChange={(e) => handleQueryChange(e.target.value)}
          onKeyDown={handleKeyDown}
          onFocus={() => setShowSuggestions(true)}
          className="pl-10 pr-20"
        />
        
        {/* Search Actions */}
        <div className="absolute right-2 top-1/2 flex -translate-y-1/2 items-center gap-1">
          {(query || filters.length > 0) && (
            <Button
              variant="ghost"
              size="sm"
              onClick={clearAll}
              className="h-6 w-6 p-0"
            >
              <X className="h-3 w-3" />
            </Button>
          )}
          
          {showFilters && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setIsExpanded(!isExpanded)}
              className="h-6 w-6 p-0"
            >
              {isExpanded ? (
                <ChevronUp className="h-3 w-3" />
              ) : (
                <ChevronDown className="h-3 w-3" />
              )}
            </Button>
          )}
        </div>
      </div>

      {/* Search Suggestions */}
      {showSuggestions && filteredSuggestions.length > 0 && (
        <div
          ref={suggestionsRef}
          className="absolute top-full z-50 mt-1 w-full rounded-md border bg-popover p-0 shadow-md"
        >
          <ScrollArea className="max-h-48">
            {filteredSuggestions.map((suggestion, index) => (
              <div
                key={suggestion.id}
                className={cn(
                  'flex cursor-pointer items-center px-3 py-2 text-sm hover:bg-accent',
                  index === highlightedIndex && 'bg-accent'
                )}
                onClick={() => selectSuggestion(suggestion)}
              >
                <Search className="mr-2 h-3 w-3 text-muted-foreground" />
                <span>{suggestion.text}</span>
                {suggestion.type === 'recent' && (
                  <Clock className="ml-auto h-3 w-3 text-muted-foreground" />
                )}
              </div>
            ))}
          </ScrollArea>
        </div>
      )}

      {/* Expanded Filters Section */}
      {isExpanded && (
        <div className="mt-2 space-y-3 rounded-md border bg-card p-4">
          {/* Active Filters */}
          {filters.length > 0 && (
            <div className="space-y-2">
              <h4 className="text-sm font-medium">Active Filters</h4>
              <div className="flex flex-wrap gap-2">
                {filters.map((filter) => (
                  <FilterChip
                    key={filter.id}
                    filter={filter}
                    onValueChange={(value) => handleFilterChange(filter.id, value)}
                    onRemove={() => removeFilter(filter.id)}
                  />
                ))}
              </div>
            </div>
          )}

          {/* Add Filters */}
          {availableFilters.length > 0 && (
            <div className="space-y-2">
              <h4 className="text-sm font-medium">Add Filters</h4>
              <div className="flex flex-wrap gap-2">
                {availableFilters
                  .filter(template => !filters.some(f => f.id === template.id))
                  .map((template) => (
                    <Button
                      key={template.id}
                      variant="outline"
                      size="sm"
                      onClick={() => addFilter(template)}
                      className="h-7 text-xs"
                    >
                      <Filter className="mr-1 h-3 w-3" />
                      {template.label}
                    </Button>
                  ))}
              </div>
            </div>
          )}

          {/* Saved Searches */}
          {showSavedSearches && savedSearches.length > 0 && (
            <div className="space-y-2">
              <h4 className="text-sm font-medium">Saved Searches</h4>
              <div className="flex flex-wrap gap-2">
                {savedSearches.map((savedSearch) => (
                  <Button
                    key={savedSearch.id}
                    variant="outline"
                    size="sm"
                    onClick={() => loadSavedSearch(savedSearch)}
                    className="h-7 text-xs"
                  >
                    <Bookmark className="mr-1 h-3 w-3" />
                    {savedSearch.name}
                  </Button>
                ))}
              </div>
            </div>
          )}

          {/* Recent Searches */}
          {showHistory && recentSearches.length > 0 && (
            <div className="space-y-2">
              <h4 className="text-sm font-medium">Recent Searches</h4>
              <div className="flex flex-wrap gap-2">
                {recentSearches.slice(0, 5).map((recent, index) => (
                  <Button
                    key={index}
                    variant="ghost"
                    size="sm"
                    onClick={() => handleQueryChange(recent)}
                    className="h-7 text-xs"
                  >
                    <Clock className="mr-1 h-3 w-3" />
                    {recent}
                  </Button>
                ))}
              </div>
            </div>
          )}

          {/* Actions */}
          <div className="flex justify-between pt-2">
            <Button
              variant="outline"
              size="sm"
              onClick={clearAll}
              disabled={!query && filters.length === 0}
            >
              Clear All
            </Button>
            
            {onSaveSearch && (
              <Button
                variant="outline"
                size="sm"
                onClick={saveCurrentSearch}
                disabled={!query && filters.length === 0}
              >
                <Bookmark className="mr-1 h-3 w-3" />
                Save Search
              </Button>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

// Filter Chip Component
interface FilterChipProps {
  filter: SearchFilter
  onValueChange: (value: string) => void
  onRemove: () => void
}

function FilterChip({ filter, onValueChange, onRemove }: FilterChipProps) {
  const [isEditing, setIsEditing] = useState(false)

  return (
    <div className="flex items-center gap-1 rounded-full bg-secondary px-2 py-1 text-xs">
      <span className="font-medium">{filter.label}:</span>
      
      {isEditing ? (
        <Input
          type="text"
          value={filter.value}
          onChange={(e) => onValueChange(e.target.value)}
          onBlur={() => setIsEditing(false)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') setIsEditing(false)
          }}
          className="h-5 w-20 px-1 text-xs"
          autoFocus
        />
      ) : (
        <span
          className="cursor-pointer underline"
          onClick={() => setIsEditing(true)}
        >
          {filter.value || 'any'}
        </span>
      )}
      
      <Button
        variant="ghost"
        size="sm"
        onClick={onRemove}
        className="h-4 w-4 p-0 hover:bg-destructive hover:text-destructive-foreground"
      >
        <X className="h-2 w-2" />
      </Button>
    </div>
  )
}

// Utility function for debouncing
function debounce<T extends (...args: any[]) => any>(
  func: T,
  delay: number
): (...args: Parameters<T>) => void {
  let timeoutId: NodeJS.Timeout
  return (...args: Parameters<T>) => {
    clearTimeout(timeoutId)
    timeoutId = setTimeout(() => func(...args), delay)
  }
}