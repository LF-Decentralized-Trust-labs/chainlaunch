import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Plus, FileCode, Package, CheckCircle2, AlertCircle, Clock } from 'lucide-react'
import { useState, useEffect } from 'react'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { useNavigate } from 'react-router-dom'

const statusIcons = {
  committed: <CheckCircle2 className="h-4 w-4 text-green-500" />,
  approved: <Clock className="h-4 w-4 text-yellow-500" />,
  installed: <AlertCircle className="h-4 w-4 text-blue-500" />,
}

const statusLabels = {
  committed: 'Committed',
  approved: 'Approved',
  installed: 'Installed',
}

const chaincodeFormSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  networkId: z.string().min(1, 'Network is required'),
})

type ChaincodeFormValues = z.infer<typeof chaincodeFormSchema>

export default function FabricChaincodesPage() {
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false)
  const [chaincodeDefs, setChaincodeDefs] = useState<any[]>([])
  const [networks, setNetworks] = useState<any[]>([])
  const form = useForm<ChaincodeFormValues>({
    resolver: zodResolver(chaincodeFormSchema),
    defaultValues: { name: '', networkId: '' },
  })
  const navigate = useNavigate()

  // Fetch networks (mock for now, replace with API call)
  useEffect(() => {
    // TODO: Replace with real API call
    setNetworks([
      { id: '1', name: 'Fabric Network 1' },
      { id: '2', name: 'Fabric Network 2' },
    ])
  }, [])

  const onSubmit = async (data: any) => {
    setChaincodeDefs((prev) => [
      ...prev,
      {
        id: Date.now().toString(),
        name: data.name,
        networkId: data.networkId,
        networkName: networks.find((n) => n.id === data.networkId)?.name || '',
        versions: [],
      },
    ])
    setIsCreateDialogOpen(false)
    form.reset()
  }

  return (
    <div className="flex-1 p-8">
      <div className="mb-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-semibold">Fabric Chaincode Definitions</h1>
            <p className="text-muted-foreground">Manage chaincode definitions for your Fabric networks</p>
          </div>
          <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="mr-2 h-4 w-4" />
                Create Chaincode Definition
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Create Chaincode Definition</DialogTitle>
                <DialogDescription>Define a new chaincode for your Fabric network.</DialogDescription>
              </DialogHeader>
              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                  <FormField
                    control={form.control}
                    name="name"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Name</FormLabel>
                        <FormControl>
                          <Input {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name="networkId"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Network</FormLabel>
                        <FormControl>
                          <select {...field} className="w-full border rounded p-2">
                            <option value="">Select a network</option>
                            {networks.map((n) => (
                              <option key={n.id} value={n.id}>{n.name}</option>
                            ))}
                          </select>
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <DialogFooter>
                    <Button type="submit">Create</Button>
                  </DialogFooter>
                </form>
              </Form>
            </DialogContent>
          </Dialog>
        </div>
      </div>
      <div className="space-y-6">
        {chaincodeDefs.length === 0 ? (
          <Card className="p-6 text-center text-muted-foreground">No chaincode definitions yet.</Card>
        ) : (
          chaincodeDefs.map((def) => (
            <Card key={def.id} className="p-6 flex items-center justify-between">
              <div>
                <div className="font-semibold text-lg">{def.name}</div>
                <div className="text-sm text-muted-foreground">Network: {def.networkName}</div>
              </div>
              <Button variant="outline" size="sm" onClick={() => navigate('definition', { state: { definition: def } })}>View Details</Button>
            </Card>
          ))
        )}
      </div>
    </div>
  )
} 