import { useState } from 'react'
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Trash2, Plus } from 'lucide-react'
import { z } from 'zod'
import { useFieldArray, useFormContext } from 'react-hook-form'

// Schema for the SetAnchorPeersPayload
export const setAnchorPeersSchema = z.object({
  msp_id: z.string().min(1, "MSP ID is required"),
  anchor_peers: z.array(
    z.object({
      host: z.string().min(1, "Host is required"),
      port: z.number().int().positive("Port must be a positive integer")
    })
  ).min(1, "At least one anchor peer is required")
})

export type SetAnchorPeersFormValues = z.infer<typeof setAnchorPeersSchema>

interface SetAnchorPeersOperationProps {
  index: number
  onRemove: () => void
}

export function SetAnchorPeersOperation({ index, onRemove }: SetAnchorPeersOperationProps) {
  const formContext = useFormContext()

  const { fields: anchorPeersFields, append: appendAnchorPeer, remove: removeAnchorPeer } = 
    useFieldArray({
      name: `operations.${index}.payload.anchor_peers`,
      control: formContext.control
    })

  const handleAddAnchorPeer = () => {
    appendAnchorPeer({ host: '', port: 7051 })
  }

  return (
    <Card className="mb-6">
      <CardHeader className="pb-3">
        <div className="flex justify-between items-center">
          <CardTitle className="text-lg font-medium">Set Anchor Peers</CardTitle>
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

          <div className="space-y-4">
            <div className="flex justify-between items-center">
              <FormLabel>Anchor Peers</FormLabel>
              <Button 
                type="button" 
                variant="outline" 
                size="sm" 
                onClick={handleAddAnchorPeer}
                className="h-8"
              >
                <Plus className="h-4 w-4 mr-1" />
                Add Peer
              </Button>
            </div>
            
            {anchorPeersFields.map((field, i) => (
              <div key={field.id} className="flex gap-4 items-start border p-3 rounded-md">
                <div className="flex-1 space-y-3">
                  <FormField
                    control={formContext.control}
                    name={`operations.${index}.payload.anchor_peers.${i}.host`}
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Host</FormLabel>
                        <FormControl>
                          <Input placeholder="peer0.org1.example.com" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  
                  <FormField
                    control={formContext.control}
                    name={`operations.${index}.payload.anchor_peers.${i}.port`}
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Port</FormLabel>
                        <FormControl>
                          <Input 
                            type="number" 
                            placeholder="7051" 
                            {...field} 
                            onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
                
                <Button 
                  type="button" 
                  variant="ghost" 
                  size="icon" 
                  onClick={() => removeAnchorPeer(i)}
                  className="h-8 w-8 mt-8 text-destructive"
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))}
            
            {anchorPeersFields.length === 0 && (
              <div className="text-center p-4 border border-dashed rounded-md">
                <p className="text-sm text-muted-foreground">No anchor peers added yet</p>
                <Button 
                  type="button" 
                  variant="outline" 
                  size="sm" 
                  onClick={handleAddAnchorPeer}
                  className="mt-2"
                >
                  <Plus className="h-4 w-4 mr-1" />
                  Add Peer
                </Button>
              </div>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
} 