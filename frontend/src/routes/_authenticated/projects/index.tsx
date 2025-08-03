import { createFileRoute } from '@tanstack/react-router'
import { ProjectList } from '@/pages/projects/ProjectList'

export const Route = createFileRoute('/_authenticated/projects/')({
  component: ProjectList,
})