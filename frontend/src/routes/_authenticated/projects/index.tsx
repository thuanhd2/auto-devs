import { createFileRoute } from '@tanstack/react-router'
import ComingSoon from '@/components/coming-soon'

export const Route = createFileRoute('/_authenticated/projects/')({
  component: ProjectsPage,
})

function ProjectsPage() {
  return (
    <div className='space-y-4'>
      <div className='flex items-center justify-between'>
        <h1 className='text-3xl font-bold tracking-tight'>Projects</h1>
      </div>
      <ComingSoon />
    </div>
  )
}