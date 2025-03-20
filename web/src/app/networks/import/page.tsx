import { ImportNetworkForm } from "@/components/network-import/ImportNetworkForm"

export default function ImportNetworkPage() {
  return (
    <div className="container mx-auto py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold tracking-tight">Import Network</h1>
        <p className="text-muted-foreground">
          Import an existing Hyperledger Fabric or Besu network using a genesis block
        </p>
      </div>
      
      <ImportNetworkForm />
    </div>
  )
} 