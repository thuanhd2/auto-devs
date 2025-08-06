import { useState } from 'react'
import type { Task } from '@/types/task'
import { Check, X, Edit, Download, FileText, Eye } from 'lucide-react'
import { toast } from 'sonner'
import { useUpdateTask, useApprovePlan } from '@/hooks/use-tasks'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { PlanEditor } from './plan-editor'
import { PlanPreview } from './plan-preview'

interface PlanReviewProps {
  task: Task
  onPlanUpdate?: (updatedTask: Task) => void
  onStatusChange?: (taskId: string, newStatus: Task['status']) => void
}

export function PlanReview({
  task,
  onPlanUpdate,
  onStatusChange,
}: PlanReviewProps) {
  const firstPlan = task.plans && task.plans.length > 0 ? task.plans[0] : null
  const planContent = firstPlan?.content || ''
  const [isEditing, setIsEditing] = useState(false)
  const [editedPlan, setEditedPlan] = useState(planContent)
  const updateTaskMutation = useUpdateTask()
  const approvePlanMutation = useApprovePlan()

  const isLoading = updateTaskMutation.isPending || approvePlanMutation.isPending
  const canReview = task.status === 'PLAN_REVIEWING'
  const hasPlan = Boolean(planContent?.trim())

  const handleApprovePlan = async () => {
    try {
      await approvePlanMutation.mutateAsync(task.id)
      onStatusChange?.(task.id, 'IMPLEMENTING')
      // The success toast and task updates are handled by the mutation hook
    } catch (error) {
      // Error handled by mutation hook
    }
  }