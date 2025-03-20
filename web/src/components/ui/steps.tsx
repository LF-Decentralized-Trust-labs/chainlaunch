import { cn } from '@/lib/utils'
import { CheckCircle2 } from 'lucide-react'

interface Step {
  id: string
  title: string
}

interface StepsProps {
  steps: Step[]
  currentStep: string
  className?: string
}

export function Steps({ steps, currentStep, className }: StepsProps) {
  return (
    <nav aria-label="Progress" className={className}>
      <ol role="list" className="space-y-4 md:flex md:space-y-0 md:space-x-8">
        {steps.map((step, stepIdx) => {
          const isCurrent = currentStep === step.id
          const isComplete = steps.findIndex(s => s.id === currentStep) > stepIdx

          return (
            <li key={step.title} className="md:flex-1">
              <div
                className={cn(
                  "group flex flex-col border-l-4 py-2 pl-4 md:border-l-0 md:border-t-4 md:pl-0 md:pt-4 md:pb-0",
                  isCurrent && "border-primary",
                  isComplete && "border-primary",
                  !isCurrent && !isComplete && "border-border"
                )}
              >
                <span className="text-sm font-medium">
                  {isComplete && <CheckCircle2 className="mr-1.5 h-4 w-4 inline text-primary" />}
                  Step {stepIdx + 1}
                </span>
                <span className={cn(
                  "text-sm",
                  (isCurrent || isComplete) && "text-primary font-medium",
                  !isCurrent && !isComplete && "text-muted-foreground"
                )}>
                  {step.title}
                </span>
              </div>
            </li>
          )
        })}
      </ol>
    </nav>
  )
} 