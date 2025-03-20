import { Card } from "@/components/ui/card"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Badge } from "@/components/ui/badge"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
import { cn } from "@/lib/utils"

interface SingleKeySelectProps {
  keys: Array<{
    id?: number
    name?: string
    algorithm?: string
    createdAt?: string
  }>
  value?: number
  onChange: (value: number) => void
}

export function SingleKeySelect({ keys, value, onChange }: SingleKeySelectProps) {
  return (
    <Card className="border-dashed">
      <ScrollArea className="h-[200px] p-4">
        <RadioGroup value={value?.toString()} onValueChange={(value) => onChange(Number(value))}>
          <div className="space-y-4">
            {keys.map((key) => (
              <div
                key={key.id}
                className={cn(
                  'flex items-center space-x-4 rounded-md border p-4',
                  value === key.id && 'border-primary'
                )}
              >
                <RadioGroupItem value={key.id!.toString()} id={key.id!.toString()} />
                <div className="flex-1 space-y-1">
                  <p className="text-sm font-medium leading-none">{key.name}</p>
                  <p className="text-sm text-muted-foreground">
                    Created {new Date(key.createdAt!).toLocaleDateString()}
                  </p>
                </div>
                <Badge variant="outline">{key.algorithm}</Badge>
              </div>
            ))}
          </div>
        </RadioGroup>
      </ScrollArea>
    </Card>
  )
} 