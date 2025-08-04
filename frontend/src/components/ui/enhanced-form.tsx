import * as React from 'react'
import { useForm, UseFormReturn, FieldValues } from 'react-hook-form'
import { cn } from '@/lib/utils'
import { Progress } from './progress'
import { LoadingButton } from './loading-button'
import { Alert, AlertDescription } from './alert'
import { CheckCircle, AlertCircle, Clock } from 'lucide-react'

export interface FormStep {
  id: string
  title: string
  description?: string
  fields: string[]
  validation?: (values: any) => boolean | string[]
}

export interface EnhancedFormProps<T extends FieldValues> {
  children: React.ReactNode
  form: UseFormReturn<T>
  onSubmit: (values: T) => Promise<void> | void
  steps?: FormStep[]
  showProgress?: boolean
  className?: string
  disabled?: boolean
  submitText?: string
  submitLoadingText?: string
  successMessage?: string
  errorMessage?: string
  showValidationSummary?: boolean
}

export function EnhancedForm<T extends FieldValues>({
  children,
  form,
  onSubmit,
  steps,
  showProgress = false,
  className,
  disabled = false,
  submitText = 'Submit',
  submitLoadingText = 'Submitting...',
  successMessage,
  errorMessage,
  showValidationSummary = true,
}: EnhancedFormProps<T>) {
  const [isSubmitting, setIsSubmitting] = React.useState(false)
  const [submitError, setSubmitError] = React.useState<string | null>(null)
  const [submitSuccess, setSubmitSuccess] = React.useState(false)

  const { formState: { errors, isValid, touchedFields } } = form

  // Calculate form completion progress
  const getFormProgress = React.useCallback(() => {
    if (!steps) return 0

    const allFields = steps.flatMap(step => step.fields)
    const filledFields = allFields.filter(fieldName => {
      const value = form.getValues(fieldName as keyof T)
      return value !== undefined && value !== '' && value !== null
    })

    return allFields.length > 0 ? (filledFields.length / allFields.length) * 100 : 0
  }, [form, steps])

  // Get current step based on filled fields
  const getCurrentStep = React.useCallback(() => {
    if (!steps) return 0

    for (let i = 0; i < steps.length; i++) {
      const step = steps[i]
      const isStepValid = step.validation 
        ? step.validation(form.getValues()) === true
        : step.fields.every(fieldName => {
            const value = form.getValues(fieldName as keyof T)
            return value !== undefined && value !== '' && value !== null
          })
      
      if (!isStepValid) return i
    }

    return steps.length - 1
  }, [form, steps])

  // Get validation summary
  const getValidationSummary = React.useCallback(() => {
    const errorFields = Object.keys(errors)
    const totalFields = steps ? steps.flatMap(s => s.fields).length : Object.keys(form.getValues()).length
    const validFields = totalFields - errorFields.length

    return {
      total: totalFields,
      valid: validFields,
      invalid: errorFields.length,
      errors: errorFields.map(field => errors[field]?.message).filter(Boolean),
    }
  }, [errors, form, steps])

  const handleSubmit = async (values: T) => {
    setIsSubmitting(true)
    setSubmitError(null)
    setSubmitSuccess(false)

    try {
      await onSubmit(values)
      setSubmitSuccess(true)
      if (successMessage) {
        // Success message is handled by parent component or notification system
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : 'An error occurred'
      setSubmitError(errorMessage || message)
    } finally {
      setIsSubmitting(false)
    }
  }

  const progress = getFormProgress()
  const currentStep = getCurrentStep()
  const validationSummary = getValidationSummary()

  return (
    <div className={cn('space-y-6', className)}>
      {/* Progress Section */}
      {showProgress && steps && (
        <div className="space-y-4">
          <div className="flex items-center justify-between text-sm">
            <span className="font-medium">Form Progress</span>
            <span className="text-muted-foreground">{Math.round(progress)}% complete</span>
          </div>
          
          <Progress value={progress} className="h-2" />
          
          {/* Step indicators */}
          <div className="flex items-center justify-between">
            {steps.map((step, index) => {
              const isCompleted = index < currentStep
              const isCurrent = index === currentStep
              const isUpcoming = index > currentStep

              return (
                <div key={step.id} className="flex flex-col items-center space-y-1">
                  <div
                    className={cn(
                      'flex h-8 w-8 items-center justify-center rounded-full text-xs font-medium',
                      isCompleted && 'bg-green-500 text-white',
                      isCurrent && 'bg-blue-500 text-white',
                      isUpcoming && 'bg-muted text-muted-foreground'
                    )}
                  >
                    {isCompleted ? (
                      <CheckCircle className="h-4 w-4" />
                    ) : isCurrent ? (
                      <Clock className="h-4 w-4" />
                    ) : (
                      index + 1
                    )}
                  </div>
                  <span className="text-xs text-center max-w-20">{step.title}</span>
                </div>
              )
            })}
          </div>
        </div>
      )}

      {/* Validation Summary */}
      {showValidationSummary && validationSummary.invalid > 0 && (
        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            {validationSummary.invalid} field{validationSummary.invalid === 1 ? '' : 's'} {validationSummary.invalid === 1 ? 'needs' : 'need'} attention:
            <ul className="mt-2 space-y-1">
              {validationSummary.errors.map((error, index) => (
                <li key={index} className="text-sm">• {error}</li>
              ))}
            </ul>
          </AlertDescription>
        </Alert>
      )}

      {/* Success Message */}
      {submitSuccess && successMessage && (
        <Alert className="border-green-500 bg-green-50">
          <CheckCircle className="h-4 w-4 text-green-600" />
          <AlertDescription className="text-green-800">
            {successMessage}
          </AlertDescription>
        </Alert>
      )}

      {/* Error Message */}
      {submitError && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{submitError}</AlertDescription>
        </Alert>
      )}

      {/* Form Content */}
      <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
        {children}
        
        {/* Submit Button */}
        <div className="flex justify-end pt-4">
          <LoadingButton
            type="submit"
            loading={isSubmitting}
            loadingText={submitLoadingText}
            disabled={disabled || isSubmitting}
            className="min-w-32"
          >
            {submitText}
          </LoadingButton>
        </div>
      </form>
    </div>
  )
}

