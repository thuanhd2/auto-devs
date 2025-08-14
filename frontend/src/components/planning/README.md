# Plan Review Interface Components

This directory contains React components for reviewing, editing, and managing AI-generated implementation plans.

## Components Overview

### 1. PlanReview

The main component that provides a complete plan review interface with approval/rejection functionality.

**Features:**

- Plan display with markdown rendering
- Approval/rejection buttons for `PLAN_REVIEWING` status
- Integrated editing capabilities
- Export functionality
- Responsive design

**Usage:**

```tsx
import { PlanReview } from '@/components/planning'

;<PlanReview
  task={task}
  onPlanUpdate={(updatedTask) => console.log('Plan updated:', updatedTask)}
  onStatusChange={(taskId, newStatus) =>
    console.log('Status changed:', newStatus)
  }
/>
```

### 2. PlanEditor

A feature-rich markdown editor for plan content with auto-save, history, and export features.

**Features:**

- Live preview with tabs (editor/preview)
- Auto-save functionality with toggle
- Undo/redo history (up to 50 actions)
- Export to markdown file
- Keyboard shortcuts (Ctrl+S, Ctrl+Z, Ctrl+Y)
- Character/word/line count
- Validation and error handling

**Usage:**

```tsx
import { PlanEditor } from '@/components/planning'

;<PlanEditor
  initialValue={task.plan}
  onSave={async (content) => {
    // Save logic here
    await updatePlan(content)
  }}
  onCancel={() => setEditMode(false)}
  autoSave={true}
  autoSaveInterval={3000}
/>
```

### 3. PlanPreview

A markdown preview component with syntax highlighting and export options.

**Features:**

- Markdown parsing and HTML rendering
- Copy to clipboard
- Print functionality
- Full-screen view in new tab
- Print-friendly styling
- Responsive design

**Usage:**

```tsx
import { PlanPreview } from '@/components/planning'

;<PlanPreview content={planContent} showActions={true} printFriendly={false} />
```

## Task Status Integration

The components work with the following task statuses:

- `PLANNING` - AI is generating the plan
- `PLAN_REVIEWING` - Plan is ready for human review
- `IMPLEMENTING` - Plan approved, ready for implementation

## Responsive Design

All components are designed to work across different screen sizes:

### Mobile (≤768px)

- Single column layout
- Stacked buttons
- Simplified toolbar
- Touch-friendly interactions

### Tablet (769px-1024px)

- Two-column layout where appropriate
- Condensed toolbar
- Optimized spacing

### Desktop (>1024px)

- Full feature set
- Multi-column layout
- Extended toolbar options
- Keyboard shortcuts

## API Integration

The components integrate with the existing task management API:

### Task Update

```typescript
// Update task plan
const response = await tasksApi.updateTask(taskId, {
  plan: updatedPlanContent
})

// Update task status
const response = await tasksApi.updateTask(taskId, {
  status: 'IMPLEMENTING' // or 'PLANNING'
})
```

## Styling and Theming

Components use the existing design system:

- ShadcnUI components for consistent styling
- Tailwind CSS for responsive design
- Lucide icons for UI elements
- Consistent color scheme with the rest of the app

## Keyboard Shortcuts

### Editor Component

- `Ctrl/Cmd + S` - Save changes
- `Ctrl/Cmd + Z` - Undo
- `Ctrl/Cmd + Y` or `Ctrl/Cmd + Shift + Z` - Redo

### Preview Component

- `Ctrl/Cmd + P` - Print (when focused)
- `Ctrl/Cmd + C` - Copy content (when focused)

## Performance Considerations

- **Auto-save debouncing**: 3-second delay to prevent excessive API calls
- **History management**: Limited to 50 undo/redo steps
- **Markdown parsing**: Client-side rendering for fast preview updates
- **Large content handling**: Graceful degradation for very long plans

## Testing

The components include a demo page for testing different scenarios:

```tsx
import { PlanReviewDemo } from '@/components/planning'

// Renders interactive demo with sample data

;<PlanReviewDemo />
```

Test scenarios include:

- Tasks with complex plans
- Tasks without plans
- Tasks with short plans
- Different task statuses
- Mobile/tablet responsive views

## Error Handling

Components handle common error scenarios:

- Network failures during save operations
- Invalid markdown content
- Missing task data
- API timeout errors

Errors are displayed using toast notifications and inline error states.

## Accessibility

- Semantic HTML structure
- ARIA labels for screen readers
- Keyboard navigation support
- Focus management
- High contrast color support

## Integration with Existing Components

The planning components integrate seamlessly with:

- `TaskDetailSheet` - Shows plan review interface
- `TaskEditForm` - Fallback for basic plan editing
- `KanbanBoard` - Status transitions
- WebSocket updates for real-time plan updates

## Future Enhancements

Planned improvements:

- Enhanced markdown support (tables, diagrams)
- Collaborative editing
- Plan templates
- Advanced export formats (PDF, HTML)
- Plan comparison/diff view
- Integration with version control
