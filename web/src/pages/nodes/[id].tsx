import { HttpNodeResponse, ResponseErrorResponse } from '@/api/client'
import {
	deleteNodesByIdMutation,
	getNodesByIdEventsOptions,
	getNodesByIdOptions,
	postNodesByIdRestartMutation,
	postNodesByIdStartMutation,
	postNodesByIdStopMutation,
} from '@/api/client/@tanstack/react-query.gen'
import { BesuNodeConfig } from '@/components/nodes/BesuNodeConfig'
import { FabricOrdererConfig } from '@/components/nodes/FabricOrdererConfig'
import { FabricPeerConfig } from '@/components/nodes/FabricPeerConfig'
import { FabricNodeChannels } from '@/components/nodes/FabricNodeChannels'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { CertificateViewer } from '@/components/ui/certificate-viewer'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import { TimeAgo } from '@/components/ui/time-ago'
import { cn } from '@/lib/utils'
import { useMutation, useQuery } from '@tanstack/react-query'
import { format } from 'date-fns/format'
import { AlertCircle, CheckCircle2, Clock, Play, PlayCircle, RefreshCcw, RefreshCw, Square, StopCircle, XCircle } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { toast } from 'sonner'

interface DeploymentConfig {
	type?: string
	mode?: string
	organizationId?: number
	mspId?: string
	signKeyId?: number
	tlsKeyId?: number
	signCert?: string
	tlsCert?: string
	caCert?: string
	tlsCaCert?: string
	listenAddress?: string
	chaincodeAddress?: string
	eventsAddress?: string
	operationsListenAddress?: string
	externalEndpoint?: string
	adminAddress?: string
	domainNames?: string[]
}

function isFabricNode(node: HttpNodeResponse): node is HttpNodeResponse & { deploymentConfig: DeploymentConfig } {
	return node.platform === 'FABRIC' && (node.fabricPeer !== undefined || node.fabricOrderer !== undefined)
}

function isBesuNode(node: HttpNodeResponse): node is HttpNodeResponse {
	return node.platform === 'BESU' && node.besuNode !== undefined
}

function getNodeActions(status: string) {
	switch (status.toLowerCase()) {
		case 'running':
			return [
				{ label: 'Stop', action: 'stop', icon: Square },
				{ label: 'Restart', action: 'restart', icon: RefreshCw },
			]
		case 'stopped':
			return [{ label: 'Start', action: 'start', icon: Play }]
		case 'error':
			return [
				{ label: 'Start', action: 'start', icon: Play },
				{ label: 'Restart', action: 'restart', icon: RefreshCw },
			]
		case 'starting':
		case 'stopping':
			return [] // No actions while transitioning
		default:
			return [
				{ label: 'Start', action: 'start', icon: Play },
				{ label: 'Stop', action: 'stop', icon: Square },
				{ label: 'Restart', action: 'restart', icon: RefreshCw },
			]
	}
}

function getEventIcon(type: string) {
	switch (type.toUpperCase()) {
		case 'START':
			return PlayCircle
		case 'STOP':
			return StopCircle
		case 'RESTART':
			return RefreshCcw
		default:
			return Clock
	}
}

function getEventStatusIcon(status: string) {
	switch (status.toUpperCase()) {
		case 'SUCCESS':
		case 'COMPLETED':
			return CheckCircle2
		case 'FAILED':
			return XCircle
		case 'PENDING':
			return Clock
		default:
			return AlertCircle
	}
}

function getEventStatusColor(status: string) {
	switch (status.toUpperCase()) {
		case 'SUCCESS':
		case 'COMPLETED':
			return 'text-green-500'
		case 'FAILED':
			return 'text-red-500'
		case 'PENDING':
			return 'text-yellow-500'
		default:
			return 'text-gray-500'
	}
}

