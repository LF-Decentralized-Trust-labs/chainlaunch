import { FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Trash2 } from 'lucide-react'
import { z } from 'zod'
import { useFormContext } from 'react-hook-form'

// Schema for the RemoveOrgPayload
export const removeOrgSchema = z.object({
  msp_id: z.string().min(1, "MSP ID is required")
})

export type RemoveOrgFormValues = z.infer<typeof removeOrgSchema>

interface RemoveOrgOperationProps {
  index: number
  onRemove: () => void
}

export function RemoveOrgOperation({ index, onRemove }: RemoveOrgOperationProps) {
  const formContext = useFormContext()

  return (
    <Card className="mb-6">
      <CardHeader className="pb-3">
        <div className="flex justify-between items-center">
          <CardTitle className="text-lg font-medium">Remove Organization</CardTitle>
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
            name={`operations.${index}.payload.msp_id`}
            render={({ field }) => (
              <FormItem>
                <FormLabel>MSP ID</FormLabel>
                <FormControl>
                  <Input placeholder="Org1MSP" {...field} />
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