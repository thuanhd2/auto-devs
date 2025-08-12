import { createFileRoute } from '@tanstack/react-router'
import GithubIntegration from '@/pages/settings/github-integration'

export const Route = createFileRoute(
  '/_authenticated/settings/github-integration'
)({
  component: GithubIntegration,
})