export default function NodeDetailPage() {
	const { id } = useParams<{ id: string }>()
	const navigate = useNavigate()
	const [searchParams, setSearchParams] = useSearchParams()
	const [logs, setLogs] = useState<string>('')
	const logsRef = useRef<HTMLTextAreaElement>(null)
	const abortControllerRef = useRef<AbortController | null>(null)

	// Get the active tab from URL or default to 'logs'
	const activeTab = searchParams.get('tab') || 'logs'

	// Update URL when tab changes
	const handleTabChange = (value: string) => {
		searchParams.set('tab', value)
		setSearchParams(searchParams)
	}

	const {
		data: node,
		isLoading,
		refetch,
		error,
	} = useQuery({
		...getNodesByIdOptions({
			path: { id: parseInt(id!) },
		}),
	})

	const startNode = useMutation({
		...postNodesByIdStartMutation(),
		onSuccess: () => {
			toast.success('Node started successfully')
			refetch()
		},
		onError: (error: any) => {
			toast.error(`Failed to start node: ${error.message}`)
		},
	})

	const stopNode = useMutation({
		...postNodesByIdStopMutation(),
		onSuccess: () => {
			toast.success('Node stopped successfully')
			refetch()
		},
		onError: (error: any) => {
			toast.error(`Failed to stop node: ${error.message}`)
		},
	})

	const restartNode = useMutation({
		...postNodesByIdRestartMutation(),
		onSuccess: () => {
			toast.success('Node restarted successfully')
			refetch()
		},
		onError: (error: any) => {
			toast.error(`Failed to restart node: ${error.message}`)
		},
	})

	const deleteNode = useMutation({
		...deleteNodesByIdMutation(),
		onSuccess: () => {
			toast.success('Node deleted successfully')
			navigate('/nodes')
		},
		onError: (error: any) => {
			toast.error(`Failed to delete node: ${error.message}`)
		},
	})

	const { data: events, refetch: refetchEvents } = useQuery({
		...getNodesByIdEventsOptions({
			path: { id: parseInt(id!) },
		}),
	})

	const handleAction = async (action: string) => {
		if (!node) return

		try {
			switch (action) {
				case 'start':
					await startNode.mutateAsync({ path: { id: node.id! } })
					refetchEvents()
					break
				case 'stop':
					await stopNode.mutateAsync({ path: { id: node.id! } })
					refetchEvents()
					break
				case 'restart':
					await restartNode.mutateAsync({ path: { id: node.id! } })
					refetchEvents()
					break
				case 'delete':
					await deleteNode.mutateAsync({ path: { id: node.id! } })
					break
			}
		} catch (error) {
			// Error handling is done in the mutation callbacks
		}
	}

	useEffect(() => {
		const fetchLogs = async () => {
			try {
				// Cancel previous request if exists
				if (abortControllerRef.current) {
					abortControllerRef.current.abort()
				}

				// Create new abort controller for this request
				const abortController = new AbortController()
				abortControllerRef.current = abortController

				const response = await fetch(`/api/v1/nodes/${id}/logs`, {
					signal: abortController.signal,
					credentials: 'include',
				})

				if (!response.body) {
					throw new Error('No response body')
				}

				const reader = response.body.getReader()
				const decoder = new TextDecoder()
				let fullText = ''

				while (true) {
					const { value, done } = await reader.read()
					if (done) break

					const text = decoder.decode(value)
					fullText += text
				}

				// Set full text at once and scroll to bottom
				setLogs(fullText)
				if (logsRef.current) {
					setTimeout(() => {
						if (logsRef.current) {
							logsRef.current.scrollTop = logsRef.current.scrollHeight
						}
					}, 100)
				}
			} catch (error) {
				if (error instanceof Error && error.name === 'AbortError') {
					// Ignore abort errors
					return
				}
				console.error('Error fetching logs:', error)
			}
		}

		if (id) {
			fetchLogs()
		}

		return () => {
			// Cleanup: abort any ongoing request when component unmounts
			if (abortControllerRef.current) {
				abortControllerRef.current.abort()
			}
		}
	}, [id])

	if (isLoading) {
		return <div>Loading...</div>
	}

	if (error) {
		return <div>Error loading node: {(error as any).error.message}</div>
	}
	if (!node) {
		return <div>Node not found</div>
	}
	return (
		<div className="flex-1 space-y-6 p-8">
			<div className="flex items-center justify-between">
				<div>
					<h1 className="text-2xl font-semibold">{node.name}</h1>
					<p className="text-muted-foreground">Node Details</p>
				</div>
				<div className="flex gap-2">
					{getNodeActions(node.status!).map(({ label, action, icon: Icon }) => (
						<Button
							key={action}
							onClick={() => handleAction(action)}
							variant="outline"
							size="sm"
							disabled={['starting', 'stopping'].includes(node.status!.toLowerCase()) || startNode.isPending || stopNode.isPending || restartNode.isPending}
						>
							<Icon className="mr-2 h-4 w-4" />
							{label}
						</Button>
					))}
				</div>
			</div>

			<div className="grid gap-6 md:grid-cols-2">
				<Card>
					<CardHeader>
						<CardTitle>General Information</CardTitle>
						<CardDescription>Basic node details and configuration</CardDescription>
					</CardHeader>
					<CardContent className="space-y-4">
						<div className="grid grid-cols-2 gap-4">
							<div>
								<p className="text-sm font-medium text-muted-foreground">Status</p>
								<Badge variant="default">{node.status}</Badge>
							</div>
							<div>
								<p className="text-sm font-medium text-muted-foreground">Platform</p>
								<p>{isFabricNode(node) ? 'Fabric' : 'Besu'}</p>
							</div>
							<div>
								<p className="text-sm font-medium text-muted-foreground">Created At</p>
								<TimeAgo date={node.createdAt!} />
							</div>
							{node.updatedAt && (
								<div>
									<p className="text-sm font-medium text-muted-foreground">Updated At</p>
									<TimeAgo date={node.updatedAt} />
								</div>
							)}
						</div>
					</CardContent>
				</Card>

				<>
					{node.fabricPeer && <FabricPeerConfig config={node.fabricPeer} />}
					{node.fabricOrderer && <FabricOrdererConfig config={node.fabricOrderer} />}
					{node.besuNode && <BesuNodeConfig config={node.besuNode} />}
				</>
			</div>

			<Tabs defaultValue={activeTab} className="space-y-4" onValueChange={handleTabChange}>
				<TabsList>
					<TabsTrigger value="logs">Logs</TabsTrigger>
					<TabsTrigger value="crypto">Crypto Material</TabsTrigger>
					<TabsTrigger value="events">Events</TabsTrigger>
					{isFabricNode(node) && <TabsTrigger value="channels">Channels</TabsTrigger>}
				</TabsList>

				<TabsContent value="logs" className="space-y-4">
					<Card>
						<CardHeader>
							<CardTitle>Logs</CardTitle>
							<CardDescription>Real-time node logs</CardDescription>
						</CardHeader>
						<CardContent>
							<Textarea
								ref={logsRef}
								value={logs}
								readOnly
								className="font-mono text-sm h-[400px] bg-muted"
								style={{
									whiteSpace: 'pre',
									overflowWrap: 'normal',
									overflowX: 'auto',
								}}
							/>
						</CardContent>
					</Card>
				</TabsContent>

				<TabsContent value="crypto" className="space-y-4">
					<Card>
						<CardHeader>
							<CardTitle>Certificates</CardTitle>
							<CardDescription>Node certificates and keys</CardDescription>
						</CardHeader>
						<CardContent className="space-y-6">
							{node.fabricPeer && (
								<>
									<div className="space-y-4">
										<CertificateViewer label="Signing Certificate" certificate={node.fabricPeer?.signCert || ''} />
										<CertificateViewer label="TLS Certificate" certificate={node.fabricPeer?.tlsCert || ''} />
										<CertificateViewer label="CA Certificate" certificate={node.fabricPeer?.signCaCert || ''} />
										<CertificateViewer label="TLS CA Certificate" certificate={node.fabricPeer?.tlsCaCert || ''} />
									</div>
								</>
							)}
							{node.fabricOrderer && (
								<>
									<div className="space-y-4">
										<CertificateViewer label="Signing Certificate" certificate={node.fabricOrderer?.signCert || ''} />
										<CertificateViewer label="TLS Certificate" certificate={node.fabricOrderer?.tlsCert || ''} />
										<CertificateViewer label="CA Certificate" certificate={node.fabricOrderer?.signCaCert || ''} />
										<CertificateViewer label="TLS CA Certificate" certificate={node.fabricOrderer?.tlsCaCert || ''} />
									</div>
								</>
							)}
						</CardContent>
					</Card>
				</TabsContent>

				<TabsContent value="events">
					<Card>
						<CardHeader>
							<CardTitle>Event History</CardTitle>
							<CardDescription>Recent node operations and their status</CardDescription>
						</CardHeader>
						<CardContent>
							<div className="space-y-8">
								{events?.items?.map((event) => {
									const EventIcon = getEventIcon(event.type!)
									const StatusIcon = getEventStatusIcon(event.type!)
									return (
										<div key={event.id} className="flex gap-4">
											<div className="mt-1">
												<EventIcon className="h-5 w-5 text-muted-foreground" />
											</div>
											<div className="flex-1 space-y-1">
												<div className="flex items-center justify-between">
													<div className="flex items-center gap-2">
														<span className="font-medium">{event.type}</span>
														<StatusIcon className={cn('h-4 w-4', getEventStatusColor(event.type!))} />
														<span className="text-sm text-muted-foreground">{event.type}</span>
													</div>
													<time className="text-sm text-muted-foreground">{format(new Date(event.created_at!), 'PPp')}</time>
												</div>
												{event.data && typeof event.data === 'object' ? (
													<div className="rounded-md bg-muted p-2 text-sm">
														<pre className="whitespace-pre-wrap font-mono text-xs">{JSON.stringify(event.data, null, 2)}</pre>
													</div>
												) : null}
											</div>
										</div>
									)
								})}
							</div>
						</CardContent>
					</Card>
				</TabsContent>

				<TabsContent value="channels">
					{isFabricNode(node) && <FabricNodeChannels nodeId={node.id!} />}
				</TabsContent>
			</Tabs>
		</div>
	)
}
