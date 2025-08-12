import { createFileRoute } from '@tanstack/react-router'
import CodeEditor from '@/pages/settings/code-editor'

export const Route = createFileRoute('/_authenticated/settings/code-editor')({
  component: CodeEditor,
})
