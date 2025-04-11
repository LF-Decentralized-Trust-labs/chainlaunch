import { FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Trash2 } from 'lucide-react'
import { z } from 'zod'
import { useFormContext } from 'react-hook-form'

// Schema for the RemoveConsenterPayload
export const removeConsenterSchema = z.object({
  host: z.string().min(1, "Host is required"),
  port: z.number().int().positive("Port must be a positive integer")
})

export type RemoveConsenterFormValues = z.infer<typeof removeConsenterSchema>

interface RemoveConsenterOperationProps {
  index: number
  onRemove: () => void
}

export function RemoveConsenterOperation({ index, onRemove }: RemoveConsenterOperationProps) {
  const formContext = useFormContext()

  return (
    <Card className="mb-6">
      <CardHeader className="pb-3">
        <div className="flex justify-between items-center">
          <CardTitle className="text-lg font-medium">Remove Consenter</CardTitle>
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
        <div className="grid grid-cols-2 gap-4">
          <FormField
            control={formContext.control}
            name={`operations.${index}.payload.host`}
            render={({ field }) => (
              <FormItem>
                <FormLabel>Host</FormLabel>
                <FormControl>
                  <Input placeholder="orderer0.example.com" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          
          <FormField
            control={formContext.control}
            name={`operations.${index}.payload.port`}
            render={({ field }) => (
              <FormItem>
                <FormLabel>Port</FormLabel>
                <FormControl>
                  <Input 
                    type="number" 
                    placeholder="7050" 
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