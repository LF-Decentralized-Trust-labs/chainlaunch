import { getNodesByIdOptions, getNodesDefaultsFabricPeerOptions, getOrganizationsOptions, putNodesByIdMutation } from '@/api/client/@tanstack/react-query.gen'
import { TypesAddressOverride } from '@/api/client/types.gen'
import { FabricNodeForm } from '@/components/nodes/fabric-node-form'
import { useQuery, useMutation } from '@tanstack/react-query'
import { useMemo } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { toast } from 'sonner'

type FabricNodeFormValues = {
	name: string
	fabricProperties: {
		nodeType: 'FABRIC_PEER' | 'FABRIC_ORDERER'
		mode: 'docker' | 'service'
		version: string
		organizationId?: number
		listenAddress: string
		operationsListenAddress: string
		externalEndpoint: string
		domains?: string[]
		chaincodeAddress?: string
		eventsAddress?: string
		adminAddress?: string
		addressOverrides?: TypesAddressOverride[]
	}
}

export default function EditFabricNodePage() {
	const navigate = useNavigate()
	const { id } = useParams<{ id: string }>()

	const { data: node } = useQuery({
		...getNodesByIdOptions({
			path: { id: parseInt(id!) },
		}),
		enabled: !!id,
	})

	const { data: organizations } = useQuery({
		...getOrganizationsOptions(),
	})

	const { data: peerDefaults } = useQuery({
		...getNodesDefaultsFabricPeerOptions(),
	})

	const updateNode = useMutation({
		...putNodesByIdMutation(),
		onSuccess: () => {
			toast.success('Node updated successfully')
			navigate(`/nodes/${id}`)
		},
		onError: (error: any) => {
			if (error instanceof Error) {
				toast.error(`Failed to update node: ${error.message}`)
			} else if (error.message) {
				toast.error(`Failed to update node: ${error.message}`)
			} else {
				toast.error('An unknown error occurred')
			}
		},
	})

	const onSubmit = (data: FabricNodeFormValues) => {
		const organization = organizations?.find((org) => org.id === data.fabricProperties.organizationId)

		const fabricPeer = {
			nodeType: 'FABRIC_PEER' as const,
			mode: data.fabricProperties.mode,
			organizationId: data.fabricProperties.organizationId,
			listenAddress: data.fabricProperties.listenAddress,
			operationsListenAddress: data.fabricProperties.operationsListenAddress,
			externalEndpoint: data.fabricProperties.externalEndpoint,
			domainNames: data.fabricProperties.domains || [],
			name: data.name,
			chaincodeAddress: data.fabricProperties.chaincodeAddress || '',
			eventsAddress: data.fabricProperties.eventsAddress || '',
			mspId: organization?.mspId || '',
			version: data.fabricProperties.version || '2.5.12',
			addressOverrides: data.fabricProperties.addressOverrides || [],
		}

		updateNode.mutate({
			path: { id: parseInt(id!) },
			body: {
				name: data.name,
				blockchainPlatform: 'FABRIC',
				fabricPeer,
			},
		})
	}

	// Convert node data to form values
	const defaultValues = useMemo(() => {
		return node?.fabricPeer
			? ({
					name: node.name || '',
					fabricProperties: {
						nodeType: 'FABRIC_PEER',
						mode: node.fabricPeer?.mode || 'service',
						version: node.fabricPeer?.version || '3.0.0',
						organizationId: node.fabricPeer?.organizationId,
						listenAddress: node.fabricPeer?.listenAddress || '',
						operationsListenAddress: node.fabricPeer?.operationsAddress || '',
						externalEndpoint: node.fabricPeer?.externalEndpoint || '',
						domains: node.fabricPeer?.domainNames || [],
						chaincodeAddress: node.fabricPeer?.chaincodeAddress || '',
						eventsAddress: node.fabricPeer?.eventsAddress || '',
						addressOverrides: node.fabricPeer?.addressOverrides || [],
					},
			  } as FabricNodeFormValues)
			: ({} as FabricNodeFormValues)
	}, [node])
	if (!node) {
		return (
			<div className="flex-1 p-8">
				<div className="max-w-2xl mx-auto">
					<div className="mb-8">
						<h1 className="text-2xl font-semibold">Loading...</h1>
					</div>
				</div>
			</div>
		)
	}

	if (node.nodeType !== 'FABRIC_PEER') {
		return (
			<div className="flex-1 p-8">
				<div className="max-w-2xl mx-auto">
					<div className="mb-8">
						<h1 className="text-2xl font-semibold">Invalid Node Type</h1>
						<p className="text-muted-foreground">This page only supports editing Fabric peer nodes.</p>
					</div>
				</div>
			</div>
		)
	}

	return (
		<div className="flex-1 p-8">
			<div className="max-w-2xl mx-auto">
				<div className="mb-8">
					<h1 className="text-2xl font-semibold">Edit Fabric Node</h1>
					<p className="text-muted-foreground">Update your Fabric node configuration</p>
				</div>

				<FabricNodeForm
					onSubmit={onSubmit}
					isSubmitting={updateNode.isPending}
					organizations={organizations?.map((org) => ({ id: org.id!, name: org.mspId! })) || []}
					defaults={peerDefaults}
					hideNodeType={true}
					submitText="Update Node"
					defaultValues={defaultValues}
				/>
			</div>
		</div>
	)
}
