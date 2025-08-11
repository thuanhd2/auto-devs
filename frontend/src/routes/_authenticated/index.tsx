import { createFileRoute, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated/')({
  beforeLoad: () => {
    // Redirect từ /abc sang /xyz
    throw redirect({
      to: '/projects',
    })
  },
})
