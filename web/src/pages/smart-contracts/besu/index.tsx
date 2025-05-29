import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Plus, FileCode, Package, CheckCircle2, AlertCircle } from 'lucide-react'
import { useState } from 'react'
import { Badge } from '@/components/ui/badge'

// Mock data for demonstration
const mockContracts = [
  {
    id: '1',
    name: 'SimpleStorage',
    description: 'Simple key-value storage contract',
    versions: [
      {
        version: '1.0.0',
        status: 'deployed',
        network: 'besu-network-1',
        address: '0x1234...5678',
        deployedAt: '2024-03-15T10:00:00Z',
        transactionHash: '0xabcd...efgh',
      },
      {
        version: '1.1.0',
        status: 'pending',
        network: 'besu-network-1',
        address: '0x8765...4321',
        deployedAt: '2024-03-16T15:30:00Z',
        transactionHash: '0xijkl...mnop',
      },
    ],
  },
  {
    id: '2',
    name: 'TokenContract',
    description: 'ERC20 token implementation',
    versions: [
      {
        version: '1.0.0',
        status: 'deployed',
        network: 'besu-network-1',
        address: '0x9876...5432',
        deployedAt: '2024-03-14T09:15:00Z',
        transactionHash: '0xqrst...uvwx',
      },
    ],
  },
]

const statusIcons = {
  deployed: <CheckCircle2 className="h-4 w-4 text-green-500" />,
  pending: <AlertCircle className="h-4 w-4 text-yellow-500" />,
}

const statusLabels = {
  deployed: 'Deployed',
  pending: 'Pending',
}

export default function BesuContractsPage() {
  const [isDeployDialogOpen, setIsDeployDialogOpen] = useState(false)
  const [selectedContract, setSelectedContract] = useState<string | null>(null)

  return (
    <div className="flex-1 p-8">
      <div className="mb-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-semibold">Besu Smart Contracts</h1>
            <p className="text-muted-foreground">Deploy and manage Ethereum smart contracts on your Besu networks</p>
          </div>
          <Dialog open={isDeployDialogOpen} onOpenChange={setIsDeployDialogOpen}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="mr-2 h-4 w-4" />
                Deploy Contract
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Deploy Smart Contract</DialogTitle>
                <DialogDescription>Upload and deploy a new smart contract to your network.</DialogDescription>
              </DialogHeader>
              <div className="space-y-4">
                <div className="border-2 border-dashed rounded-lg p-8 text-center">
                  <FileCode className="mx-auto h-12 w-12 text-muted-foreground" />
                  <p className="mt-2 text-sm text-muted-foreground">Drag and drop your Solidity contract files here</p>
                  <Button variant="outline" className="mt-4">
                    Browse Files
                  </Button>
                </div>
              </div>
              <DialogFooter>
                <Button>Deploy Contract</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      <div className="space-y-6">
        {mockContracts.map((contract) => (
          <Card key={contract.id}>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle className="flex items-center gap-2">
                    <FileCode className="h-5 w-5" />
                    {contract.name}
                  </CardTitle>
                  <CardDescription>{contract.description}</CardDescription>
                </div>
                <Button variant="outline" size="sm" onClick={() => setSelectedContract(contract.id)}>
                  <Package className="mr-2 h-4 w-4" />
                  New Version
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {contract.versions.map((version, index) => (
                  <div key={index} className="flex items-center justify-between border-b pb-4 last:border-0 last:pb-0">
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <span className="font-medium">Version {version.version}</span>
                        <Badge variant="outline" className="flex items-center gap-1">
                          {statusIcons[version.status as keyof typeof statusIcons]}
                          {statusLabels[version.status as keyof typeof statusLabels]}
                        </Badge>
                      </div>
                      <div className="text-sm text-muted-foreground">
                        <p>Network: {version.network}</p>
                        <p>Address: {version.address}</p>
                        <p>Deployed on: {new Date(version.deployedAt).toLocaleString()}</p>
                        <p>Transaction: {version.transactionHash}</p>
                      </div>
                    </div>
                    <div className="flex gap-2">
                      <Button variant="outline" size="sm">
                        View Details
                      </Button>
                      <Button variant="outline" size="sm">
                        Interact
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
} 