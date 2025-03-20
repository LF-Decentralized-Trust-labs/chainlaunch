import { Alert, AlertDescription } from "@/components/ui/alert"
import { AlertTriangle } from "lucide-react"

interface OrgAnchorWarningProps {
  organizationName: string
}

export function OrgAnchorWarning({ organizationName }: OrgAnchorWarningProps) {
  return (
    <Alert variant="warning" className="w-[300px]">
      <AlertTriangle className="h-4 w-4" />
      <AlertDescription className="text-sm">
        {organizationName} has no anchor peers
      </AlertDescription>
    </Alert>
  )
} 