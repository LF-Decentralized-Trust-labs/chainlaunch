import { Button } from '@/components/ui/button'
import { Link } from 'react-router-dom'
import { NodeCreationWizard } from '@/components/node-creation-wizard'

export default function CreateNodePage() {
  return (
    <div className="flex-1 p-8">
      <div className="max-w-4xl mx-auto">
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-2xl font-semibold">Create Node</h1>
            <p className="text-muted-foreground">Set up a new blockchain node</p>
          </div>
          <Button variant="outline" asChild>
            <Link to="/nodes">Back to Nodes</Link>
          </Button>
        </div>

        <NodeCreationWizard />
      </div>
    </div>
  )
} 