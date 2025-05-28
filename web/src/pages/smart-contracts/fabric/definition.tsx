import { getNetworksFabricByIdNodesOptions } from '@/api/client/@tanstack/react-query.gen'
import { getNetworksFabricByIdNodes } from '@/api/client/sdk.gen'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { zodResolver } from '@hookform/resolvers/zod'
import { useQuery } from '@tanstack/react-query'
import { Plus } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { useLocation, useNavigate } from 'react-router-dom'
import * as z from 'zod'

const versionFormSchema = z.object({
	endorsementPolicy: z.string().min(1, 'Endorsement policy is required'),
	dockerImage: z.string().min(1, 'Docker image is required'),
	version: z.string().min(1, 'Version is required'),
	sequence: z.number().min(1, 'Sequence must be at least 1'),
})

type VersionFormValues = z.infer<typeof versionFormSchema>

const LIFECYCLE_ACTIONS = ['install', 'approve', 'deploy', 'commit'] as const
type LifecycleAction = (typeof LIFECYCLE_ACTIONS)[number]
const actionLabels: Record<LifecycleAction, string> = {
	install: 'Install',
	approve: 'Approve',
	deploy: 'Deploy',
	commit: 'Commit',
}
const actionColors: Record<LifecycleAction, string> = {
	install: 'bg-blue-100 text-blue-800',
	approve: 'bg-yellow-100 text-yellow-800',
	deploy: 'bg-purple-100 text-purple-800',
	commit: 'bg-green-100 text-green-800',
}

