import { createFileRoute } from '@tanstack/react-router'
import SettingsAiExecutor from '@/pages/settings/ai-executor'

export const Route = createFileRoute('/_authenticated/settings/ai-executor')({
  component: SettingsAiExecutor,
})
