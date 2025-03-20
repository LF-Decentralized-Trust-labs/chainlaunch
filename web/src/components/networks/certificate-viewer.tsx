import { Card } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Copy } from "lucide-react"
import { toast } from "sonner"

interface CertificateViewerProps {
  title: string
  certificate: string
}

export function CertificateViewer({ title, certificate }: CertificateViewerProps) {
  const handleCopy = () => {
    navigator.clipboard.writeText(certificate)
    toast.success("Certificate copied to clipboard")
  }

  return (
    <Card className="p-4">
      <div className="flex items-center justify-between mb-2">
        <h3 className="text-sm font-medium">{title}</h3>
        <Button variant="ghost" size="sm" onClick={handleCopy}>
          <Copy className="h-4 w-4" />
        </Button>
      </div>
      <pre className="text-xs bg-muted p-4 rounded-md overflow-x-auto">
        {certificate}
      </pre>
    </Card>
  )
} 