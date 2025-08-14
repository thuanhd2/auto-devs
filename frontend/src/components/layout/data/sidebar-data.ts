import {
  IconNotification,
  IconPackages,
  IconPalette,
  IconSettings,
  IconBrandGithub,
  IconRobot,
  IconCode,
  IconInfoCircle,
} from '@tabler/icons-react'
import { type SidebarData } from '../types'

export const sidebarData: SidebarData = {
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
    {
      title: 'Settings',
      items: [
        {
          title: 'Settings',
          icon: IconSettings,
          items: [
            {
              title: 'AI Executor',
              url: '/settings/ai-executor',
              icon: IconRobot,
            },
            {
              title: 'Github Integration',
              url: '/settings/github-integration',
              icon: IconBrandGithub,
            },
            {
              title: 'Appearance',
              url: '/settings/appearance',
              icon: IconPalette,
            },
            {
              title: 'Notifications',
              url: '/settings/notifications',
              icon: IconNotification,
            },
            {
              title: 'Code Editor',
              url: '/settings/code-editor',
              icon: IconCode,
            },
          ],
        },
        {
          title: 'About this project',
          url: '/about-us',
          icon: IconInfoCircle,
        },
      ],
    },
  ],
}
