import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { ChevronRight, Settings, Shield, Users, Network, Server } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useState } from 'react'

interface ChannelConfigCardProps {
	config: any // We'll use any for now, but you should define a proper type
}

export function ChannelConfigCard({ config }: ChannelConfigCardProps) {
	const [openSections, setOpenSections] = useState<string[]>([])

	const toggleSection = (section: string) => {
		setOpenSections((prev) => (prev.includes(section) ? prev.filter((s) => s !== section) : [...prev, section]))
	}
	const channelGroup = config?.data?.data?.[0]?.payload?.data?.config?.channel_group
	const consensusType = channelGroup?.groups?.Orderer?.values?.ConsensusType?.value
	const batchSize = channelGroup?.groups?.Orderer?.values?.BatchSize?.value
	const batchTimeout = channelGroup?.groups?.Orderer?.values?.BatchTimeout?.value

	if (!channelGroup) return null

	const renderPolicies = (policies: any) => {
		return Object.entries(policies || {}).map(([name, policy]: [string, any]) => (
			<div key={name} className="space-y-1">
				<div className="flex items-center justify-between">
					<div className="flex items-center gap-2">
						<Shield className="h-4 w-4 text-muted-foreground" />
						<span className="font-medium">{name}</span>
					</div>
					<Badge variant="outline">{policy.policy?.type === 1 ? 'Signature' : policy.policy?.type === 3 ? 'ImplicitMeta' : 'Unknown'}</Badge>
				</div>
				{policy.policy?.type === 3 && (
					<div className="text-sm text-muted-foreground pl-6">
						Rule: {policy.policy.value.rule} of {policy.policy.value.sub_policy}
					</div>
				)}
			</div>
		))
	}

	const renderEndpoints = (endpoints: string[]) => {
		return (
			<div className="space-y-2 pl-6">
				{endpoints.map((endpoint, index) => (
					<div key={index} className="flex items-center gap-2 text-sm text-muted-foreground">
						<Server className="h-4 w-4" />
						<span>{endpoint}</span>
					</div>
				))}
			</div>
		)
	}

	const renderConsenters = (consenters: any[]) => {
		return (
			<div className="space-y-2 pl-6">
				{consenters.map((consenter, index) => (
					<div key={index} className="flex items-center gap-2 text-sm text-muted-foreground">
						<Network className="h-4 w-4" />
						<span>{`${consenter.host}:${consenter.port}`}</span>
					</div>
				))}
			</div>
		)
	}

	const renderAnchorPeers = (anchorPeers: any[]) => {
		return (
			<div className="space-y-2 pl-6">
				{anchorPeers.map((peer, index) => (
					<div key={index} className="flex items-center gap-2 text-sm text-muted-foreground">
						<Users className="h-4 w-4" />
						<span>{`${peer.host}:${peer.port}`}</span>
					</div>
				))}
			</div>
		)
	}

	const renderOrganizations = (organizations: any) => {
		return Object.entries(organizations || {}).map(([mspId, org]: [string, any]) => (
			<Collapsible key={mspId} open={openSections.includes(mspId)} onOpenChange={() => toggleSection(mspId)}>
				<CollapsibleTrigger className="flex items-center gap-2 w-full hover:bg-muted/50 p-2 rounded-md">
					<ChevronRight className={cn('h-4 w-4 transition-transform', openSections.includes(mspId) && 'transform rotate-90')} />
					<Users className="h-4 w-4" />
					<span className="font-medium">{mspId}</span>
				</CollapsibleTrigger>
				<CollapsibleContent className="pl-8 pr-4 pb-2 space-y-4">
					<div className="space-y-4 pt-2">
						<div>
							<h4 className="text-sm font-medium mb-2">Policies</h4>
							<div className="space-y-3">{renderPolicies(org.policies)}</div>
						</div>
						{org.values?.Endpoints && (
							<div>
								<h4 className="text-sm font-medium mb-2">Endpoints</h4>
								{renderEndpoints(org.values.Endpoints.value.addresses)}
							</div>
						)}
						{org.values?.AnchorPeers && (
							<div>
								<h4 className="text-sm font-medium mb-2">Anchor Peers</h4>
								{renderAnchorPeers(org.values.AnchorPeers.value.anchor_peers)}
							</div>
						)}
					</div>
				</CollapsibleContent>
			</Collapsible>
		))
	}

	return (
		<Card className="p-6">
			<div className="flex items-center gap-4 mb-6">
				<div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center">
					<Settings className="h-6 w-6 text-primary" />
				</div>
				<div>
					<h2 className="text-lg font-semibold">Channel Configuration</h2>
					<p className="text-sm text-muted-foreground">Channel policies and organization details</p>
				</div>
			</div>

			<ScrollArea className="h-[400px] pr-4">
				<div className="space-y-6">
					{consensusType && (
						<div>
							<h3 className="text-sm font-medium mb-3">Consensus Configuration</h3>
							<div className="space-y-3">
								<div className="space-y-2">
									<div className="text-sm">Type: {consensusType.type}</div>
									<div className="text-sm">State: {consensusType.state}</div>
									{consensusType.metadata?.options && (
										<div className="space-y-1">
											<div className="text-sm font-medium">Options:</div>
											<div className="pl-4 space-y-1 text-sm text-muted-foreground">
												<div>Election Tick: {consensusType.metadata.options.election_tick}</div>
												<div>Heartbeat Tick: {consensusType.metadata.options.heartbeat_tick}</div>
												<div>Max Inflight Blocks: {consensusType.metadata.options.max_inflight_blocks}</div>
												<div>Tick Interval: {consensusType.metadata.options.tick_interval}</div>
											</div>
										</div>
									)}
									{consensusType.metadata?.consenters && (
										<div>
											<div className="text-sm font-medium mb-2">Consenters:</div>
											{renderConsenters(consensusType.metadata.consenters)}
										</div>
									)}
								</div>
								{batchSize && (
									<div className="space-y-1">
										<div className="text-sm font-medium">Batch Size:</div>
										<div className="pl-4 space-y-1 text-sm text-muted-foreground">
											<div>Max Message Count: {batchSize.max_message_count}</div>
											<div>Absolute Max Bytes: {batchSize.absolute_max_bytes}</div>
											<div>Preferred Max Bytes: {batchSize.preferred_max_bytes}</div>
										</div>
									</div>
								)}
								{batchTimeout && (
									<div className="space-y-1">
										<div className="text-sm font-medium">Batch Timeout:</div>
										<div className="pl-4 text-sm text-muted-foreground">{batchTimeout.timeout}</div>
									</div>
								)}
							</div>
						</div>
					)}

					<div>
						<h3 className="text-sm font-medium mb-3">Application Organizations</h3>
						<div className="space-y-2">{renderOrganizations(channelGroup?.groups?.Application?.groups)}</div>
					</div>

					<div>
						<h3 className="text-sm font-medium mb-3">Orderer Organizations</h3>
						<div className="space-y-2">{renderOrganizations(channelGroup?.groups?.Orderer?.groups)}</div>
					</div>

					<div>
						<h3 className="text-sm font-medium mb-3">Channel Policies</h3>
						<div className="space-y-3">{renderPolicies(channelGroup?.policies)}</div>
					</div>
				</div>
			</ScrollArea>
		</Card>
	)
}
