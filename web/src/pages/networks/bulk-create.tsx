import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Steps } from '@/components/ui/steps'
import { zodResolver } from '@hookform/resolvers/zod'
import { ArrowLeft, ArrowRight, Server } from 'lucide-react'
import { useForm } from 'react-hook-form'
import { Link } from 'react-router-dom'
import * as z from 'zod'
import { ProtocolSelector } from '@/components/protocol-selector'

const steps = [
  { id: 'protocol', title: 'Protocol' },
  { id: 'nodes', title: 'Number of Nodes' },
  { id: 'network', title: 'Network Configuration' },
  { id: 'nodes-config', title: 'Nodes Configuration' },
  { id: 'review', title: 'Review & Create' },
]

const formSchema = z.object({
  protocol: z.string().min(1, 'Please select a protocol'),
  numberOfNodes: z.number().min(1).max(10),
})

type FormValues = z.infer<typeof formSchema>

export default function BulkCreateNetworkPage() {
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      numberOfNodes: 4,
    },
  })

  const onSubmit = (data: FormValues) => {
    console.log(data)
    // Handle form submission
  }

  return (
    <div className="flex-1 p-8">
      <div className="max-w-3xl mx-auto">
        <div className="flex items-center gap-2 text-muted-foreground mb-8">
          <Button variant="ghost" size="sm" asChild>
            <Link to="/networks">
              <ArrowLeft className="mr-2 h-4 w-4" />
              Networks
            </Link>
          </Button>
        </div>

        <div className="flex items-center gap-4 mb-8">
          <Server className="h-8 w-8" />
          <div>
            <h1 className="text-2xl font-semibold">Create Network</h1>
            <p className="text-muted-foreground">Create a new blockchain network</p>
          </div>
        </div>

        <Steps steps={steps} currentStep="protocol" className="mb-8" />

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
            <Card className="p-6">
              <div className="space-y-6">
                <ProtocolSelector control={form.control} name="protocol" />

                <FormField
                  control={form.control}
                  name="numberOfNodes"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Number of Nodes</FormLabel>
                      <FormControl>
                        <Input 
                          type="number" 
                          min={1} 
                          max={10} 
                          {...field} 
                          onChange={(e) => field.onChange(parseInt(e.target.value))} 
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            </Card>

            <div className="flex justify-between">
              <Button variant="outline" asChild>
                <Link to="/networks">Cancel</Link>
              </Button>
              <Button type="submit">
                Next
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </div>
          </form>
        </Form>
      </div>
    </div>
  )
} 