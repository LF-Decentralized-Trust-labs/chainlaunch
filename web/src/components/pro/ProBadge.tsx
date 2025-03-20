import { Badge } from "@/components/ui/badge"
import { StarIcon } from "lucide-react"

export function ProBadge() {
  return (
    <Badge variant="outline" className="ml-auto text-xs font-medium bg-primary/10 text-primary border-primary/20 flex items-center gap-1 px-1.5 py-0.5">
      <StarIcon className="h-3 w-3" />
      <span>PRO</span>
    </Badge>
  )
} 