import { FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Trash2 } from 'lucide-react'
import { z } from 'zod'
import { useFormContext } from 'react-hook-form'

// Schema for the UpdateBatchSizePayload
export const updateBatchSizeSchema = z.object({
  absolute_max_bytes: z.number().int().positive("Absolute max bytes must be a positive integer"),
  max_message_count: z.number().int().positive("Max message count must be a positive integer"),
  preferred_max_bytes: z.number().int().positive("Preferred max bytes must be a positive integer")
})

export type UpdateBatchSizeFormValues = z.infer<typeof updateBatchSizeSchema>

interface UpdateBatchSizeOperationProps {
  index: number
  onRemove: () => void
}

export function UpdateBatchSizeOperation({ index, onRemove }: UpdateBatchSizeOperationProps) {
  const formContext = useFormContext()

  return (
    <Card className="mb-6">
      <CardHeader className="pb-3">
        <div className="flex justify-between items-center">
          <CardTitle className="text-lg font-medium">Update Batch Size</CardTitle>
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
            name={`operations.${index}.payload.absolute_max_bytes`}
            render={({ field }) => (
              <FormItem>
                <FormLabel>Absolute Max Bytes</FormLabel>
                <FormControl>
                  <Input 
                    type="number" 
                    placeholder="10485760" 
                    {...field} 
                    onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          
          <FormField
            control={formContext.control}
            name={`operations.${index}.payload.max_message_count`}
            render={({ field }) => (
              <FormItem>
                <FormLabel>Max Message Count</FormLabel>
                <FormControl>
                  <Input 
                    type="number" 
                    placeholder="500" 
                    {...field} 
                    onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          
          <FormField
            control={formContext.control}
            name={`operations.${index}.payload.preferred_max_bytes`}
            render={({ field }) => (
              <FormItem>
                <FormLabel>Preferred Max Bytes</FormLabel>
                <FormControl>
                  <Input 
                    type="number" 
                    placeholder="2097152" 
                    {...field} 
                    onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>
      </CardContent>
    </Card>
  )
} 