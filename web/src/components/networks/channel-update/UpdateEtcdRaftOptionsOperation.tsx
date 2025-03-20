import { FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Trash2 } from 'lucide-react'
import { z } from 'zod'
import { useFormContext } from 'react-hook-form'

// Schema for the UpdateEtcdRaftOptionsPayload
export const updateEtcdRaftOptionsSchema = z.object({
  election_tick: z.number().int().positive("Election tick must be a positive integer"),
  heartbeat_tick: z.number().int().positive("Heartbeat tick must be a positive integer"),
  max_inflight_blocks: z.number().int().positive("Max inflight blocks must be a positive integer"),
  snapshot_interval_size: z.number().int().positive("Snapshot interval size must be a positive integer"),
  tick_interval: z.string().min(1, "Tick interval is required")
})

export type UpdateEtcdRaftOptionsFormValues = z.infer<typeof updateEtcdRaftOptionsSchema>

interface UpdateEtcdRaftOptionsOperationProps {
  index: number
  onRemove: () => void
}

export function UpdateEtcdRaftOptionsOperation({ index, onRemove }: UpdateEtcdRaftOptionsOperationProps) {
  const formContext = useFormContext()

  return (
    <Card className="mb-6">
      <CardHeader className="pb-3">
        <div className="flex justify-between items-center">
          <CardTitle className="text-lg font-medium">Update Etcd Raft Options</CardTitle>
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
          <div className="grid grid-cols-2 gap-4">
            <FormField
              control={formContext.control}
              name={`operations.${index}.payload.election_tick`}
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Election Tick</FormLabel>
                  <FormControl>
                    <Input 
                      type="number" 
                      placeholder="10" 
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
              name={`operations.${index}.payload.heartbeat_tick`}
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Heartbeat Tick</FormLabel>
                  <FormControl>
                    <Input 
                      type="number" 
                      placeholder="1" 
                      {...field} 
                      onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
          
          <div className="grid grid-cols-2 gap-4">
            <FormField
              control={formContext.control}
              name={`operations.${index}.payload.max_inflight_blocks`}
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Max Inflight Blocks</FormLabel>
                  <FormControl>
                    <Input 
                      type="number" 
                      placeholder="5" 
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
              name={`operations.${index}.payload.snapshot_interval_size`}
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Snapshot Interval Size</FormLabel>
                  <FormControl>
                    <Input 
                      type="number" 
                      placeholder="16777216" 
                      {...field} 
                      onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
          
          <FormField
            control={formContext.control}
            name={`operations.${index}.payload.tick_interval`}
            render={({ field }) => (
              <FormItem>
                <FormLabel>Tick Interval</FormLabel>
                <FormControl>
                  <Input placeholder="500ms" {...field} />
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