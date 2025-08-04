import { useWebSocketConnection } from '@/context/websocket-context'
import { ConnectionStatus } from '@/components/ui/connection-status'
import { Separator } from '@/components/ui/separator'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from '@/components/ui/sidebar'
import { NavGroup } from '@/components/layout/nav-group'
import { NavUser } from '@/components/layout/nav-user'
import { TeamSwitcher } from '@/components/layout/team-switcher'
import { ProjectSelector } from '@/components/project-selector'
import { sidebarData } from './data/sidebar-data'

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const { connectionState, queuedMessageCount, reconnect, clearMessageQueue } =
    useWebSocketConnection()

  return (
    <Sidebar collapsible='icon' variant='floating' {...props}>
      <SidebarHeader>
        <TeamSwitcher teams={sidebarData.teams} />
        {/* <div className="px-2 py-2">
          <ProjectSelector className="w-full" />
        </div> */}
        {/* <div className='px-2 py-1'>
          <ConnectionStatus
            connectionState={connectionState}
            queuedMessageCount={queuedMessageCount}
            onReconnect={reconnect}
            onClearQueue={clearMessageQueue}
            variant='compact'
          />
        </div> */}
        <Separator />
      </SidebarHeader>
      <SidebarContent>
        {sidebarData.navGroups.map((props) => (
          <NavGroup key={props.title} {...props} />
        ))}
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={sidebarData.user} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
