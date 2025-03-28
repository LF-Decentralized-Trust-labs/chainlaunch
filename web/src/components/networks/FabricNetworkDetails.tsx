import { HttpNetworkResponse } from '@/api/client'
import {
	getNetworksFabricByIdChannelConfigOptions,
	getNetworksFabricByIdCurrentChannelConfigOptions,
	getNetworksFabricByIdNodesOptions,
	getNodesOptions,
	getOrganizationsOptions,
	postNetworksFabricByIdAnchorPeersMutation,
	postNetworksFabricByIdOrderersByOrdererIdJoinMutation,
	postNetworksFabricByIdPeersByPeerIdJoinMutation,
} from '@/api/client/@tanstack/react-query.gen'
import { BesuIcon } from '@/components/icons/besu-icon'
import { FabricIcon } from '@/components/icons/fabric-icon'
import { AddNodeDialog } from '@/components/networks/add-node-dialog'
import { AnchorPeerConfig } from '@/components/networks/anchor-peer-config'
import { ChannelConfigCard } from '@/components/networks/channel-config-card'
import { ConsenterConfig } from '@/components/networks/consenter-config'
import { NetworkTabs, TabValue } from '@/components/networks/network-tabs'
import { NodeCard } from '@/components/networks/node-card'
import { OrgAnchorWarning } from '@/components/networks/org-anchor-warning'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { TimeAgo } from '@/components/ui/time-ago'
import { useMutation, useQuery } from '@tanstack/react-query'
import { Activity, AlertTriangle, Anchor, ArrowLeft, Check, Code, Copy, Network, Plus, Settings } from 'lucide-react'
import { useMemo, useState } from 'react'
import ReactMarkdown from 'react-markdown'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import SyntaxHighlighter from 'react-syntax-highlighter'
import { docco } from 'react-syntax-highlighter/dist/esm/styles/hljs'
import rehypeRaw from 'rehype-raw'
import { toast } from 'sonner'
import { AddMultipleNodesDialog } from './add-multiple-nodes-dialog'

interface FabricNetworkDetailsProps {
	network: HttpNetworkResponse
}

// Update the CHAINCODE_INSTRUCTIONS to be a function that takes parameters
const getChainCodeInstructions = (channelName: string, mspId: string) => {
	// Get the current origin and append /api/v1
	const apiUrl = typeof window !== 'undefined' ? `${window.location.origin}/api/v1` : 'http://localhost:8100/api/v1'

	return `
# Chaincode Installation Guide

## Clone the Repository

First, clone the chaincode repository:

\`\`\`bash
git clone https://github.com/kfs-learn/chaincode-typescript
cd chaincode-typescript
\`\`\`

## Install Required Tools

### Install bun.sh

We need to install bun.sh to run the project:

\`\`\`bash
curl -fsSL https://bun.sh/install | bash
\`\`\`

### Install Node.JS using NVM

First, install NVM:

\`\`\`bash
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.1/install.sh | bash
\`\`\`

Then, install Node.JS using NVM:

\`\`\`bash
nvm install v22
nvm use default v22
\`\`\`

### Install Dependencies

Install the project dependencies:

\`\`\`bash
bun install
\`\`\`

## Start Chaincode

### Pull Network Configuration

First, set up environment variables and pull the network configuration:

\`\`\`bash
export CHANNEL_NAME=${channelName}
export MSP_ID=${mspId}
export URL="${apiUrl}"
export CHAINLAUNCH_USERNAME=admin
export CHAINLAUNCH_PASSWORD="<chainlaunch_password>"

chainlaunch fabric network-config pull \\
    --network=$CHANNEL_NAME \\
    --msp-id=$MSP_ID \\
    --url=$URL \\
    --username="$CHAINLAUNCH_USERNAME" \\
    --password="$CHAINLAUNCH_PASSWORD" \\
    --output=network-config.yaml
\`\`\`

### Start the Chaincode Service

Set up additional environment variables and start the chaincode:

\`\`\`bash
export CHANNEL_NAME=${channelName}
export CHAINCODE_NAME=basic
export CHAINCODE_ADDRESS="localhost:9996"  # Chaincode listening address
export USER_NAME=admin
export MSP_ID=${mspId}

chainlaunch fabric install --local \\
    --config=$PWD/network-config.yaml \\
    --channel=$CHANNEL_NAME \\
    --chaincode=$CHAINCODE_NAME \\
    -o $MSP_ID -u $USER_NAME \\
    --policy="OR('\${MSP_ID}.member')" \\
    --chaincodeAddress="\${CHAINCODE_ADDRESS}" \\
    --envFile=$PWD/.env

bun run build
bun start:dev
\`\`\`

### Initialize and Test the Chaincode

Initialize the ledger and verify it's working:

\`\`\`bash
export CHANNEL_NAME=${channelName}
export CHAINCODE_NAME=basic
export MSP_ID=${mspId}

# Initialize the ledger
chainlaunch fabric invoke \\
    --chaincode=$CHAINCODE_NAME \\
    --config=network-config.yaml \\
    --channel $CHANNEL_NAME \\
    --fcn InitLedger \\
    --user=admin \\
    --mspID=$MSP_ID

# Query all assets to verify
chainlaunch fabric query \\
    --chaincode=$CHAINCODE_NAME \\
    --config=network-config.yaml \\
    --channel $CHANNEL_NAME \\
    --fcn GetAllAssets \\
    --user=admin \\
    --mspID=$MSP_ID
\`\`\`
`
}

