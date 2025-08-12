import { createFileRoute } from '@tanstack/react-router'
import AboutUs from '@/pages/about-us'

export const Route = createFileRoute('/_authenticated/about-us/')({
  component: AboutUs,
})
