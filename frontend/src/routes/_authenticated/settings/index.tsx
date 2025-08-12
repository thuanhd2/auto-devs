import { createFileRoute } from '@tanstack/react-router'
import Settings from '@/pages/settings'

export const Route = createFileRoute('/_authenticated/settings/')({
  component: Settings,
})
