import { createFileRoute } from '@tanstack/react-router'
import SettingsAppearance from '@/pages/settings/appearance'

export const Route = createFileRoute('/_authenticated/settings/appearance')({
  component: SettingsAppearance,
})
