import {
  IconBrowserCheck,
  IconLayoutDashboard,
  IconNotification,
  IconPackages,
  IconPalette,
  IconSettings,
  IconTool,
  IconUserCog,
} from '@tabler/icons-react'
import { Command } from 'lucide-react'
import { type SidebarData } from '../types'

export const sidebarData: SidebarData = {
  user: {
    name: 'Auto-Devs User',
    email: 'user@auto-devs.com',
    avatar: '/avatars/shadcn.jpg',
  },
  teams: [
    {
      name: 'Auto-Devs',
      logo: Command,
      plan: 'AI Development Platform',
    },
  ],
  navGroups: [
    {
      title: 'General',
      items: [
        {
          title: 'Projects',
          url: '/projects',
          icon: IconPackages,
        },
      ],
    },
  ],
}
