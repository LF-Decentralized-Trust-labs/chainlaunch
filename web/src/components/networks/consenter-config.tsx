import { Badge } from "@/components/ui/badge"
import { Card } from "@/components/ui/card"
import { Network } from "lucide-react"

interface ConsenterConfigProps {
  consenters: Array<{
    host: string
    port: number
    client_tls_cert: string
    server_tls_cert: string
  }>
}

export function ConsenterConfig({ consenters }: ConsenterConfigProps) {
  return (
    <div className="space-y-4">
      {consenters.map((consenter, index) => (
        <Card key={index} className="p-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="h-8 w-8 rounded-md bg-primary/10 flex items-center justify-center">
                <Network className="h-4 w-4 text-primary" />
              </div>
              <div>
                <h4 className="font-medium">{consenter.host}</h4>
                <p className="text-sm text-muted-foreground">Port: {consenter.port}</p>
              </div>
            </div>
            <Badge variant="outline">Active</Badge>
          </div>
        </Card>
      ))}
    </div>
  )
} 