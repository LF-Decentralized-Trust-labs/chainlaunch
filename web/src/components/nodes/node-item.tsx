import { HttpNodeResponse } from '@/api/client'
import { BesuIcon } from '@/components/icons/besu-icon'
import { FabricIcon } from '@/components/icons/fabric-icon'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { MoreVertical, Network, Trash } from 'lucide-react'

interface NodeItemProps {
	node: HttpNodeResponse
	onDelete: (node: HttpNodeResponse) => void
}

export function NodeItem({ node, onDelete }: NodeItemProps) {
	const NodeIcon = node.platform === 'FABRIC' ? FabricIcon : node.platform === 'BESU' ? BesuIcon : Network

	return (
		<Card className="p-4">
			<div className="flex items-center justify-between">
				<div className="flex items-center gap-4">
					<div className="h-10 w-10 rounded-lg bg-primary/10 flex items-center justify-center">
						<NodeIcon className="h-5 w-5 text-primary" />
					</div>
					<div>
						<h3 className="font-medium">{node.name}</h3>
						<p className="text-sm text-muted-foreground">
							{node.platform} • {node.status}
						</p>
					</div>
				</div>
				<DropdownMenu>
					<DropdownMenuTrigger asChild>
						<Button variant="ghost" size="icon">
							<MoreVertical className="h-4 w-4" />
						</Button>
					</DropdownMenuTrigger>
					<DropdownMenuContent align="end">
						<DropdownMenuItem className="text-destructive" onSelect={() => onDelete(node)}>
							<Trash className="h-4 w-4 mr-2" />
							Delete
						</DropdownMenuItem>
					</DropdownMenuContent>
				</DropdownMenu>
			</div>
			<div className="mt-4 flex gap-2">
				<div className="text-xs px-2 py-1 rounded-md bg-primary/10 text-primary">ID: {node.id}</div>
				{node.fabricPeer?.externalEndpoint && <div className="text-xs px-2 py-1 rounded-md bg-muted">Endpoint: {node.fabricPeer.externalEndpoint}</div>}
				{node.fabricOrderer?.externalEndpoint && <div className="text-xs px-2 py-1 rounded-md bg-muted">Endpoint: {node.fabricOrderer.externalEndpoint}</div>}
				{node.besuNode?.externalIp && <div className="text-xs px-2 py-1 rounded-md bg-muted">IP: {node.besuNode.externalIp}</div>}
			</div>
		</Card>
	)
}
