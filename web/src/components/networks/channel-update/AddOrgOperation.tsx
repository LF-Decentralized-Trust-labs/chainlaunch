import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Trash2 } from 'lucide-react'
import { useState } from 'react'
import { useFieldArray, useFormContext } from 'react-hook-form'
import { z } from 'zod'

// Schema for the AddOrgPayload
export const addOrgSchema = z.object({
  msp_id: z.string().min(1, "MSP ID is required"),
  root_certs: z.array(z.string()).min(1, "At least one root certificate is required"),
  tls_root_certs: z.array(z.string()).min(1, "At least one TLS root certificate is required")
})

export type AddOrgFormValues = z.infer<typeof addOrgSchema>

interface AddOrgOperationProps {
  index: number
  onRemove: () => void
}

export function AddOrgOperation({ index, onRemove }: AddOrgOperationProps) {
  const formContext = useFormContext()
  const [newRootCert, setNewRootCert] = useState('')
  const [newTlsRootCert, setNewTlsRootCert] = useState('')

  const { fields: rootCertsFields, append: appendRootCert, remove: removeRootCert } = 
    useFieldArray({
      name: `operations.${index}.payload.root_certs`,
      control: formContext.control
    })

  const { fields: tlsRootCertsFields, append: appendTlsRootCert, remove: removeTlsRootCert } = 
    useFieldArray({
      name: `operations.${index}.payload.tls_root_certs`,
      control: formContext.control
    })

  const handleAddRootCert = () => {
    if (newRootCert.trim()) {
      appendRootCert(newRootCert.trim())
      setNewRootCert('')
    }
  }

  const handleAddTlsRootCert = () => {
    if (newTlsRootCert.trim()) {
      appendTlsRootCert(newTlsRootCert.trim())
      setNewTlsRootCert('')
    }
  }

  return (
    <Card className="mb-6">
      <CardHeader className="pb-3">
        <div className="flex justify-between items-center">
          <CardTitle className="text-lg font-medium">Add Organization</CardTitle>
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

          <div className="space-y-2">
            <FormLabel>Root Certificates</FormLabel>
            {rootCertsFields.map((field, i) => (
              <div key={field.id} className="flex gap-2">
                <Input 
                  {...formContext.register(`operations.${index}.payload.root_certs.${i}`)}
                  className="flex-1"
                />
                <Button 
                  type="button" 
                  variant="ghost" 
                  size="icon" 
                  onClick={() => removeRootCert(i)}
                  className="h-10 w-10 text-destructive"
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))}
            <div className="flex gap-2">
              <Input 
                value={newRootCert}
                onChange={(e) => setNewRootCert(e.target.value)}
                placeholder="Paste PEM certificate"
                className="flex-1"
              />
              <Button 
                type="button" 
                onClick={handleAddRootCert}
                className="whitespace-nowrap"
              >
                Add Certificate
              </Button>
            </div>
          </div>

          <div className="space-y-2">
            <FormLabel>TLS Root Certificates</FormLabel>
            {tlsRootCertsFields.map((field, i) => (
              <div key={field.id} className="flex gap-2">
                <Input 
                  {...formContext.register(`operations.${index}.payload.tls_root_certs.${i}`)}
                  className="flex-1"
                />
                <Button 
                  type="button" 
                  variant="ghost" 
                  size="icon" 
                  onClick={() => removeTlsRootCert(i)}
                  className="h-10 w-10 text-destructive"
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))}
            <div className="flex gap-2">
              <Input 
                value={newTlsRootCert}
                onChange={(e) => setNewTlsRootCert(e.target.value)}
                placeholder="Paste PEM certificate"
                className="flex-1"
              />
              <Button 
                type="button" 
                onClick={handleAddTlsRootCert}
                className="whitespace-nowrap"
              >
                Add Certificate
              </Button>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
} 