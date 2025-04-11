import { FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Trash2 } from 'lucide-react'
import { z } from 'zod'
import { useFormContext } from 'react-hook-form'

// Schema for the UpdateConsenterPayload
export const updateConsenterSchema = z.object({
  host: z.string().min(1, "Current host is required"),
  port: z.number().int().positive("Current port must be a positive integer"),
  new_host: z.string().min(1, "New host is required"),
  new_port: z.number().int().positive("New port must be a positive integer"),
  client_tls_cert: z.string().min(1, "Client TLS certificate is required"),
  server_tls_cert: z.string().min(1, "Server TLS certificate is required")
})

export type UpdateConsenterFormValues = z.infer<typeof updateConsenterSchema>

interface UpdateConsenterOperationProps {
  index: number
  onRemove: () => void
}

export function UpdateConsenterOperation({ index, onRemove }: UpdateConsenterOperationProps) {
  const formContext = useFormContext()

  return (
    <Card className="mb-6">
      <CardHeader className="pb-3">
        <div className="flex justify-between items-center">
          <CardTitle className="text-lg font-medium">Update Consenter</CardTitle>
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
        <div className="space-y-6">
          <div>
            <h3 className="text-sm font-medium mb-2">Current Consenter</h3>
            <div className="grid grid-cols-2 gap-4">
              <FormField
                control={formContext.control}
                name={`operations.${index}.payload.host`}
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Current Host</FormLabel>
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
                    <FormLabel>Current Port</FormLabel>
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
          </div>
          
          <div>
            <h3 className="text-sm font-medium mb-2">New Consenter</h3>
            <div className="grid grid-cols-2 gap-4">
              <FormField
                control={formContext.control}
                name={`operations.${index}.payload.new_host`}
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>New Host</FormLabel>
                    <FormControl>
                      <Input placeholder="orderer0.example.com" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              
              <FormField
                control={formContext.control}
                name={`operations.${index}.payload.new_port`}
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>New Port</FormLabel>
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
          </div>
          
          <FormField
            control={formContext.control}
            name={`operations.${index}.payload.client_tls_cert`}
            render={({ field }) => (
              <FormItem>
                <FormLabel>Client TLS Certificate</FormLabel>
                <FormControl>
                  <Textarea 
                    placeholder="-----BEGIN CERTIFICATE-----" 
                    className="font-mono text-xs h-24"
                    {...field} 
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          
          <FormField
            control={formContext.control}
            name={`operations.${index}.payload.server_tls_cert`}
            render={({ field }) => (
              <FormItem>
                <FormLabel>Server TLS Certificate</FormLabel>
                <FormControl>
                  <Textarea 
                    placeholder="-----BEGIN CERTIFICATE-----" 
                    className="font-mono text-xs h-24"
                    {...field} 
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