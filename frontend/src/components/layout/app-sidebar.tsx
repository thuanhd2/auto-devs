import { Link, useLocation } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { sidebarData } from './data/sidebar-data'

export function AppSidebar() {
  return (
    <div className='bg-background flex h-full w-64 flex-col border-r'>
      <div className='p-4'>
        <h2 className='text-lg font-semibold'>Auto-Devs</h2>
      </div>
      <Separator />
      <ScrollArea className='flex-1'>
        <div className='space-y-2 p-4'>
          {sidebarData.navGroups.map((group) => (
            <div key={group.title} className='space-y-2'>
              <h3 className='text-muted-foreground text-sm font-medium'>
                {group.title}
              </h3>
              {group.items.map((item) => (
                <Link
                  key={item.href}
                  to={item.href}
                  className={cn(
                    'hover:bg-accent hover:text-accent-foreground flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors',
                    useLocation().pathname === item.href &&
                      'bg-accent text-accent-foreground'
                  )}
                >
                  {item.icon && <item.icon className='h-4 w-4' />}
                  {item.title}
                </Link>
              ))}
            </div>
          ))}
        </div>
      </ScrollArea>
    </div>
  )
}
