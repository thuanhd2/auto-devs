import { createFileRoute } from '@tanstack/react-router'
import { ProjectBoard } from '@/components/kanban/project-board'

export const Route = createFileRoute('/_authenticated/projects/')({
  component: ProjectsPage,
})

function ProjectsPage() {
  // For now, we'll use a hardcoded project ID
  // In a real app, this would come from route params or user selection
  const projectId = 'demo-project-id'

  return (
    <div className="h-full">
      <ProjectBoard projectId={projectId} />
    </div>
  )
}