import { useWebSocketConnection } from '@/context/websocket-context'
import { ConnectionStatus } from '@/components/ui/connection-status'
import { Separator } from '@/components/ui/separator'
import { Sidebar, SidebarContent, SidebarRail } from '@/components/ui/sidebar'
import { NavGroup } from '@/components/layout/nav-group'
import { sidebarData } from './data/sidebar-data'

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const { connectionState, queuedMessageCount, reconnect, clearMessageQueue } =
    useWebSocketConnection()

  return (
    <Sidebar collapsible='icon' variant='floating' {...props}>
      <SidebarContent>
        {sidebarData.navGroups.map((props) => (
          <NavGroup key={props.title} {...props} />
        ))}
      </SidebarContent>
      <SidebarRail />
    </Sidebar>
  )
}