function CopyButton({ text }: { text: string }) {
	const [copied, setCopied] = useState(false)

	const copy = () => {
		navigator.clipboard.writeText(text)
		setCopied(true)
		setTimeout(() => setCopied(false), 2000)
	}

	return (
		<button onClick={copy} className="absolute right-2 top-2 p-2 hover:bg-muted-foreground/20 rounded-md transition-colors">
			{copied ? <Check className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4 text-muted-foreground" />}
		</button>
	)
}

export default function FabricNetworkDetails({ network }: FabricNetworkDetailsProps) {
	const { id } = useParams()
	const [searchParams, setSearchParams] = useSearchParams()
	const currentTab = (searchParams.get('tab') as TabValue) || 'details'

	const { data: fabricOrgs, isLoading: fabricOrgsLoading } = useQuery({
		...getOrganizationsOptions({}),
	})
	const { data: genesisChannelConfig, isLoading: channelConfigLoading } = useQuery({
		...getNetworksFabricByIdChannelConfigOptions({
			path: { id: Number(id) },
		}),
	})
	const {
		data: currentChannelConfig,
		isLoading: currentChannelConfigLoading,
		refetch: refetchCurrentChannelConfig,
	} = useQuery({
		...getNetworksFabricByIdCurrentChannelConfigOptions({
			path: { id: Number(id) },
		}),
		retry: 0,
	})

	const channelConfig = useMemo(() => currentChannelConfig || genesisChannelConfig, [currentChannelConfig, genesisChannelConfig])

	const { data: nodes, isLoading: nodesLoading } = useQuery({
		...getNodesOptions({}),
	})

	const {
		data: networkNodes,
		isLoading: networkNodesLoading,
		refetch: refetchNetworkNodes,
	} = useQuery({
		...getNetworksFabricByIdNodesOptions({
			path: { id: Number(id) },
		}),
	})

	const joinPeerNode = useMutation({
		...postNetworksFabricByIdPeersByPeerIdJoinMutation(),
		onSuccess: () => {
			toast.success('Peer node joined successfully')
			refetchNetworkNodes()
		},
		onError: (error: any) => {
			toast.error('Failed to join peer node', {
				description: error.message,
			})
		},
	})

	const joinOrdererNode = useMutation({
		...postNetworksFabricByIdOrderersByOrdererIdJoinMutation(),
		onSuccess: () => {
			toast.success('Orderer node joined successfully')
			refetchNetworkNodes()
		},
		onError: (error: any) => {
			toast.error('Failed to join orderer node', {
				description: error.message,
			})
		},
	})

	const updateAnchorPeersMutation = useMutation({
		...postNetworksFabricByIdAnchorPeersMutation(),
		onSuccess: () => {
			refetchNetworkNodes()
			refetchCurrentChannelConfig()
		},
		onError: (error: any) => {
			toast.error('Failed to update anchor peers', {
				description: error.message,
			})
		},
	})

	const handleJoinAllNodes = async () => {
		if (!networkNodes || !network || !networkNodes.nodes) return

		const unjoindedNodes = networkNodes.nodes.filter((node) => node.status !== 'joined')

		try {
			const promises = unjoindedNodes.map((networkNode) => {
				const { node } = networkNode
				if (node?.nodeType === 'FABRIC_PEER') {
					return joinPeerNode.mutateAsync({
						path: { id: network.id!, peerId: node.id! },
					})
				} else if (node?.nodeType === 'FABRIC_ORDERER') {
					return joinOrdererNode.mutateAsync({
						path: { id: network.id!, ordererId: node.id! },
					})
				}
				return Promise.resolve()
			})

			await Promise.all(promises)
			toast.success('All nodes joined successfully')
		} catch (error: any) {
			toast.error('Failed to join some nodes', {
				description: error.message,
			})
		}
	}

	const availableNodes =
		nodes?.items
			?.filter((node) => !networkNodes?.nodes?.some((networkNode) => networkNode.node?.id === node.id))
			.map((node) => ({
				id: node.id!,
				name: node.name!,
				nodeType: node.nodeType!,
			})) ?? []

	const handleTabChange = (tab: TabValue) => {
		setSearchParams({ tab })
	}

	if (fabricOrgsLoading || channelConfigLoading || currentChannelConfigLoading || nodesLoading || networkNodesLoading) {
		return (
			<div className="flex-1 p-8">
				<div className="max-w-4xl mx-auto">
					<div className="mb-8">
						<Skeleton className="h-8 w-32 mb-2" />
						<Skeleton className="h-5 w-64" />
					</div>
					<div className="space-y-8">
						<Card className="p-6">
							<div className="space-y-4">
								<div className="flex items-center gap-4">
									<Skeleton className="h-12 w-12 rounded-lg" />
									<div>
										<Skeleton className="h-6 w-48 mb-2" />
										<Skeleton className="h-4 w-32" />
									</div>
								</div>
								<Skeleton className="h-24 w-full" />
							</div>
						</Card>
					</div>
				</div>
			</div>
		)
	}

	if (!network) {
		return (
			<div className="flex-1 p-8">
				<div className="max-w-4xl mx-auto text-center">
					<Network className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
					<h1 className="text-2xl font-semibold mb-2">Network not found</h1>
					<p className="text-muted-foreground mb-8">The network you're looking for doesn't exist or you don't have access to it.</p>
					<Button asChild>
						<Link to="/networks">
							<ArrowLeft className="mr-2 h-4 w-4" />
							Back to Networks
						</Link>
					</Button>
				</div>
			</div>
		)
	}

	const NetworkIcon = network.platform === 'FABRIC' ? FabricIcon : BesuIcon
	return (
		<div className="flex-1 p-8">
			<div className="max-w-4xl mx-auto">
				<div className="flex items-center gap-2 text-muted-foreground mb-8">
					<Button variant="ghost" size="sm" asChild>
						<Link to="/networks">
							<ArrowLeft className="mr-2 h-4 w-4" />
							Networks
						</Link>
					</Button>
				</div>

				<div className="mb-4">
					<div className="flex items-center justify-between">
						<div>
							<div className="flex items-center gap-3 mb-1">
								<h1 className="text-2xl font-semibold">{network.name}</h1>
								<Badge className="gap-1">
									<Activity className="h-3 w-3" />
									{network.status}
								</Badge>
							</div>
							<p className="text-muted-foreground">
								Created <TimeAgo date={network.createdAt!} />
							</p>
						</div>

						<div className="flex items-center gap-2">
							<AddNodeDialog networkId={network.id!} availableNodes={availableNodes} onNodeAdded={refetchNetworkNodes} />
							<AddMultipleNodesDialog networkId={network.id!} availableNodes={availableNodes} onNodesAdded={refetchNetworkNodes} />
							{networkNodes && networkNodes.nodes && networkNodes.nodes.some((node) => node.status !== 'joined') && (
								<Button size="sm" variant="outline" onClick={handleJoinAllNodes} disabled={joinPeerNode.isPending || joinOrdererNode.isPending}>
									<Plus className="mr-2 h-4 w-4" />
									Join All Nodes
								</Button>
							)}
							<Badge variant="outline" className="text-sm">
								{network.platform}
							</Badge>
						</div>
					</div>
				</div>

				{channelConfig?.config?.data?.data?.[0]?.payload?.data?.config?.channel_group?.groups?.Application?.groups && (
					<div className="mb-8 space-y-2">
						{Object.entries(channelConfig.config.data.data[0].payload.data.config.channel_group.groups.Application.groups)
							// Only show warnings for orgs that belong to us
							.filter(([mspId]) => fabricOrgs?.some((org) => org.mspId === mspId))
							.map(([mspId, orgConfig]) => {
								const anchorPeers = orgConfig.values?.AnchorPeers?.value?.anchor_peers || []

								if (anchorPeers.length === 0) {
									return (
										<Alert key={mspId} variant="warning" className="flex items-center justify-between">
											<div className="flex items-center">
												<AlertTriangle className="h-4 w-4" />
												<AlertDescription className="text-sm ml-2">{mspId} has no anchor peers</AlertDescription>
											</div>
											<Button variant="outline" size="sm" onClick={() => handleTabChange('anchor-peers')}>
												Configure Anchor Peers
											</Button>
										</Alert>
									)
								}
								return null
							})}
					</div>
				)}

				<Card className="p-6">
					<NetworkTabs
						tab={currentTab}
						setTab={handleTabChange}
						networkDetails={
							<div className="space-y-6">
								<div className="flex items-center gap-4 mb-6">
									<div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center">
										<NetworkIcon className="h-6 w-6 text-primary" />
									</div>
									<div>
										<h2 className="text-lg font-semibold">Network Information</h2>
										<p className="text-sm text-muted-foreground">Details about your blockchain network</p>
									</div>
								</div>

								{network.platform === 'fabric' && network && channelConfig && (
									<>
										<div>
											<h3 className="text-sm font-medium mb-2">Channel Name</h3>
											<p className="text-sm text-muted-foreground">{network.name}</p>
										</div>
										<div>
											<h3 className="text-sm font-medium mb-3">Nodes</h3>
											<div className="space-y-4">
												{networkNodes?.nodes?.map((node) => (
													<NodeCard key={node.id} networkNode={node} networkId={network.id!} onJoined={refetchNetworkNodes} onUnjoined={refetchNetworkNodes} />
												))}
											</div>
										</div>
										<ChannelConfigCard config={channelConfig.config} />
									</>
								)}

								{channelConfig?.config?.data?.data?.[0]?.payload?.data?.config?.channel_group?.groups?.Application?.groups && (
									<div className="space-y-4">
										{Object.entries(channelConfig.config.data.data[0].payload.data.config.channel_group.groups.Application.groups)
											// Only show warnings for orgs that belong to us
											.filter(([mspId]) => fabricOrgs?.some((org) => org.mspId === mspId))
											.map(([mspId, orgConfig]) => {
												const anchorPeers = orgConfig.values?.AnchorPeers?.value?.anchor_peers || []

												if (anchorPeers.length === 0) {
													return <OrgAnchorWarning key={mspId} organizationName={mspId} />
												}
												return null
											})}
									</div>
								)}
							</div>
						}
						anchorPeers={
							<div className="space-y-4">
								<div className="flex items-center gap-4 mb-6">
									<div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center">
										<Anchor className="h-6 w-6 text-primary" />
									</div>
									<div>
										<h2 className="text-lg font-semibold">Anchor Peers</h2>
										<p className="text-sm text-muted-foreground">Configure anchor peers for each organization</p>
									</div>
								</div>

								{channelConfig?.config?.data?.data?.[0]?.payload?.data?.config?.channel_group?.groups?.Application?.groups &&
									(() => {
										const filteredOrgs = Object.entries(channelConfig.config.data.data[0].payload.data.config.channel_group.groups.Application.groups).filter(([mspId]) =>
											fabricOrgs?.some((org) => org.mspId === mspId)
										)

										if (filteredOrgs.length === 0) {
											return (
												<Card className="p-6 flex flex-col items-center justify-center text-center">
													<AlertTriangle className="h-10 w-10 text-muted-foreground mb-4" />
													<h3 className="text-lg font-medium mb-2">No Peer Organizations Found</h3>
													<p className="text-sm text-muted-foreground mb-4">There are no peer organizations belonging to this node in the network.</p>
													<p className="text-xs text-muted-foreground">Peer organizations are required to configure anchor peers.</p>
												</Card>
											)
										}

										return filteredOrgs.map(([mspId, orgConfig]) => {
											const orgID = fabricOrgs?.find((org) => org.mspId === mspId)?.id!
											const organization = {
												id: orgID,
												name: mspId,
												mspId: mspId,
											}

											const currentAnchorPeers = orgConfig.values?.AnchorPeers?.value?.anchor_peers || []
											const orgNodes = networkNodes?.nodes?.filter((node) => node.node?.mspId === mspId) || []
											return (
												<AnchorPeerConfig
													key={mspId}
													organization={organization}
													peers={orgNodes}
													currentAnchorPeers={currentAnchorPeers}
													onUpdateAnchorPeers={async (newAnchorPeers) => {
														const updateAnchorPeersPromise = updateAnchorPeersMutation.mutateAsync({
															path: {
																id: network.id!,
															},
															body: {
																organizationId: orgID,
																anchorPeers: newAnchorPeers,
															},
														})
														await toast.promise(updateAnchorPeersPromise, {
															loading: 'Updating anchor peers...',
															success: 'Anchor peers updated successfully',
															error: (err) => `Failed to update anchor peers: ${err.message}`,
														})
													}}
												/>
											)
										})
									})()}
							</div>
						}
						consenters={
							<div className="space-y-4">
								<div className="flex items-center gap-4 mb-6">
									<div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center">
										<Settings className="h-6 w-6 text-primary" />
									</div>
									<div>
										<h2 className="text-lg font-semibold">Consenters</h2>
										<p className="text-sm text-muted-foreground">Manage network consenters configuration</p>
									</div>
								</div>

								{channelConfig?.config?.data?.data?.[0]?.payload?.data?.config?.channel_group?.groups?.Orderer?.values?.ConsensusType?.value?.metadata?.consenters ? (
									<ConsenterConfig consenters={channelConfig.config.data.data[0].payload.data.config.channel_group.groups.Orderer.values.ConsensusType.value.metadata.consenters} />
								) : (
									<Card className="p-4">
										<p className="text-sm text-muted-foreground text-center">No consenters configured</p>
									</Card>
								)}
							</div>
						}
						chaincode={
							<div className="space-y-4">
								<div className="flex items-center gap-4 mb-6">
									<div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center">
										<Code className="h-6 w-6 text-primary" />
									</div>
									<div>
										<h2 className="text-lg font-semibold">Chaincode Installation</h2>
										<p className="text-sm text-muted-foreground">Instructions for installing and managing chaincode</p>
									</div>
								</div>

								<Card className="p-6">
									<div className="prose dark:prose-invert max-w-none prose-pre:bg-muted prose-pre:border prose-pre:border-border prose-pre:rounded-lg prose-pre:p-4 prose-code:text-primary prose-code:bg-muted prose-code:px-1 prose-code:py-0.5 prose-code:rounded prose-code:before:content-none prose-code:after:content-none">
										<ReactMarkdown
											rehypePlugins={[rehypeRaw]}
											components={{
												h1: ({ children }) => <h1 className="text-2xl font-bold mb-4 mt-0">{children}</h1>,
												h2: ({ children }) => <h2 className="text-xl font-semibold mt-6 mb-3">{children}</h2>,
												code: ({ node, className, children, ...props }) => {
													const match = /language-(\w+)/.exec(className || '')
													const content = Array.isArray(children) ? children.join('') : String(children)

													return match ? (
														<div className="relative group">
															{/* <div className="absolute left-2 top-2 text-xs text-muted-foreground bg-background/80 px-2 py-1 rounded-md">{match[1]}</div> */}
															<CopyButton text={content.replace(/\n$/, '')} />
															<SyntaxHighlighter style={docco} language="javascript">
																{content}
															</SyntaxHighlighter>
														</div>
													) : (
														<code {...props} className={`${className} !bg-muted !text-primary px-1.5 py-0.5 rounded`}>
															{children}
														</code>
													)
												},
												p: ({ children }) => <p className="my-4 leading-7">{children}</p>,
												ul: ({ children }) => <ul className="my-6 ml-6 list-disc [&>li]:mt-2">{children}</ul>,
												ol: ({ children }) => <ol className="my-6 ml-6 list-decimal [&>li]:mt-2">{children}</ol>,
												blockquote: ({ children }) => <blockquote className="mt-6 border-l-2 border-border pl-6 italic">{children}</blockquote>,
											}}
										>
											{getChainCodeInstructions(network.name!, fabricOrgs?.[0]?.mspId || 'Org1MSP')}
										</ReactMarkdown>
									</div>
								</Card>
							</div>
						}
						channelUpdate={
							<div className="space-y-6">
								<p>Coming Soon</p>
							</div>
						}
						proposals={
							<div className="space-y-6">
								<p>Pro Only</p>
							</div>
						}
						share={<p>Pro Only</p>}
					/>
				</Card>
			</div>
		</div>
	)
}
