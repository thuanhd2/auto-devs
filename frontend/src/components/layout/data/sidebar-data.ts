import {
  IconHome,
  IconUsers,
  IconFileText,
  IconSettings,
} from '@tabler/icons-react'

export const sidebarData = {
  navGroups: [
    {
      title: 'Main',
      items: [
        {
          title: 'Dashboard',
          href: '/dashboard',
          icon: IconHome,
        },
        {
          title: 'Projects',
          href: '/projects',
          icon: IconFileText,
        },
        {
          title: 'Team',
          href: '/team',
          icon: IconUsers,
        },
        {
          title: 'Settings',
          href: '/settings',
          icon: IconSettings,
        },
      ],
    },
  ],
}
