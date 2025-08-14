import { useState, useEffect } from 'react'
import { formatDistanceToNow } from 'date-fns'
import { Users, Eye, Edit3 } from 'lucide-react'
import { useWebSocketProject } from '@/context/websocket-context'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'

interface UserPresenceProps {
  projectId: string
  showDetails?: boolean
  maxAvatars?: number
}

interface UserActivity {
  userId: string
  username: string
  activity: 'viewing' | 'editing' | 'idle'
  lastActive: Date
  currentTask?: string
}

/**
 * Component that shows who is currently active in the project
 */
export function UserPresence({
  projectId,
  showDetails = true,
  maxAvatars = 5,
}: UserPresenceProps) {
  const { onlineUsers, userCount } = useWebSocketProject(projectId)
  const [userActivities, setUserActivities] = useState<
    Map<string, UserActivity>
  >(new Map())

  // Convert onlineUsers map to array with activity info
  const activeUsers = Array.from(onlineUsers.entries()).map(
    ([userId, user]) => {
      const activity = userActivities.get(userId)
      return {
        userId,
        username: user.username,
        joinedAt: user.joinedAt,
        activity: activity?.activity || 'viewing',
        lastActive: activity?.lastActive || user.joinedAt,
        currentTask: activity?.currentTask,
      }
    }
  )

  // Sort users by activity and last active time
  const sortedUsers = activeUsers.sort((a, b) => {
    // Prioritize editors over viewers
    if (a.activity === 'editing' && b.activity !== 'editing') return -1
    if (b.activity === 'editing' && a.activity !== 'editing') return 1

    // Then by last active time
    return b.lastActive.getTime() - a.lastActive.getTime()
  })

  const visibleUsers = sortedUsers.slice(0, maxAvatars)
  const hiddenCount = Math.max(0, sortedUsers.length - maxAvatars)

  // Simulate user activity updates (in a real app, this would come from WebSocket)
  useEffect(() => {
    const interval = setInterval(() => {
      setUserActivities((prev) => {
        const updated = new Map(prev)

        // Randomly update user activities for demo
        activeUsers.forEach((user) => {
          const activities: UserActivity['activity'][] = [
            'viewing',
            'editing',
            'idle',
          ]
          const randomActivity =
            activities[Math.floor(Math.random() * activities.length)]

          updated.set(user.userId, {
            userId: user.userId,
            username: user.username,
            activity: randomActivity,
            lastActive: new Date(),
            currentTask:
              randomActivity === 'editing' ? 'Sample Task' : undefined,
          })
        })

        return updated
      })
    }, 10000) // Update every 10 seconds

    return () => clearInterval(interval)
  }, [activeUsers, setUserActivities])

  if (userCount === 0) {
    return null
  }

  return (
    <TooltipProvider>
      <div className='flex items-center gap-2'>
        <div className='flex items-center'>
          <Users className='mr-1 h-4 w-4 text-gray-500' />
          <span className='text-sm text-gray-600'>{userCount}</span>
        </div>

        <div className='flex -space-x-2'>
          {visibleUsers.map((user) => (
            <UserAvatar key={user.userId} user={user} />
          ))}

          {hiddenCount > 0 && (
            <Tooltip>
              <TooltipTrigger>
                <div className='flex h-8 w-8 items-center justify-center rounded-full border-2 border-white bg-gray-200 text-xs font-medium text-gray-600'>
                  +{hiddenCount}
                </div>
              </TooltipTrigger>
              <TooltipContent>
                <p>
                  {hiddenCount} more user{hiddenCount > 1 ? 's' : ''}
                </p>
              </TooltipContent>
            </Tooltip>
          )}
        </div>

        {showDetails && visibleUsers.length > 0 && (
          <div className='ml-2 hidden items-center gap-1 md:flex'>
            {visibleUsers.slice(0, 3).map((user) => (
              <UserActivityBadge key={user.userId} user={user} />
            ))}
          </div>
        )}
      </div>
    </TooltipProvider>
  )
}

/**
 * Individual user avatar with presence indicator
 */
function UserAvatar({ user }: { user: any }) {
  const getActivityColor = (activity: string) => {
    switch (activity) {
      case 'editing':
        return 'bg-green-500'
      case 'viewing':
        return 'bg-blue-500'
      case 'idle':
        return 'bg-yellow-500'
      default:
        return 'bg-gray-500'
    }
  }

  const getActivityIcon = (activity: string) => {
    switch (activity) {
      case 'editing':
        return <Edit3 className='h-2 w-2' />
      case 'viewing':
        return <Eye className='h-2 w-2' />
      default:
        return null
    }
  }

  return (
    <Tooltip>
      <TooltipTrigger>
        <div className='relative'>
          <Avatar className='h-8 w-8 border-2 border-white'>
            <AvatarImage
              src={`https://api.dicebear.com/7.x/avatars/svg?seed=${user.username}`}
            />
            <AvatarFallback className='text-xs'>
              {user.username.slice(0, 2).toUpperCase()}
            </AvatarFallback>
          </Avatar>

          {/* Activity indicator */}
          <div
            className={`absolute -right-0.5 -bottom-0.5 h-3 w-3 rounded-full border border-white ${getActivityColor(user.activity)} flex items-center justify-center`}
          >
            {getActivityIcon(user.activity)}
          </div>
        </div>
      </TooltipTrigger>
      <TooltipContent>
        <div className='text-center'>
          <p className='font-medium'>{user.username}</p>
          <p className='text-xs text-gray-500 capitalize'>{user.activity}</p>
          {user.currentTask && (
            <p className='text-xs text-gray-400'>
              Working on: {user.currentTask}
            </p>
          )}
          <p className='text-xs text-gray-400'>
            Joined {formatDistanceToNow(user.joinedAt)} ago
          </p>
        </div>
      </TooltipContent>
    </Tooltip>
  )
}

/**
 * Activity badge showing what users are doing
 */
function UserActivityBadge({ user }: { user: any }) {
  const getBadgeVariant = (activity: string) => {
    switch (activity) {
      case 'editing':
        return 'default'
      case 'viewing':
        return 'secondary'
      case 'idle':
        return 'outline'
      default:
        return 'secondary'
    }
  }

  return (
    <Badge variant={getBadgeVariant(user.activity)} className='text-xs'>
      {user.username}: {user.activity}
    </Badge>
  )
}

/**
 * Compact user presence counter for smaller spaces
 */
export function UserPresenceCompact({ projectId }: { projectId: string }) {
  const { userCount } = useWebSocketProject(projectId)

  if (userCount === 0) {
    return null
  }

  return (
    <div className='flex items-center gap-1 text-sm text-gray-500'>
      <Users className='h-4 w-4' />
      <span>{userCount}</span>
    </div>
  )
}
