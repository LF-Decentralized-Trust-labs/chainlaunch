import { FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Trash2 } from 'lucide-react'
import { z } from 'zod'
import { useFormContext } from 'react-hook-form'

// Schema for the UpdateBatchTimeoutPayload
export const updateBatchTimeoutSchema = z.object({
  timeout: z.string().min(1, "Timeout is required")
})

export type UpdateBatchTimeoutFormValues = z.infer<typeof updateBatchTimeoutSchema>

interface UpdateBatchTimeoutOperationProps {
  index: number
  onRemove: () => void
}

export function UpdateBatchTimeoutOperation({ index, onRemove }: UpdateBatchTimeoutOperationProps) {
  const formContext = useFormContext()

  return (
    <Card className="mb-6">
      <CardHeader className="pb-3">
        <div className="flex justify-between items-center">
          <CardTitle className="text-lg font-medium">Update Batch Timeout</CardTitle>
          <Button 
            variant="ghost" 
            size="icon" 
            onClick={onRemove}
            className="h-8 w-8 text-destructive"
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <FormField
            control={formContext.control}
            name={`operations.${index}.payload.timeout`}
            render={({ field }) => (
              <FormItem>
                <FormLabel>Timeout</FormLabel>
                <FormControl>
                  <Input placeholder="2s" {...field} />
                </FormControl>
                <FormMessage className="text-xs">
                  Format examples: 500ms, 1s, 2m
                </FormMessage>
              </FormItem>
            )}
          />
        </div>
      </CardContent>
    </Card>
  )
} 