export default function FabricChaincodeDefinitionDetail() {
	const location = useLocation()
	const navigate = useNavigate()
	const def = location.state?.definition
	const [versions, setVersions] = useState(
		(def?.versions || []).map((v: any) => ({
			...v,
			actions: {} as Record<LifecycleAction, boolean>,
		}))
	)
	const [isAddDialogOpen, setIsAddDialogOpen] = useState(false)
	const [editIdx, setEditIdx] = useState<number | null>(null)
	const [installDialogOpen, setInstallDialogOpen] = useState(false)
	const [selectedVersionIdx, setSelectedVersionIdx] = useState<number | null>(null)
	const [selectedPeers, setSelectedPeers] = useState<Set<string>>(new Set())
	const networkId = 11
	// Fetch network peers
	const { data: networkNodesResponse } = useQuery({
		...getNetworksFabricByIdNodesOptions({
			path: { id: networkId },
		}),
	})

	const availablePeers = networkNodesResponse?.nodes?.filter((node) => node.node?.nodeType === 'FABRIC_PEER' && node.status === 'joined') || []

	const form = useForm<VersionFormValues>({
		resolver: zodResolver(versionFormSchema),
		defaultValues: {
			endorsementPolicy: '',
			dockerImage: '',
			version: '1.0',
			sequence: 1,
		},
	})
	const editForm = useForm<VersionFormValues>({
		resolver: zodResolver(versionFormSchema),
		defaultValues: {
			endorsementPolicy: '',
			dockerImage: '',
			version: '1.0',
			sequence: 1,
		},
	})

	const onSubmit = (data: VersionFormValues) => {
		setVersions((prev) => {
			const newVersions = [
				...prev,
				{
					version: data.version,
					sequence: data.sequence,
					endorsementPolicy: data.endorsementPolicy,
					dockerImage: data.dockerImage,
					actions: {} as Record<LifecycleAction, boolean>,
				},
			]
			form.reset({
				endorsementPolicy: '',
				dockerImage: '',
				version: '1.0',
				sequence: newVersions.length + 1,
			})
			return newVersions
		})
		setIsAddDialogOpen(false)
	}

	const handleAction = (idx: number, action: LifecycleAction) => {
		if (action === 'install') {
			setSelectedVersionIdx(idx)
			setInstallDialogOpen(true)
			return
		}
		setVersions((prev) =>
			prev.map((v, i) => {
				if (i !== idx) return v
				return {
					...v,
					actions: {
						...v.actions,
						[action]: true,
					},
				}
			})
		)
	}

	const handleInstall = () => {
		if (selectedVersionIdx === null) return
		setVersions((prev) =>
			prev.map((v, i) => {
				if (i !== selectedVersionIdx) return v
				return {
					...v,
					actions: {
						...v.actions,
						install: true,
					},
				}
			})
		)
		setInstallDialogOpen(false)
		setSelectedPeers(new Set())
		setSelectedVersionIdx(null)
	}

	const handleEdit = (idx: number) => {
		setEditIdx(idx)
		editForm.reset({
			endorsementPolicy: versions[idx].endorsementPolicy,
			dockerImage: versions[idx].dockerImage,
			version: versions[idx].version,
			sequence: versions[idx].sequence,
		})
	}

	const onEditSubmit = (data: VersionFormValues) => {
		if (editIdx === null) return
		setVersions((prev) => prev.map((v, i) => (i === editIdx ? { ...v, ...data } : v)))
		setEditIdx(null)
	}

	if (!def) {
		return (
			<Card className="p-6">
				No chaincode definition found.{' '}
				<Button variant="link" onClick={() => navigate(-1)}>
					Back
				</Button>
			</Card>
		)
	}

	return (
		<div className="flex-1 p-8 w-full">
			<Button variant="link" onClick={() => navigate(-1)} className="mb-4">
				Back
			</Button>
			<Card className="p-6 mb-6">
				<div className="font-semibold text-lg mb-2">{def.name}</div>
				<div className="text-sm text-muted-foreground mb-1">Network: {def.networkName}</div>
			</Card>
			<div className="flex items-center justify-between mb-4">
				<div className="font-semibold">Chaincode Definitions</div>
				<Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
					<DialogTrigger asChild>
						<Button size="sm" variant="secondary">
							<Plus className="w-4 h-4 mr-2" />
							Add Definition
						</Button>
					</DialogTrigger>
					<DialogContent>
						<DialogHeader>
							<DialogTitle>Add Chaincode Definition</DialogTitle>
							<DialogDescription>Create a new chaincode definition with version and sequence.</DialogDescription>
						</DialogHeader>
						<Form {...form}>
							<form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
								<div className="grid grid-cols-2 gap-4">
									<FormField
										control={form.control}
										name="version"
										render={({ field }) => (
											<FormItem>
												<FormLabel>Version</FormLabel>
												<FormControl>
													<Input {...field} />
												</FormControl>
												<FormMessage />
											</FormItem>
										)}
									/>
									<FormField
										control={form.control}
										name="sequence"
										render={({ field }) => (
											<FormItem>
												<FormLabel>Sequence</FormLabel>
												<FormControl>
													<Input type="number" {...field} onChange={(e) => field.onChange(parseInt(e.target.value))} />
												</FormControl>
												<FormMessage />
											</FormItem>
										)}
									/>
								</div>
								<FormField
									control={form.control}
									name="endorsementPolicy"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Endorsement Policy</FormLabel>
											<FormControl>
												<Textarea {...field} />
											</FormControl>
											<FormMessage />
										</FormItem>
									)}
								/>
								<FormField
									control={form.control}
									name="dockerImage"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Docker Image</FormLabel>
											<FormControl>
												<Input {...field} />
											</FormControl>
											<FormMessage />
										</FormItem>
									)}
								/>
								<DialogFooter>
									<Button type="submit">Add Definition</Button>
								</DialogFooter>
							</form>
						</Form>
					</DialogContent>
				</Dialog>
			</div>
			{versions.length === 0 ? (
				<Card className="p-6 text-center text-muted-foreground">No chaincode definitions yet.</Card>
			) : (
				versions.map((v, idx) => (
					<Card key={idx} className="p-4 mb-4">
						<div className="flex items-center gap-4 mb-2">
							<Badge variant="outline">Version {v.version}</Badge>
							<Badge variant="outline">Sequence {v.sequence}</Badge>
							{LIFECYCLE_ACTIONS.map(
								(action) =>
									v.actions[action] && (
										<span key={action} className={`px-2 py-1 rounded text-xs font-semibold ${actionColors[action]}`}>
											{actionLabels[action]}
										</span>
									)
							)}
						</div>
						<div className="mb-1 text-sm">
							<span className="font-medium">Endorsement Policy:</span> {v.endorsementPolicy}
						</div>
						<div className="mb-1 text-sm">
							<span className="font-medium">Docker Image:</span> {v.dockerImage}
						</div>
						<div className="mt-2 flex gap-2">
							<Button size="sm" variant="outline" onClick={() => handleEdit(idx)}>
								Edit
							</Button>
							{LIFECYCLE_ACTIONS.map((action) => (
								<Button key={action} size="sm" variant={v.actions[action] ? 'outline' : 'default'} onClick={() => handleAction(idx, action)}>
									{actionLabels[action]}
								</Button>
							))}
						</div>
						<Dialog open={editIdx === idx} onOpenChange={(open) => !open && setEditIdx(null)}>
							<DialogContent>
								<DialogHeader>
									<DialogTitle>Edit Chaincode Definition</DialogTitle>
								</DialogHeader>
								<Form {...editForm}>
									<form onSubmit={editForm.handleSubmit(onEditSubmit)} className="space-y-4">
										<div className="grid grid-cols-2 gap-4">
											<FormField
												control={editForm.control}
												name="version"
												render={({ field }) => (
													<FormItem>
														<FormLabel>Version</FormLabel>
														<FormControl>
															<Input {...field} />
														</FormControl>
														<FormMessage />
													</FormItem>
												)}
											/>
											<FormField
												control={editForm.control}
												name="sequence"
												render={({ field }) => (
													<FormItem>
														<FormLabel>Sequence</FormLabel>
														<FormControl>
															<Input type="number" {...field} onChange={(e) => field.onChange(parseInt(e.target.value))} />
														</FormControl>
														<FormMessage />
													</FormItem>
												)}
											/>
										</div>
										<FormField
											control={editForm.control}
											name="endorsementPolicy"
											render={({ field }) => (
												<FormItem>
													<FormLabel>Endorsement Policy</FormLabel>
													<FormControl>
														<Textarea {...field} />
													</FormControl>
													<FormMessage />
												</FormItem>
											)}
										/>
										<FormField
											control={editForm.control}
											name="dockerImage"
											render={({ field }) => (
												<FormItem>
													<FormLabel>Docker Image</FormLabel>
													<FormControl>
														<Input {...field} />
													</FormControl>
													<FormMessage />
												</FormItem>
											)}
										/>
										<DialogFooter>
											<Button type="submit">Save</Button>
										</DialogFooter>
									</form>
								</Form>
							</DialogContent>
						</Dialog>
					</Card>
				))
			)}
			<Dialog open={installDialogOpen} onOpenChange={setInstallDialogOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Install Chaincode</DialogTitle>
						<DialogDescription>Select the peers where you want to install the chaincode.</DialogDescription>
					</DialogHeader>
					<div className="space-y-4">
						{availablePeers.map((peer) => (
							<div key={peer.id} className="flex items-center space-x-2">
								<Checkbox
									id={`peer-${peer.id}`}
									checked={selectedPeers.has(peer.id!.toString())}
									onCheckedChange={(checked) => {
										setSelectedPeers((prev) => {
											const next = new Set(prev)
											if (checked) {
												next.add(peer.id!.toString())
											} else {
												next.delete(peer.id!.toString())
											}
											return next
										})
									}}
								/>
								<label htmlFor={`peer-${peer.id}`} className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
									{peer.node?.name} ({peer.node?.fabricPeer?.mspId})
								</label>
							</div>
						))}
					</div>
					<DialogFooter>
						<Button onClick={handleInstall} disabled={selectedPeers.size === 0}>
							Install
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	)
}
