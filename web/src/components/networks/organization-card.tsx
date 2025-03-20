import { ServiceNetworkNode } from '@/api/client'
import { getOrganizationsByIdOptions } from '@/api/client/@tanstack/react-query.gen'
import { Card } from '@/components/ui/card'
import { useQuery } from '@tanstack/react-query'
import { Users } from 'lucide-react'
import { Skeleton } from '../ui/skeleton'
import { NodeCard } from './node-card'

interface OrganizationCardProps {
	organizationId: number
	nodes: Array<ServiceNetworkNode>
	networkId: number
	onJoined: () => void
	onUnjoined: () => void
}

export function OrganizationCard({ organizationId, nodes, networkId, onJoined, onUnjoined }: OrganizationCardProps) {
	const { data: organization, isLoading } = useQuery({
		...getOrganizationsByIdOptions({
			path: { id: organizationId },
		}),
	})

	if (isLoading) {
		return (
			<Card className="p-4">
				<div className="flex items-center gap-4">
					<Skeleton className="h-10 w-10 rounded-full" />
					<div className="space-y-2">
						<Skeleton className="h-5 w-40" />
						<Skeleton className="h-4 w-24" />
					</div>
				</div>
			</Card>
		)
	}

	if (!organization) return null

	return (
		<Card className="p-4">
			<div className="flex items-center gap-4 mb-4">
				<div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center">
					<Users className="h-5 w-5 text-primary" />
				</div>
				<div>
					<h3 className="font-medium">{organization.mspId!}</h3>
					<p className="text-sm text-muted-foreground">MSP ID: {organization.mspId}</p>
				</div>
			</div>
			<div className="space-y-4 pl-14">
				{nodes.map((node) => (
					<NodeCard key={node.id} networkNode={node} networkId={networkId} onJoined={onJoined} onUnjoined={onUnjoined} />
				))}
			</div>
		</Card>
	)
}