// Hook for managing multi-step forms
export function useMultiStepForm<T extends FieldValues>(
  steps: FormStep[],
  form: UseFormReturn<T>
) {
  const [currentStepIndex, setCurrentStepIndex] = React.useState(0)

  const currentStep = steps[currentStepIndex]
  const isFirstStep = currentStepIndex === 0
  const isLastStep = currentStepIndex === steps.length - 1

  const canProceedToNext = React.useCallback(() => {
    if (!currentStep) return false
    
    return currentStep.fields.every(fieldName => {
      const fieldError = form.formState.errors[fieldName as keyof T]
      const fieldValue = form.getValues(fieldName as keyof T)
      
      return !fieldError && fieldValue !== undefined && fieldValue !== '' && fieldValue !== null
    })
  }, [currentStep, form])

  const goToNext = React.useCallback(() => {
    if (!isLastStep && canProceedToNext()) {
      setCurrentStepIndex(prev => prev + 1)
    }
  }, [isLastStep, canProceedToNext])

  const goToPrevious = React.useCallback(() => {
    if (!isFirstStep) {
      setCurrentStepIndex(prev => prev - 1)
    }
  }, [isFirstStep])

  const goToStep = React.useCallback((stepIndex: number) => {
    if (stepIndex >= 0 && stepIndex < steps.length) {
      setCurrentStepIndex(stepIndex)
    }
  }, [steps.length])

  const progress = ((currentStepIndex + 1) / steps.length) * 100

  return {
    currentStep,
    currentStepIndex,
    isFirstStep,
    isLastStep,
    canProceedToNext: canProceedToNext(),
    goToNext,
    goToPrevious,
    goToStep,
    progress,
    totalSteps: steps.length,
  }
}

// Component for rendering step navigation
interface StepNavigationProps {
  onNext: () => void
  onPrevious: () => void
  canProceedToNext: boolean
  isFirstStep: boolean
  isLastStep: boolean
  nextText?: string
  previousText?: string
  disabled?: boolean
}

export function StepNavigation({
  onNext,
  onPrevious,
  canProceedToNext,
  isFirstStep,
  isLastStep,
  nextText = 'Next',
  previousText = 'Previous',
  disabled = false,
}: StepNavigationProps) {
  return (
    <div className="flex justify-between pt-4">
      <LoadingButton
        type="button"
        variant="outline"
        onClick={onPrevious}
        disabled={isFirstStep || disabled}
      >
        {previousText}
      </LoadingButton>
      
      {!isLastStep && (
        <LoadingButton
          type="button"
          onClick={onNext}
          disabled={!canProceedToNext || disabled}
        >
          {nextText}
        </LoadingButton>
      )}
    </div>
  )
}